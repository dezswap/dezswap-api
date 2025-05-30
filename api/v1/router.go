package v1

import (
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

// RegisterRoutes sets up v1 API endpoints
func RegisterRoutes(rg *gin.RouterGroup, chainId string, version string, networkMetadata pkg.NetworkMetadata, db *gorm.DB, cache cache.Cache, logger logging.Logger) {
	statusService := service.NewStatusService(db, cache)
	pairService := service.NewPairService(chainId, db)
	poolService := service.NewPoolService(chainId, db)
	tokenService := service.NewTokenService(chainId, db)
	statService := service.NewStatService(chainId, db)

	controller.InitStatusController(statusService, rg, version, logger)
	controller.InitPairController(pairService, rg, networkMetadata, logger)
	controller.InitPoolController(poolService, rg, networkMetadata, logger)
	controller.InitTokenController(tokenService, rg, logger)
	controller.InitStatController(statService, rg, logger)

	// CoinGecko endpoint
	r := rg.Group("/coingecko")
	coinGeckoPairService := cgs.NewPairService(chainId, db)
	coinGeckoTickerService := cgs.NewTickerService(chainId, db)

	coingecko.InitPairController(coinGeckoPairService, r, logger)
	coingecko.InitTickerController(coinGeckoTickerService, r, logger)

	// CoinMarketCap endpoint
	r = rg.Group("/coinmarketcap")
	coinMarketCapTickerService := cmcs.NewTickerService(chainId, db)
	coinmarketcap.InitTickerController(coinMarketCapTickerService, r, logger)

	dashboardService := ds.NewDashboardService(chainId, db)
	dashboard.InitDashboardController(dashboardService, rg.Group("/dashboard"), logger)

	noticeService := ns.NewService(db)
	notice.InitNoticeController(noticeService, rg.Group("/notices"), logger)

	routerRepo := api.NewRouterDbRepo(chainId, db)
	routerService := rs.New(routerRepo)
	router.InitRouterController(routerService, rg.Group("/routes"), logger)
}
