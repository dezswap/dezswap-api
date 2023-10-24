package dashboard

import (
	"github.com/dezswap/dezswap-api/api/controller"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
)

func InitDashboardController() controller.DashboardController {
	c := dashboardController{}
	c.logger.Debug("InitDashboardController")
	// TODO: remove when implement this is temporary code for lint
	c.mapper = mapper{}
	return &c
}

type dashboardController struct {
	logger logging.Logger
	mapper
}

// Dashboard godoc
//
//	@Summary		Recent 24H data with it's change rate
//	@Description	get Recent
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	RecentRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/recent [get]
func (c *dashboardController) Recent(ctx *gin.Context) {
}

// Dashboard godoc
//
//	@Summary		Volumes of user selected duration
//	@Description	get Volumes
//	@Tags			dashboard
//	@Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	VolumesRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/volumes [get]
func (c *dashboardController) Volumes(ctx *gin.Context) {
}

// Dashboard godoc
//
//	@Summary		TVLs of dezswap selected duration
//	@Description	get TVLs
//	@Tags			dashboard
//	@Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TVLsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/tvls [get]
func (c *dashboardController) TVLs(ctx *gin.Context) {
}

// Dashboard godoc
//
//	@Summary		Dezswap's statistics
//	@Description	get Statistic data of dezswap (address count, tx count, fee)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	StatisticRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/statistics [get]
func (c *dashboardController) Statistic(ctx *gin.Context) {
}

// Dashboard godoc
//
//	@Summary		Dezswap's Pools
//	@Description	get Pools data of dezswap (address, tvl, volume, fee, apr)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PoolsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/pools [get]
func (c *dashboardController) Pools(ctx *gin.Context) {
}

// Dashboard godoc
//
//	@Summary		Dezswap's Tokens
//	@Description	get Tokens data of dezswap (address, price, priceChange, volume, tvl)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TokensRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/tokens [get]
func (c *dashboardController) Tokens(ctx *gin.Context) {
}

// Dashboard godoc
//
//	@Summary		Dezswap's Transactions
//	@Description	get Transactions data of dezswap (action, totalValue, asset0amount, asset1amount, time)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TxReses
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/txs [get]
func (c *dashboardController) Txs(ctx *gin.Context) {
}
