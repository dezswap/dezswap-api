package api

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	gin_cache "github.com/chenyahui/gin-cache"
	"gorm.io/gorm"

	geckoController "github.com/dezswap/dezswap-api/api/controller/coingecko"
	comarcapController "github.com/dezswap/dezswap-api/api/controller/coinmarketcap"
	dashboardController "github.com/dezswap/dezswap-api/api/controller/dashboard"
	nc "github.com/dezswap/dezswap-api/api/controller/notice"
	routerController "github.com/dezswap/dezswap-api/api/controller/router"

	"github.com/dezswap/dezswap-api/api/service/coingecko"
	"github.com/dezswap/dezswap-api/api/service/coinmarketcap"
	"github.com/dezswap/dezswap-api/api/service/dashboard"
	ns "github.com/dezswap/dezswap-api/api/service/notice"
	rs "github.com/dezswap/dezswap-api/api/service/router"

	"github.com/dezswap/dezswap-api/api/controller"
	"github.com/dezswap/dezswap-api/api/docs"
	"github.com/gin-contrib/cors"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/dezswap/dezswap-api/pkg/db/api"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/dezswap/dezswap-api/pkg/xpla"
	"github.com/evalphobia/logrus_sentry"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	// swagger embed files
)

type app struct {
	engine *gin.Engine
	config configs.ApiConfig
	logger logging.Logger
}

func RunServer(c configs.Config, cache cache.Cache, db *gorm.DB) {
	logger := logging.New(c.Api.Server.Name, c.Log)
	app := app{
		gin.Default(),
		c.Api,
		logger,
	}
	serverConfig := c.Api.Server
	gin.SetMode(serverConfig.Mode)
	app.setMiddlewares(cache)
	app.initApis(c.Api, db)
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
		app.engine.Use(gin_cache.Cache(cache, time.Second*time.Duration(xpla.NetworkMetadata.BlockSecond),
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

func (app *app) initApis(c configs.ApiConfig, db *gorm.DB) {
	chainId := c.Server.ChainId
	if chainId == "" {
		panic("chainId is empty")
	}
	pairService := service.NewPairService(chainId, db)
	poolService := service.NewPoolService(chainId, db)
	tokenService := service.NewTokenService(chainId, db)
	statService := service.NewStatService(chainId, db)

	version := c.Server.Version
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

	routerRepo := api.NewRouterDbRepo(chainId, db)
	routerService := rs.New(routerRepo)
	routerController.InitRouterController(routerService, router.Group("/routes"), app.logger)
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
