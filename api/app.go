package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	gin_cache "github.com/chenyahui/gin-cache"

	geckoController "github.com/dezswap/dezswap-api/api/controller/coingecko"
	comarcapController "github.com/dezswap/dezswap-api/api/controller/coinmarketcap"
	dashboardController "github.com/dezswap/dezswap-api/api/controller/dashboard"
	nc "github.com/dezswap/dezswap-api/api/controller/notice"
	"github.com/redis/go-redis/v9"

	"github.com/dezswap/dezswap-api/api/service/coingecko"
	"github.com/dezswap/dezswap-api/api/service/coinmarketcap"
	"github.com/dezswap/dezswap-api/api/service/dashboard"
	ns "github.com/dezswap/dezswap-api/api/service/notice"

	"github.com/dezswap/dezswap-api/api/controller"
	"github.com/dezswap/dezswap-api/api/docs"
	"github.com/gin-contrib/cors"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/dezswap/dezswap-api/pkg/cache/memory"
	cache_redis "github.com/dezswap/dezswap-api/pkg/cache/redis"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/evalphobia/logrus_sentry"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	cacheStore := app.cacheStore(c.Api.Cache)
	app.setMiddlewares(cacheStore)
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
		app.engine.Use(gin_cache.CacheByRequestURI(cache, time.Second*5))
	}
}

func (app *app) initApis(c configs.ApiConfig) {
	dbConfig := c.DB
	dbDsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database,
	)
	writer := io.MultiWriter(os.Stdout)
	db, err := gorm.Open(postgres.Open(dbDsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.New(
			log.New(writer, "\r\n", log.LstdFlags),
			logger.Config{
				IgnoreRecordNotFoundError: true,
				SlowThreshold:             time.Second,
				Colorful:                  false,
				LogLevel:                  logger.Warn,
			},
		),
	})
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
	statService := service.NewStatService(chainId, db)

	version := c.Server.Version
	app.engine.UseRawPath = true

	router := app.engine.Group(version)
	controller.InitPairController(pairService, router, app.logger)
	controller.InitPoolController(poolService, router, app.logger)
	controller.InitTokenController(tokenService, router, app.logger)
	controller.InitStatController(statService, router, app.logger)

	// CoinGecko endpoint
	r := router.Group("/coingecko")
	coinGeckoPairService := coingecko.NewPairService(chainId, db)
	coinGeckoTickerService := coingecko.NewTickerService(chainId, db)

	geckoController.InitPairController(coinGeckoPairService, r, app.logger)
	geckoController.InitTickerController(coinGeckoTickerService, r, app.logger)

	// CoinMarketCap endpoint
	r = router.Group("/coinmarketcap")
	coinMarketCapTickerService := coinmarketcap.NewTickerService(chainId, db)
	comarcapController.InitTickerController(coinMarketCapTickerService, r, app.logger)

	dashboardService := dashboard.NewDashboardService(chainId, db)
	dashboardController.InitDashboardController(dashboardService, router.Group("/dashboard"), app.logger)

	noticeService := ns.NewService(db)
	nc.InitNoticeController(noticeService, router.Group("/notices"), app.logger)
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

func (app *app) cacheStore(c configs.CacheConfig) cache.Cache {
	if c.RedisConfig.Host != "" {
		option := redis.Options{
			Addr:     fmt.Sprintf("%s:%s", c.RedisConfig.Host, c.RedisConfig.Port),
			Username: c.RedisConfig.User,
			Password: c.RedisConfig.Password,
			DB:       c.RedisConfig.DB,
			Protocol: c.RedisConfig.Protocol,
		}

		client := redis.NewClient(&option)
		if err := client.Ping(context.Background()).Err(); err != nil {
			panic(err)
		}
		return cache_redis.New(cache.NewByteCodec(), client)
	}
	if c.MemoryCache {
		return memory.NewMemoryCache(cache.NewByteCodec())
	}

	return nil
}
