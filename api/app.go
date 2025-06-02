package api

import (
	"fmt"
	"github.com/dezswap/dezswap-api/api/docs"
	v1 "github.com/dezswap/dezswap-api/api/v1"
	"github.com/dezswap/dezswap-api/pkg"
	"net/http"
	"regexp"
	"time"

	gin_cache "github.com/chenyahui/gin-cache"
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

func RunServer(c configs.Config, cache cache.Cache, db *gorm.DB) {
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
	app.setMiddlewares(cache)

	v1Router := app.engine.Group(ApiVersion)
	v1.RegisterRoutes(v1Router, serverConfig.ChainId, AppVersion, app.NetworkMetadata, db, cache, app.logger)

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
		app.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	app.run()
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

func (app *app) setMiddlewares(cache cache.Cache) {
	app.engine.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		app.logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}))

	allowedOrigins := []string{`\.dezswap\.io$`, `dezswap\.netlify\.app$`, `^https?:\/\/localhost(:\d+)?$`}
	conf := cors.DefaultConfig()
	conf.AllowOriginFunc = func(origin string) bool {
		for _, o := range allowedOrigins {
			matched, _ := regexp.MatchString(o, origin)
			if matched {
				return true
			}
		}
		return false
	}
	conf.AllowMethods = []string{"GET", "OPTIONS"}
	app.engine.Use(cors.New(conf))
	if cache != nil {
		app.engine.Use(gin_cache.Cache(cache, time.Second*time.Duration(app.BlockSecond),
			gin_cache.WithCacheStrategyByRequest(func(c *gin.Context) (bool, gin_cache.Strategy) {
				return true, gin_cache.Strategy{
					CacheKey: c.Request.Host + c.Request.RequestURI,
				}
			}),
			gin_cache.WithDiscardHeaders(gin_cache.CorsHeaders()),
		))
	}
	app.engine.UseRawPath = true
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
