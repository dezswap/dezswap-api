package controller

import (
	"github.com/dezswap/dezswap-api/api/v1/service"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"net/http"
)

type statusController struct {
	service service.StatusService
	version string
	logger  logging.Logger
}

func InitStatusController(service service.StatusService, r *gin.RouterGroup, version string, logger logging.Logger) StatusController {
	c := statusController{service, version, logger}
	c.register(r)
	return &c
}

func (c *statusController) register(r *gin.RouterGroup) {
	r.GET("/version", c.Version)
}

// Version godoc
// @Summary      Get application version
// @Description  Returns the current application version
// @Tags         status
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /version [get]
func (c *statusController) Version(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"version": c.version,
	})
}
