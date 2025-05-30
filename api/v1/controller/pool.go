package controller

import (
	"errors"
	service2 "github.com/dezswap/dezswap-api/api/v1/service"
	"github.com/dezswap/dezswap-api/pkg"
	"net/http"

	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
)

type poolController struct {
	service2.Getter[service2.Pool]
	logger logging.Logger
	poolMapper
}

func InitPoolController(s service2.Getter[service2.Pool], route *gin.RouterGroup, networkMetadata pkg.NetworkMetadata, logger logging.Logger) PoolController {
	c := poolController{s, logger, poolMapper{networkMetadata}}
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
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	PoolsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/pools [get]
func (c *poolController) Pools(ctx *gin.Context) {
	pools, err := c.GetAll()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	res := c.poolsToRes(pools)
	ctx.JSON(http.StatusOK, res)
}

// Pool godoc
//
//	@Summary		Get a pool
//	@Description	get Pool by Address
//	@Accept			json
//	@Produce		json
//	@Param			address	path		string	true	"Pool Address"
//	@Success		200		{object}	PoolRes
//	@Failure		400		{object}	httputil.BadRequestError
//	@Failure		404		{object}	httputil.NotFoundError
//	@Failure		500		{object}	httputil.InternalServerError
//	@Router			/pools/{address} [get]
func (c *poolController) Pool(ctx *gin.Context) {
	address := ctx.Param("address")
	pool, err := c.Get(address)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	if pool == nil {
		httputil.NewError(ctx, http.StatusNotFound, errors.New("pool not found"))
		return
	}

	res := c.poolToRes(*pool)
	ctx.JSON(http.StatusOK, res)
}
