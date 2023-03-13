package controller

import (
	"net/http"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

type pairController struct {
	service.Getter[service.Pair]
}

func InitPairController(s service.Getter[service.Pair], route *gin.RouterGroup) PairController {
	c := pairController{s}
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
//	@Tags			pairs
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PairsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		404	{object}	httputil.NotFoundError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/pairs [get]
func (c *pairController) Pairs(ctx *gin.Context) {
	res, err := c.GetAll()
	if err != nil {
		httputil.NewError(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// Pair godoc
//
//	@Summary		Get a pair
//	@Description	get Pair by Address
//	@Tags			pairs
//	@Accept			json
//	@Produce		json
//	@Param			address	path		string	true	"Pair Address"
//	@Success		200		{object}	PairRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		404	{object}	httputil.NotFoundError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/pairs/{address} [get]
func (c *pairController) Pair(ctx *gin.Context) {
	res := PairRes{}
	// if err != nil {
	// 	httputil.NewError(ctx, http.StatusNotFound, err)
	// 	return
	// }
	ctx.JSON(http.StatusOK, res)
}
