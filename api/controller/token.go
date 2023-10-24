package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
)

type tokenController struct {
	service.Getter[service.Token]
	logger logging.Logger
	tokenMapper
}

func InitTokenController(s service.Getter[service.Token], route *gin.RouterGroup, logger logging.Logger) TokenController {
	c := tokenController{s, logger, tokenMapper{}}
	c.register(route)
	return &c
}

func (c *tokenController) register(route *gin.RouterGroup) {
	route.GET("/tokens", c.Tokens)
	route.GET("/tokens/*address", c.Token)
}

// Tokens godoc
//
//	@Summary		All Tokens
//	@Description	get Tokens
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	TokensRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		404	{object}	httputil.NotFoundError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/tokens [get]
func (c *tokenController) Tokens(ctx *gin.Context) {
	tokens, err := c.GetAll()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	res := c.tokensToRes(tokens)
	ctx.JSON(http.StatusOK, res)
}

// Token godoc
//
//	@Summary		Get a token
//	@Description	get Token by Address
//	@Accept			json
//	@Produce		json
//	@Param			address	path		string	true	"Token Address"
//	@Success		200		{object}	TokenRes
//	@Failure		400		{object}	httputil.BadRequestError
//	@Failure		404		{object}	httputil.NotFoundError
//	@Failure		500		{object}	httputil.InternalServerError
//	@Router			/tokens/{address} [get]
func (c *tokenController) Token(ctx *gin.Context) {
	address := ctx.Param("address")
	if address == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
		return
	}

	address = strings.TrimPrefix(address, "/")
	token, err := c.Get(address)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	if token == nil {
		httputil.NewError(ctx, http.StatusNotFound, errors.New("token not found"))
		return
	}

	res := c.tokenToRes(*token)
	ctx.JSON(http.StatusOK, res)
}
