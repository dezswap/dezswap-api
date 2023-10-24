package controller

import (
	"errors"
	"net/http"

	"github.com/dezswap/dezswap-api/api/service"
	"github.com/dezswap/dezswap-api/pkg/httputil"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
)

type statController struct {
	service.Getter[service.PairStats]
	logger logging.Logger
	statMapper
}

func InitStatController(s service.Getter[service.PairStats], route *gin.RouterGroup, logger logging.Logger) StatController {
	c := statController{s, logger, statMapper{}}
	c.register(route)
	return &c
}

func (c *statController) register(route *gin.RouterGroup) {
	route.GET("/stats", c.Stats)
	route.GET("/stats/:period", c.Stat)
}

// Stats godoc
//
//	@Summary		All pair stats
//	@Description	get pair stats
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	StatsRes
//	@Failure		400	{object}	httputil.BadRequestError
//	@Failure		500	{object}	httputil.InternalServerError
//	@Router			/stats [get]
func (c *statController) Stats(ctx *gin.Context) {
	stats, err := c.GetAll()
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}
	res := c.statsToRes(stats)
	ctx.JSON(http.StatusOK, res)
}

// Stat godoc
//
//	@Summary		Get a stat
//	@Description	get a pair stat by period
//	@Accept			json
//	@Produce		json
//	@Param			period	path		string	true "period 24h,7d,1mon"	Enums(24h,7d,1mon)
//	@Success		200		{object}	StatRes
//	@Failure		400		{object}	httputil.BadRequestError
//	@Failure		404		{object}	httputil.NotFoundError
//	@Failure		500		{object}	httputil.InternalServerError
//	@Router			/stats/{period} [get]
func (c *statController) Stat(ctx *gin.Context) {
	period := ctx.Param("period")
	if period == "" {
		httputil.NewError(ctx, http.StatusBadRequest, errors.New("invalid address"))
		return
	}

	stat, err := c.Get(period)
	if err != nil {
		c.logger.Warn(err)
		httputil.NewError(ctx, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	if stat == nil {
		httputil.NewError(ctx, http.StatusNotFound, errors.New("stat not found"))
		return
	}

	res := c.statToRes(*stat)
	ctx.JSON(http.StatusOK, res)
}
