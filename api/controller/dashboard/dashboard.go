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

	route.GET("/aprs/:pool", c.APRsOf)
	route.GET("/aprs", c.APRs)

	route.GET("/tvls/:pool", c.TVLsOf)
	route.GET("/tvls", c.TVLs)

	route.GET("/volumes/:pool", c.VolumesOf)
	route.GET("/volumes", c.Volumes)

	route.GET("/fees/:pool", c.FeesOf)
	route.GET("/fees", c.Fees)

	route.GET("/recent", c.Recent)

	route.GET("/statistics", c.Statistic)

	route.GET("/token/:address", c.Token)
	route.GET("/tokens", c.Tokens)
	route.GET("/token_chart/:address", c.TokenChart)

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
//	@Summary		Charts of Dezswap's Pool
//	@Description	get Recent
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ChartRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//
// @Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
// @Param			pool		query	string	false	"Pool Address"
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

	poolAddr, ok := ctx.GetQuery("pool")
	addr := dashboardService.Addr(poolAddr)

	var err error
	var res ChartRes

	switch chartType {
	case ChartTypeVolume:
		var volumes dashboardService.Volumes
		if ok {
			volumes, err = c.Dashboard.VolumesOf(addr, duration)
		} else {
			volumes, err = c.Dashboard.Volumes(duration)
		}
		res = c.volumesToChartRes(volumes)
	case ChartTypeTvl:
		var tvls dashboardService.Tvls
		if ok {
			tvls, err = c.Dashboard.TvlsOf(addr, duration)
		} else {
			tvls, err = c.Dashboard.Tvls(duration)
		}
		res = c.tvlsToChartRes(tvls)
	case ChartTypeApr:
		var aprs dashboardService.Aprs
		if ok {
			aprs, err = c.Dashboard.AprsOf(addr, duration)
		} else {
			aprs, err = c.Dashboard.Aprs(duration)
		}
		res = c.aprsToChartRes(aprs)
	case ChartTypeFee:
		var fees dashboardService.Fees
		if ok {
			fees, err = c.Dashboard.FeesOf(addr, duration)
		} else {
			fees, err = c.Dashboard.Fees(duration)
		}
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
	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}
	volumes, err := c.Dashboard.Volumes(duration)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.volumesToRes(volumes))
}

// Dashboard godoc
//
//	@Summary		Pool's Volumes of user selected duration
//	@Description	get Volumes
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	VolumesRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//
// @Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
// @Param			pool	path		string	true	"Pool Address"
//
//	@Router			/dashboard/volumes/{pool} [get]
func (c *dashboardController) VolumesOf(ctx *gin.Context) {
	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}
	address := ctx.Param("pool")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
		return
	}

	volumes, err := c.Dashboard.VolumesOf(dashboardService.Addr(address), duration)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.volumesToRes(volumes))
}

// Dashboard godoc
//
//	@Summary		Fees of user selected duration
//	@Description	get Fees
//	@Tags			dashboard
//	@Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	FeesRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/fees [get]
func (c *dashboardController) Fees(ctx *gin.Context) {
	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}
	fees, err := c.Dashboard.Fees(duration)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.feesToRes(fees))
}

// Dashboard godoc
//
//	@Summary		Pool's Fees of user selected duration
//	@Description	get Fees
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	FeesRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//
// @Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
// @Param			pool	path		string	true	"Pool Address"
//
//	@Router			/dashboard/fees/{pool} [get]
func (c *dashboardController) FeesOf(ctx *gin.Context) {
	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}
	address := ctx.Param("pool")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
		return
	}

	fees, err := c.Dashboard.FeesOf(dashboardService.Addr(address), duration)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.feesToRes(fees))
}

// Dashboard godoc
//
//	@Summary		TVLs of dezswap selected duration
//	@Description	get TVLs
//	@Tags			dashboard
//	@Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TvlsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/tvls [get]
func (c *dashboardController) TVLs(ctx *gin.Context) {
	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}
	tvls, err := c.Dashboard.Tvls(duration)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.tvlsToRes(tvls))
}

// Dashboard godoc
//
//	@Summary		TVLs of dezswap selected duration
//	@Description	get TVLs
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TvlsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//
// @Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
// @Param			pool	path		string	true	"Pool Address"
// @Router			/dashboard/tvls/{pool} [get]
func (c *dashboardController) TVLsOf(ctx *gin.Context) {
	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}

	address := ctx.Param("pool")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
		return
	}

	tvls, err := c.Dashboard.TvlsOf(dashboardService.Addr(address), duration)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.tvlsToRes(tvls))
}

// Dashboard godoc
//
//	@Summary		Dezswap's Pool's APRs
//	@Description	get APRs of a dezswap pool
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
//	@Param			pool	path		string	true	"Pool Address"
//	@Success		200	{object}	AprsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/aprs/{pool} [get]
func (c *dashboardController) APRsOf(ctx *gin.Context) {
	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}

	address := ctx.Param("pool")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
		return
	}

	apr, err := c.Dashboard.AprsOf(dashboardService.Addr(address), duration)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.aprsToRes(apr))
}

// Dashboard godoc
//
//	@Summary		APRs of dezswap selected duration
//	@Description	get APRs
//	@Tags			dashboard
//	@Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	AprsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/aprs [get]
func (c *dashboardController) APRs(ctx *gin.Context) {
	duration := dashboardService.Duration(ctx.Query("duration"))
	if len(duration) == 0 {
		duration = dashboardService.All
	}
	apr, err := c.Dashboard.Aprs(duration)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	ctx.JSON(http.StatusOK, c.aprsToRes(apr))
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
//	@Param			pool	path		string	true	"Pool Address"
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
//	@Router			/dashboard/token/{address} [get]
func (c *dashboardController) Token(ctx *gin.Context) {
	address := ctx.Param("address")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address address"))
		return
	}

	tokens, err := c.Dashboard.Tokens(dashboardService.Addr(address))
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	if len(tokens) == 0 {
		httputil.NewError(ctx, http.StatusNotFound, errors.New("token not found"))
		return
	}
	res := c.tokenToRes(tokens[0])

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
//	@Summary		Dezswap's Token Chart Data
//	@Description	get Token' chart data of Dezswap by designated interval
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TokenChart
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/token_chart/{address} [get]
//	@Param			address		path	string	true	"token address"
//	@Param			data		query	string	true	"chart data type"				Enums(volume, tvl, price)
//	@Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
func (c *dashboardController) TokenChart(ctx *gin.Context) {
	address := ctx.Param("address")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address address"))
		return
	}

	data := ctx.Query("data")
	duration := ctx.Query("duration")
	if len(duration) == 0 {
		duration = "all"
	}

	var chart dashboardService.TokenChart
	var err error
	switch data {
	case "volume":
		chart, err = c.Dashboard.TokenVolumes(dashboardService.Addr(address), dashboardService.Duration(duration))
		if err != nil {
			c.logger.Warn(err)
			httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
	case "tvl":
		chart, err = c.Dashboard.TokenTvls(dashboardService.Addr(address), dashboardService.Duration(duration))
		if err != nil {
			c.logger.Warn(err)
			httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
	case "price":
		chart, err = c.Dashboard.TokenPrices(dashboardService.Addr(address), dashboardService.Duration(duration))
		if err != nil {
			c.logger.Warn(err)
			httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
	default:
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("unsupported data type"))
		return
	}

	res := c.tokenChartToRes(chart)
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
//	@Router			/dashboard/txs [get]
func (c *dashboardController) Txs(ctx *gin.Context) {

	pool := dashboardService.Addr(ctx.Query("pool"))
	token := dashboardService.Addr(ctx.Query("token"))

	if len(pool) > 0 && len(token) > 0 {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid query, must choose one of (pool or token, not both)"))
		return
	}

	var txs dashboardService.Txs
	var err error
	if len(token) > 0 {
		txs, err = c.Dashboard.TxsOfToken(token)
	} else if len(pool) > 0 {
		txs, err = c.Dashboard.Txs(pool)
	} else {
		txs, err = c.Dashboard.Txs()
	}
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	txsRes := c.txsToRes(txs)
	ctx.JSON(http.StatusOK, txsRes)
}
