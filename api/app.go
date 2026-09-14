package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/dezswap/dezswap-api/api/cachekey"
	"github.com/dezswap/dezswap-api/api/docs"
	"github.com/dezswap/dezswap-api/api/httpcache"
	"github.com/dezswap/dezswap-api/api/mcpserver"
	v1 "github.com/dezswap/dezswap-api/api/v1"
	"github.com/dezswap/dezswap-api/pkg"

	"gorm.io/gorm"

	"github.com/gin-contrib/cors"

	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/evalphobia/logrus_sentry"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	// swagger embed files
)

const ApiVersion = "v1"

// The server's own deadlines. Left to the default of none, a connection that never
// finishes sending its request, or never reads its response, holds its slot for as
// long as it cares to.
const (
	// readHeaderTimeout is what bounds a client dribbling headers a byte at a time.
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	// writeTimeout has to clear the slowest handler answering on a cold cache, which
	// is a dashboard aggregate over a month of windows.
	writeTimeout = 60 * time.Second
	idleTimeout  = 120 * time.Second
)

var AppVersion = "dev"

type app struct {
	engine *gin.Engine
	config configs.ApiConfig
	pkg.NetworkMetadata
	logger logging.Logger
}

func RunServer(ctx context.Context, c configs.Config, cache cache.Cache, db *gorm.DB) {
	serverConfig := c.Api.Server
	networkMetadata, err := pkg.GetNetworkMetadata(serverConfig.ChainId)
	if err != nil {
		panic(err)
	}

	logger := logging.New(c.Api.Server.Name, c.Log)
	app := app{
		gin.Default(),
		c.Api,
		networkMetadata,
		logger,
	}

	gin.SetMode(serverConfig.Mode)
	app.setMiddlewares()

	cacheHandlers := app.cacheHandlers(ctx, cache, db)

	v1Router := app.engine.Group(ApiVersion)
	v1.RegisterRoutes(v1Router, serverConfig.ChainId, serverConfig.CoinGeckoApiKey, AppVersion, app.NetworkMetadata, db, cache, cacheHandlers, app.logger)

	if c.Sentry.DSN != "" {
		if err := app.configureReporter(c.Sentry.DSN, serverConfig.ChainId, map[string]string{
			"x-app":      "dezswap-api",
			"x-env":      c.Log.Environment,
			"x-chain_id": c.Api.Server.ChainId,
		}); err != nil {
			panic(err)
		}
	}

	if c.Api.Server.Swagger {
		mountSwagger(app.engine)
	}

	if err := mcpserver.Mount(app.engine, c.Api.MCP, AppVersion); err != nil {
		panic(err)
	}

	app.run()
}

// mountSwagger serves the spec and the UI that reads it.
//
// Deliberately uncached. The handler picks which asset to answer with from the
// whole RequestURI, so /swagger/index.html?doc.json is the spec while
// /swagger/index.html is the page -- one path, two responses. Every key this
// service builds is the path plus what the route declares, and under one of those
// the first of the two answered would be replayed as the other.
func mountSwagger(engine *gin.Engine) {
	docs.SwaggerInfo.BasePath = fmt.Sprintf("/%s", ApiVersion)
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// cacheHandlers builds the cache middleware the router hands to its groups.
func (app *app) cacheHandlers(ctx context.Context, store cache.Cache, db *gorm.DB) httpcache.Handlers {
	blockTime := time.Second * time.Duration(app.BlockSecond)
	versioner := cachekey.NewVersioner(ctx, db, app.config.Server.ChainId, blockTime, func(err error) {
		app.logger.Warn(err)
	})

	return httpcache.New(app.config.Server.ChainId, store, versioner, blockTime)
}

func (app *app) run() {
	type NotFound struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	app.engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, NotFound{Code: http.StatusNotFound, Message: "Not Found"})
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", app.config.Server.Port),
		Handler:           app.engine,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (app *app) setMiddlewares() {
	app.engine.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		app.logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}))

	allowedOrigins := app.config.Server.CorsAllowedOrigins
	conf := cors.DefaultConfig()
	conf.AllowOriginFunc = func(origin string) bool {
		for _, o := range allowedOrigins {
			matched, _ := regexp.MatchString(o, origin)
			if matched {
				return true
			}
		}
		if app.config.MCP.Enabled {
			if slices.Contains(app.config.MCP.AllowedOrigins, origin) {
				return true
			}
		}
		return false
	}
	conf.AllowMethods = []string{"GET", "OPTIONS"}
	conf.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	if app.config.MCP.Enabled {
		// Allow MCP Streamable HTTP methods and protocol headers for browser preflight.
		// See https://modelcontextprotocol.io/specification/2025-11-25/basic/transports#streamable-http
		conf.AllowMethods = []string{"GET", "POST", "OPTIONS"}
		conf.AllowHeaders = append(conf.AllowHeaders, "Accept", "MCP-Protocol-Version", "Mcp-Session-Id", "Last-Event-ID")
		// Expose Mcp-Session-Id so browser MCP clients can continue the session.
		conf.ExposeHeaders = append(conf.ExposeHeaders, "Mcp-Session-Id")
	}
	app.engine.Use(cors.New(conf))
	app.engine.UseRawPath = true
	app.engine.UnescapePathValues = true
}

func (app *app) configureReporter(dsn, env string, tags map[string]string) error {
	hook, err := logrus_sentry.NewSentryHook(dsn, []logrus.Level{
		logrus.WarnLevel,
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	if err != nil {
		return err
	}
	hook.StacktraceConfiguration.Enable = true
	hook.SetTagsContext(tags)
	hook.SetEnvironment(env)
	logging.AddHookToLogger(app.logger, hook)
	return nil
}
