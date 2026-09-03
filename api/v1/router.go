package v1

import (
	"github.com/dezswap/dezswap-api/api/cachekey"
	"github.com/dezswap/dezswap-api/api/httpcache"
	"github.com/dezswap/dezswap-api/api/v1/controller"
	"github.com/dezswap/dezswap-api/api/v1/controller/coingecko"
	"github.com/dezswap/dezswap-api/api/v1/controller/coinmarketcap"
	"github.com/dezswap/dezswap-api/api/v1/controller/dashboard"
	"github.com/dezswap/dezswap-api/api/v1/controller/notice"
	"github.com/dezswap/dezswap-api/api/v1/controller/router"
	"github.com/dezswap/dezswap-api/api/v1/service"
	cgs "github.com/dezswap/dezswap-api/api/v1/service/coingecko"
	cmcs "github.com/dezswap/dezswap-api/api/v1/service/coinmarketcap"
	ds "github.com/dezswap/dezswap-api/api/v1/service/dashboard"
	ns "github.com/dezswap/dezswap-api/api/v1/service/notice"
	rs "github.com/dezswap/dezswap-api/api/v1/service/router"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/dezswap/dezswap-api/pkg/db/api"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes sets up v1 API endpoints.
func RegisterRoutes(rg *gin.RouterGroup, chainId string, coinGeckoApiKey string, version string, networkMetadata pkg.NetworkMetadata, db *gorm.DB, cacheStore cache.Cache, cacheHandlers httpcache.Handlers, logger logging.Logger) {
	statusService := service.NewStatusService(db, cacheStore)
	pairService := service.NewPairService(chainId, db)
	poolService := service.NewPoolService(chainId, db)
	tokenService := service.NewTokenService(chainId, db)
	statService := service.NewStatService(chainId, db)

	controller.InitStatusController(statusService, rg, version, logger)

	// Tokens and pairs are the only families whose writes are tracked.
	versioned := func(r cachekey.Resource) *gin.RouterGroup {
		g := rg.Group("")
		g.Use(cacheHandlers.Versioned(r))

		return g
	}
	controller.InitTokenController(tokenService, versioned(cachekey.Tokens), logger)
	controller.InitPairController(pairService, versioned(cachekey.Pairs), networkMetadata, logger)

	// Nothing follows when these tables change, so the clock is all that expires them.
	timed := rg.Group("")
	timed.Use(cacheHandlers.Timed())

	controller.InitPoolController(poolService, timed, networkMetadata, logger)
	controller.InitStatController(statService, timed, logger)

	// CoinGecko endpoint
	r := timed.Group("/coingecko")
	coinGeckoPairService := cgs.NewPairService(chainId, db)
	coinGeckoTickerService := cgs.NewTickerService(chainId, db, coinGeckoApiKey)

	coingecko.InitPairController(coinGeckoPairService, r, logger)
	coingecko.InitTickerController(coinGeckoTickerService, r, logger)

	// CoinMarketCap endpoint
	r = timed.Group("/coinmarketcap")
	coinMarketCapTickerService := cmcs.NewTickerService(chainId, db)
	coinmarketcap.InitTickerController(coinMarketCapTickerService, r, logger)

	dashboardService := ds.NewDashboardService(chainId, db)
	dashboard.InitDashboardController(dashboardService, timed.Group("/dashboard"), logger)

	noticeService := ns.NewService(db)
	notice.InitNoticeController(noticeService, timed.Group("/notices"), logger)

	// Routes costs one indexed lookup into an already aggregated table, while its
	// from/to are caller-supplied and an unknown address still returns 200. Caching
	// would let any caller mint an entry per string it invents, and save little.
	routerRepo := api.NewRouterDbRepo(chainId, db)
	routerService := rs.New(routerRepo)
	router.InitRouterController(routerService, rg.Group("/routes"), logger)
}
