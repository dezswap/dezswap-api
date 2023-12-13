package router

import (
	"net/http"
	"strconv"

	rs "github.com/dezswap/dezswap-api/api/service/router"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type routerController struct {
	rs.Router
	*mapper
	logger logging.Logger
}

type mapper struct{}

func InitRouterController(s rs.Router, route *gin.RouterGroup, logger logging.Logger) *routerController {
	c := routerController{s, &mapper{}, logger}
	c.register(route)
	return &c
}

func (c *routerController) register(route *gin.RouterGroup) {
	route.GET("/", c.Routes)
}

//	Routes godoc
//
//	@Tags			router
//	@Summary		All Routes
//	@Description	get routes based on the given token address
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	RoutesRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/routes [get]
//
// @Param			from		query		string	false	"Offer token address"
// @Param			to			query		string	false	"Ask token Address"
// @Param 			hopCount	query		int		false	"Number of hops between the starting token and the ending token"
func (c *routerController) Routes(ctx *gin.Context) {
	from := ctx.Query("from")
	to := ctx.Query("to")
	if from == "" && to == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("required from or to"))
		return
	}

	hopCount := 0
	if ctx.Query("hopCount") != "" {
		var err error
		hopCount, err = strconv.Atoi(ctx.Query("hopCount"))
		if err != nil {
			httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid hop count"))
			return
		}
	}

	// full path
	if from != "" && to != "" {
		routes, err := c.Router.Routes(from, to, hopCount)
		if err != nil {
			c.logger.Warn(err)
			httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}
		res := c.routesToRes(routes, from, false)
		ctx.JSON(http.StatusOK, res)
		return
	}

	addr := from
	reverse := false
	if addr == "" {
		addr = to
		reverse = true
	}

	routes, err := c.Router.RoutesOfToken(addr, hopCount, reverse)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	res := c.routesToRes(routes, addr, reverse)
	ctx.JSON(http.StatusOK, res)
}

func (m *mapper) routesToRes(routes []rs.Route, addr string, reverse bool) []RouteRes {
	res := make([]RouteRes, len(routes))
	for i, r := range routes {
		res[i] = RouteRes{
			From:     addr,
			To:       r.To,
			HopCount: r.HopCount,
			Route:    r.Route,
		}
		if reverse {
			res[i].To = addr
			res[i].From = r.To
		}
	}
	return res
}
