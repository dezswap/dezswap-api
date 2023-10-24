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

type pairController struct {
	service.Getter[coingeckoService.Pair]
	logger logging.Logger
	pairMapper
}

func InitPairController(s service.Getter[coingeckoService.Pair], route *gin.RouterGroup, logger logging.Logger) controller.PairController {
	c := pairController{s, logger, pairMapper{}}
	c.register(route)
	return &c
}

func (c *pairController) register(route *gin.RouterGroup) {
	route.GET("/pairs", c.Pairs)
	route.GET("/pairs/:address", c.Pair)
}

// Coingecko godoc
//
//	@Summary		All Pairs
//	@Description	get Pairs
//	@Tags			coingecko
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PairsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/coingecko/pairs [get]
func (c *pairController) Pairs(ctx *gin.Context) {
	pairs, err := c.GetAll()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, c.pairsToRes(pairs))
}

// Coingecko godoc
//
//	@Summary		Get a pair
//	@Description	get Pair by Address
//	@Tags			coingecko
//	@Accept			json
//	@Produce		json
//	@Param			address	path		string	true	"Pair Address"
//	@Success		200		{object}	PairRes
//	@Failure		400		{object}	httputil.BadRequestError
//	@Failure		500		{object}	httputil.InternalServerError
//	@Router			/coingecko/pairs/{address} [get]
func (c *pairController) Pair(ctx *gin.Context) {
	address := ctx.Param("address")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address address"))
		return
	}

	pair, err := c.Get(address)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	if pair == nil {
		httputil.NewError(ctx, http.StatusNotFound, errors.New("pair not found"))
		return
	}

	ctx.JSON(http.StatusOK, c.pairToRes(*pair))
}
