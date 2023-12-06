package dashboard

import (
	"net/http"

	"github.com/dezswap/dezswap-api/api/controller"
	dashboardService "github.com/dezswap/dezswap-api/api/service/dashboard"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func InitDashboardController(s dashboardService.Dashboard, route *gin.RouterGroup, logger logging.Logger) controller.DashboardController {
	c := dashboardController{
		s, logger, mapper{},
	}
	c.logger.Debug("InitDashboardController")
	c.register(route)
	return &c
}

type dashboardController struct {
	dashboardService.Dashboard
	logger logging.Logger
	mapper
}

func (c *dashboardController) register(route *gin.RouterGroup) {

	route.GET("/chart/:type", c.Chart)
	route.GET("/chart/pools/:address/:type", c.ChartByPool)
	route.GET("/chart/tokens/:address/:type", c.ChartByToken)

	route.GET("/recent", c.Recent)

	route.GET("/statistics", c.Statistic)

	route.GET("/tokens", c.Tokens)
	route.GET("/tokens/:address", c.Token)

	route.GET("/txs", c.Txs)

	route.GET("/pools", c.Pools)
	route.GET("/pools/:address", c.Pool)

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
	recent, err := c.Dashboard.Recent()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.recentToRes(recent))
}

// Dashboard godoc
//
//	@Summary		Charts of Dezswap's Pool related a given token
//	@Description	get Charts data
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ChartRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//
// @Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
// @Param			address		path	string	true	"Token Address"
// @Param			type		path	string	true	"chart type"					Enums(volume, tvl, price)
// @Router			/dashboard/chart/tokens/{address}/{type} [get]
func (c *dashboardController) ChartByToken(ctx *gin.Context) {
	chartType := ToChartType(ctx.Param("type"))
	if chartType == ChartTypeNone {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid chart type"))
		return
	}

	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}

	addr := dashboardService.Addr(ctx.Param("address"))
	if len(addr) == 0 {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("must provide token address"))
		return
	}

	var err error
	var res ChartRes

	var chart dashboardService.TokenChart
	switch chartType {
	case ChartTypeVolume:
		chart, err = c.Dashboard.TokenVolumes(addr, dashboardService.Duration(duration))
	case ChartTypeTvl:
		chart, err = c.Dashboard.TokenTvls(addr, dashboardService.Duration(duration))
	case ChartTypePrice:
		chart, err = c.Dashboard.TokenPrices(addr, dashboardService.Duration(duration))
	default:
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("unsupported chart type"))
		return
	}

	if err == nil {
		res, err = c.tokenChartToChartRes(chart)
	}

	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// Dashboard godoc
//
//	@Summary		Charts of Dezswap's Pool
//	@Description	get Charts data
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ChartRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//
// @Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
// @Param			address		path	string	true	"Pool Address"
// @Param			type		path	string	true	"chart type"					Enums(volume, tvl, apr, fee)
// @Router			/dashboard/chart/pools/{address}/{type} [get]
func (c *dashboardController) ChartByPool(ctx *gin.Context) {
	chartType := ToChartType(ctx.Param("type"))
	if chartType == ChartTypeNone {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid chart type"))
		return
	}

	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}

	addr := dashboardService.Addr(ctx.Param("address"))
	if len(addr) == 0 {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("must provide pool address"))
		return
	}

	var err error
	var res ChartRes

	switch chartType {
	case ChartTypeVolume:
		var volumes dashboardService.Volumes
		volumes, err = c.Dashboard.VolumesOf(addr, duration)
		res = c.volumesToChartRes(volumes)
	case ChartTypeTvl:
		var tvls dashboardService.Tvls
		tvls, err = c.Dashboard.TvlsOf(addr, duration)
		res = c.tvlsToChartRes(tvls)
	case ChartTypeApr:
		var aprs dashboardService.Aprs
		aprs, err = c.Dashboard.AprsOf(addr, duration)
		res = c.aprsToChartRes(aprs)
	case ChartTypeFee:
		var fees dashboardService.Fees
		fees, err = c.Dashboard.FeesOf(addr, duration)
		res = c.feesToChartRes(fees)
	default:
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid chart type"))
		return
	}

	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// Dashboard godoc
//
//	@Summary		Charts of Dezswap's Pools
//	@Description	get Charts data
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ChartRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//
// @Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
// @Param			type		path	string	true	"chart type"					Enums(volume, tvl, apr, fee)
// @Router			/dashboard/chart/{type} [get]
func (c *dashboardController) Chart(ctx *gin.Context) {
	chartType := ToChartType(ctx.Param("type"))
	if chartType == ChartTypeNone {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid chart type"))
		return
	}

	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}

	var err error
	var res ChartRes

	switch chartType {
	case ChartTypeVolume:
		var volumes dashboardService.Volumes
		volumes, err = c.Dashboard.Volumes(duration)
		res = c.volumesToChartRes(volumes)
	case ChartTypeTvl:
		var tvls dashboardService.Tvls
		tvls, err = c.Dashboard.Tvls(duration)
		res = c.tvlsToChartRes(tvls)
	case ChartTypeApr:
		var aprs dashboardService.Aprs
		aprs, err = c.Dashboard.Aprs(duration)
		res = c.aprsToChartRes(aprs)
	case ChartTypeFee:
		var fees dashboardService.Fees
		fees, err = c.Dashboard.Fees(duration)
		res = c.feesToChartRes(fees)
	default:
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid chart type"))
		return
	}

	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, res)
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
	statistic, err := c.Dashboard.Statistic()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.statisticToRes(statistic))

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
//	@Param			token		query	string	false	"token address"
//	@Router			/dashboard/pools [get]
func (c *dashboardController) Pools(ctx *gin.Context) {
	token := ctx.Query("token")

	var pools dashboardService.Pools
	var err error
	if len(token) > 0 {
		pools, err = c.Dashboard.Pools(dashboardService.Addr(token))
		if err != nil {
			c.logger.Warn(err)
			httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
	} else {
		pools, err = c.Dashboard.Pools()
		if err != nil {
			c.logger.Warn(err)
			httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
	}

	res := c.poolsToRes(pools)
	ctx.JSON(http.StatusOK, res)
}

// Dashboard godoc
//
//	@Summary		Dezswap's Pool Detail
//	@Description	get Pool's detail information of dezswap
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PoolDetailRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Param			address		path	string	true	"Pool Address"
//	@Router			/dashboard/pools/{address} [get]
func (c *dashboardController) Pool(ctx *gin.Context) {
	address := ctx.Param("address")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
		return
	}

	poolDetail, err := c.Dashboard.PoolDetail(dashboardService.Addr(address))
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	res := c.poolDetailToRes(poolDetail)
	ctx.JSON(http.StatusOK, res)
}

// Dashboard godoc
//
//	@Summary		Dezswap's Token Stats
//	@Description	get Token data of dezswap (address, price, price_change, volume_24h,  volume_24h_change, volume_7d, volume_7d_change, tvl)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TokenRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		404	{object}	httputil.NotFoundError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Param			address		path	string	true	"token address"
//	@Router			/dashboard/tokens/{address} [get]
func (c *dashboardController) Token(ctx *gin.Context) {
	address := ctx.Param("address")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address address"))
		return
	}

	token, err := c.Dashboard.Token(dashboardService.Addr(address))
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	if string(token.Addr) != address {
		httputil.NewError(ctx, http.StatusNotFound, errors.New("token not found"))
		return
	}
	res := c.tokenToRes(token)

	ctx.JSON(http.StatusOK, res)
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
	tokens, err := c.Dashboard.Tokens()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	res := c.tokensToRes(tokens)

	ctx.JSON(http.StatusOK, res)
}

// Dashboard godoc
//
//	@Summary		Dezswap's Transactions
//	@Description	get Transactions data of dezswap
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TxsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Param			pool		query	string	false	"Pool address"
//	@Param			token		query	string	false	"Token address"
//	@Param			type		query	string	false	"Transaction type, default(all)"  Enums(swap, add, remove)
//	@Router			/dashboard/txs [get]
func (c *dashboardController) Txs(ctx *gin.Context) {

	pool := dashboardService.Addr(ctx.Query("pool"))
	token := dashboardService.Addr(ctx.Query("token"))
	txType := c.txTypeToServiceTxType(TxType(ctx.Query("type")))

	if len(pool) > 0 && len(token) > 0 {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid query, must choose one of (pool or token, not both)"))
		return
	}

	var txs dashboardService.Txs
	var err error
	if len(token) > 0 {
		txs, err = c.Dashboard.TxsOfToken(txType, token)
	} else if len(pool) > 0 {
		txs, err = c.Dashboard.Txs(txType, pool)
	} else {
		txs, err = c.Dashboard.Txs(txType)
	}
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	txsRes := c.txsToRes(txs)
	ctx.JSON(http.StatusOK, txsRes)
}
