package dashboard

import (
	"net/http"
	"strconv"
	"strings"

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

	route.GET("/recent", c.Recent)

	route.GET("/statistics", c.Statistic)

	route.GET("/token/:address", c.Token)
	route.GET("/tokens", c.Tokens)
	route.GET("/token_chart/:address", c.TokenChart)

	route.GET("/txs/:poolAddress", c.TxsOfPool)
	route.GET("/txs", c.Txs)

	route.GET("/pools", c.Pools)
	route.GET("/volumes", c.Volumes)
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
//	@Summary		TVLs of dezswap selected duration
//	@Description	get TVLs
//	@Tags			dashboard
//	@Param			duration	query	string	false	"default(empty) value is all"	Enums(year, quarter, month)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TvlsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/TVLs [get]
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

	token, err := c.Dashboard.Token(dashboardService.Addr(address))
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
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
//	@Param			sort		query	string	false	"sorting e.g. price_change:asc"
//	@Param			limit		query	int		false	"the number of returning data"
//	@Param			offset		query	int		false	"the offset of returning data"
//	@Router			/dashboard/tokens [get]
func (c *dashboardController) Tokens(ctx *gin.Context) {
	var item string
	var limit, offset int
	ascending := true
	sort := ctx.Query("sort")
	if len(sort) > 0 {
		orderItem := strings.Split(sort, ":")
		if len(orderItem) > 1 {
			item = orderItem[0]
			if orderItem[1] == "desc" {
				ascending = false
			}
		}
	}
	limitStr := ctx.Query("limit")
	if len(limitStr) > 0 {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid limit"))
		}
	}
	offsetStr := ctx.Query("offset")
	if len(offsetStr) > 0 {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid offset"))
		}
	}

	tokens, err := c.Dashboard.Tokens(item, ascending, limit, offset)
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
//	@Summary		Dezswap's Transactions of a pool
//	@Description	get Transactions data of dezswap (action, totalValue, asset0amount, asset1amount, time)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TxsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/txs/{poolAddress} [get]
func (c *dashboardController) TxsOfPool(ctx *gin.Context) {

	address := ctx.Param("poolAddress")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
		return
	}

	txs, err := c.Dashboard.Txs()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	if len(txs) == 0 {
		httputil.NewError(ctx, http.StatusNotFound, errors.New("pool not found"))
		return
	}

	txsRes := c.txsToRes(txs)
	ctx.JSON(http.StatusOK, txsRes)
}

// Dashboard godoc
//
//	@Summary		Dezswap's Transactions
//	@Description	get Transactions data of dezswap (action, totalValue, asset0amount, asset1amount, time)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TxsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/dashboard/txs [get]
func (c *dashboardController) Txs(ctx *gin.Context) {
	txs, err := c.Dashboard.Txs()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	txsRes := c.txsToRes(txs)
	ctx.JSON(http.StatusOK, txsRes)
}
