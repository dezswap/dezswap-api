package api

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/dezswap/dezswap-api/api/controller"
	"github.com/dezswap/dezswap-api/api/docs"
	"github.com/gin-contrib/cors"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/evalphobia/logrus_sentry"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// swagger embed files
)

type app struct {
	engine *gin.Engine
	config configs.ApiConfig
	logger logging.Logger
}

func RunServer(c configs.Config) *app {
	logger := logging.New(c.Api.Server.Name, c.Log)
	app := app{
		gin.Default(),
		c.Api,
		logger,
	}
	serverConfig := c.Api.Server
	gin.SetMode(serverConfig.Mode)
	app.setMiddlewares()
	app.initApis(c.Api)
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
		if c.Api.Server.Version != "" {
			docs.SwaggerInfo.BasePath = "/" + c.Api.Server.Version
		}
		app.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	app.run()
	return &app
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
}

func (app *app) initApis(c configs.ApiConfig) {
	dbConfig := c.DB
	dbDsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database,
	)

	db, err := gorm.Open(postgres.Open(dbDsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	chainId := c.Server.ChainId
	if chainId == "" {
		panic("chainId is empty")
	}
	pairService := service.NewPairService(chainId, db)
	poolService := service.NewPoolService(chainId, db)
	tokenService := service.NewTokenService(chainId, db)

	version := c.Server.Version
	router := app.engine.Group(version)
	controller.InitPairController(pairService, router, app.logger)
	controller.InitPoolController(poolService, router, app.logger)
	controller.InitTokenController(tokenService, router, app.logger)
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
