package controller

import (
	"github.com/dezswap/dezswap-api/pkg"
	"net/http"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type pairController struct {
	service.Getter[service.Pair]
	logger logging.Logger
	pairMapper
}

func InitPairController(s service.Getter[service.Pair], route *gin.RouterGroup, networkMetadata pkg.NetworkMetadata, logger logging.Logger) PairController {
	c := pairController{s, logger, pairMapper{networkMetadata}}
	c.register(route)
	return &c
}

func (c *pairController) register(route *gin.RouterGroup) {
	route.GET("/pairs", c.Pairs)
	route.GET("/pairs/:address", c.Pair)
}

// Pairs godoc
//
//	@Summary		All Pairs
//	@Description	get Pairs
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PairsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/pairs [get]
func (c *pairController) Pairs(ctx *gin.Context) {
	pairs, err := c.GetAll()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	res := c.pairsToRes(pairs)
	ctx.JSON(http.StatusOK, res)
}

// Pair godoc
//
//	@Summary		Get a pair
//	@Description	get Pair by Address
//	@Accept			json
//	@Produce		json
//	@Param			address	path		string	true	"Pair Address"
//	@Success		200		{object}	PairRes
//	@Failure		400		{object}	httputil.BadRequestError
//	@Failure		404		{object}	httputil.NotFoundError
//	@Failure		500		{object}	httputil.InternalServerError
//	@Router			/pairs/{address} [get]
func (c *pairController) Pair(ctx *gin.Context) {
	address := ctx.Param("address")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
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

	res := c.pairToRes(*pair)
	ctx.JSON(http.StatusOK, res)
}
