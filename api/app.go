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
		docs.SwaggerInfo.BasePath = fmt.Sprintf("/%s", ApiVersion)
		g := app.engine.Group("")
		g.Use(cacheHandlers.Timed())
		g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	if err := mcpserver.Mount(app.engine, c.Api.MCP, AppVersion); err != nil {
		panic(err)
	}

	app.run()
}

// cacheHandlers builds the cache middleware the router hands to its groups.
func (app *app) cacheHandlers(ctx context.Context, store cache.Cache, db *gorm.DB) httpcache.Handlers {
	blockTime := time.Second * time.Duration(app.BlockSecond)
	versioner := cachekey.NewVersioner(ctx, db, app.config.Server.ChainId, blockTime, func(err error) {
		app.logger.Warn(err)
	})

	return httpcache.New(store, versioner, blockTime)
}

func (app *app) run() {
	type NotFound struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	app.engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, NotFound{Code: http.StatusNotFound, Message: "Not Found"})
	})
	if err := app.engine.Run(fmt.Sprintf(":%s", app.config.Server.Port)); err != nil {
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
