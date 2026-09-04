package controller

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dezswap/dezswap-api/api/v1/service"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
)

const (
	healthOk        = "ok"
	healthDegraded  = "degraded"
	healthUnhealthy = "unhealthy"
	healthError     = "error"
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
// @Description  Checks overall service and dependency health. Answers 503 once a
// @Description  dependency the API cannot serve without is unreachable, so the
// @Description  endpoint can back a readiness probe directly.
// @Tags         status
// @Produce      json
// @Success      200 {object} HealthResponse
// @Failure      503 {object} HealthResponse
// @Router       /health [get]
func (c *statusController) Health(ctx *gin.Context) {
	checks := []struct {
		name     string
		critical bool
		check    func(context.Context) error
	}{
		{name: "db", critical: true, check: c.service.CheckDB},
		{name: "cache", critical: false, check: c.service.CheckCache},
	}

	reqCtx := ctx.Request.Context()
	errs := make([]error, len(checks))

	var wg sync.WaitGroup
	for i, d := range checks {
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					errs[i] = fmt.Errorf("%s check panicked: %v", d.name, r)
				}
			}()

			errs[i] = d.check(reqCtx)
		})
	}
	wg.Wait()

	status := healthOk
	deps := make([]HealthDependency, 0, len(checks))
	for i, d := range checks {
		depStatus := healthOk
		if err := errs[i]; err != nil {
			// The response says only that it failed, so the reason has to be recorded
			// somewhere an operator can still reach it.
			c.logger.Debugf("health: %s check failed: %v", d.name, err)
			depStatus = healthError

			switch {
			case d.critical:
				status = healthUnhealthy
			case status == healthOk:
				// Reported, but not at the cost of overwriting a critical failure that
				// has already been found.
				status = healthDegraded
			}
		}

		deps = append(deps, HealthDependency{
			Name:   d.name,
			Status: depStatus,
		})
	}

	code := http.StatusOK
	if status == healthUnhealthy {
		code = http.StatusServiceUnavailable
	}

	res := HealthResponse{
		Status:       status,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: deps,
	}

	ctx.JSON(code, res)
}
