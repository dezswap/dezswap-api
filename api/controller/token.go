package controller

import (
	"net/http"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

type tokenController struct {
	service.Getter[service.Token]
}

func InitTokenController(s service.Getter[service.Token], route *gin.RouterGroup) TokenController {
	c := tokenController{s}
	c.register(route)
	return &c
}

func (c *tokenController) register(route *gin.RouterGroup) {
	route.GET("/tokens", c.Tokens)
	route.GET("/tokens/:address", c.Token)
}

// Tokens godoc
//
//	@Summary		All Tokens
//	@Description	get Tokens
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TokensRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		404	{object}	httputil.NotFoundError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/tokens [get]
func (c *tokenController) Tokens(ctx *gin.Context) {
	res, err := c.GetAll()
	if err != nil {
		httputil.NewError(ctx, http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// Token godoc
//
//	@Summary		Get a token
//	@Description	get Token by Address
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Param			address	path		string	true	"Token Address"
//	@Success		200		{object}	TokenRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		404	{object}	httputil.NotFoundError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/tokens/{address} [get]
func (c *tokenController) Token(ctx *gin.Context) {
	res := TokenRes{}
	// if err != nil {
	// 	httputil.NewError(ctx, http.StatusNotFound, err)
	// 	return
	// }
	ctx.JSON(http.StatusOK, res)
}
