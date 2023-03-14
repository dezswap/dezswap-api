package controller

import (
	"net/http"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
)

type poolController struct {
	service.Getter[service.Pool]
	logging.Logger
	poolMapper
}

func InitPoolController(s service.Getter[service.Pool], route *gin.RouterGroup, logger logging.Logger) PoolController {
	c := poolController{s, logger, poolMapper{}}
	c.register(route)
	return &c
}

func (c *poolController) register(route *gin.RouterGroup) {
	route.GET("/pools", c.Pools)
	route.GET("/pools/:address", c.Pool)
}

// Pools godoc
//
//	@Summary		All Pools
//	@Description	get Pools
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PoolsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		404	{object}	httputil.NotFoundError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/pools [get]
func (c *poolController) Pools(ctx *gin.Context) {
	pools, err := c.GetAll()
	if err != nil {
		httputil.NewError(ctx, http.StatusNotFound, err)
		return
	}
	res := c.poolsToRes(pools)
	ctx.JSON(http.StatusOK, res)
}

// Pool godoc
//
//	@Summary		Get a pool
//	@Description	get Pool by Address
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Param			address	path		string	true	"Pool Address"
//	@Success		200		{object}	PoolRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		404	{object}	httputil.NotFoundError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/pools/{address} [get]
func (c *poolController) Pool(ctx *gin.Context) {
	address := ctx.Param("address")
	pool, err := c.Get(address)
	if err != nil {
		httputil.NewError(ctx, http.StatusNotFound, err)
		return
	}
	res := c.poolToRes(pool)
	ctx.JSON(http.StatusOK, res)
}
