package coingecko

import (
	"net/http"

	"github.com/dezswap/dezswap-api/api/controller"

	"github.com/dezswap/dezswap-api/api/service"
	coingeckoService "github.com/dezswap/dezswap-api/api/service/coingecko"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type tickerController struct {
	service.Getter[coingeckoService.Ticker]
	logger logging.Logger
	tickerMapper
}

func InitTickerController(s service.Getter[coingeckoService.Ticker], route *gin.RouterGroup, logger logging.Logger) controller.TickerController {
	c := tickerController{s, logger, tickerMapper{}}
	c.register(route)
	return &c
}

func (c *tickerController) register(route *gin.RouterGroup) {
	route.GET("/tickers", c.Tickers)
	route.GET("/tickers/:id", c.Ticker)
}

// Coingecko godoc
//
//	@Summary		All Tickers
//	@Description	get Tickers
//	@Tags			coingecko
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TickersRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/coingecko/tickers [get]
func (c *tickerController) Tickers(ctx *gin.Context) {
	tickers, err := c.GetAll()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	res, err := c.tickersToRes(tickers)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// Coingecko godoc
//
//	@Summary		Get a ticker
//	@Description	get Ticker by Id
//	@Tags			coingecko
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Ticker ID"
//	@Success		200	{object}	TickerRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/coingecko/tickers/{id} [get]
func (c *tickerController) Ticker(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid ticker address"))
		return
	}

	ticker, err := c.Get(id)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	if ticker == nil {
		httputil.NewError(ctx, http.StatusNotFound, errors.New("ticker not found"))
		return
	}

	res, err := c.tickerToRes(*ticker)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, res)
}
