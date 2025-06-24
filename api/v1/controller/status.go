package controller

import (
	"github.com/dezswap/dezswap-api/api/v1/service"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
	r.GET("/health", c.Health)
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

// Health godoc
// @Summary      Health check
// @Description  Checks overall service and dependency health
// @Tags         status
// @Produce      json
// @Success      200 {object} HealthResponse
// @Router       /health [get]
func (c *statusController) Health(ctx *gin.Context) {
	status := "ok"

	checks := []struct {
		name  string
		check func() error
	}{
		{name: "db", check: c.service.CheckDB},
		{name: "cache", check: c.service.CheckCache},
	}

	deps := make([]HealthDependency, 0, len(checks))
	for _, d := range checks {
		depStatus := "ok"
		if err := d.check(); err != nil {
			status = "unhealthy"
			depStatus = "error: " + err.Error()
		}

		deps = append(deps, HealthDependency{
			Name:   d.name,
			Status: depStatus,
		})
	}

	res := HealthResponse{
		Status:       status,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: deps,
	}

	ctx.JSON(http.StatusOK, res)
}
