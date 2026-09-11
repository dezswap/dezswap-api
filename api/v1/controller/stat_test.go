package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dezswap/dezswap-api/api/v1/service"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// stubStatGetter answers every lookup with the same result.
type stubStatGetter struct {
	stats *service.PairStats
	err   error
}

func (s stubStatGetter) Get(string) (*service.PairStats, error) { return s.stats, s.err }
func (s stubStatGetter) GetAll() ([]service.PairStats, error)   { return nil, s.err }

// alertRecorder counts entries logged at WarnLevel or above, which is where the
// Sentry hook is attached: an entry here is an alert raised.
type alertRecorder struct{ count int }

func (r *alertRecorder) Levels() []logrus.Level {
	return []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel}
}

func (r *alertRecorder) Fire(*logrus.Entry) error {
	r.count++
	return nil
}

func recordingLogger() (logging.Logger, *alertRecorder) {
	recorder := &alertRecorder{}
	logger := logrus.New()
	logger.Out = io.Discard
	logger.AddHook(recorder)

	return logger.WithField("test", true), recorder
}

func serveStat(t *testing.T, getter stubStatGetter, period string) (*httptest.ResponseRecorder, *alertRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	logger, alerts := recordingLogger()
	engine := gin.New()
	InitStatController(getter, engine.Group(""), logger)

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stats/"+period, nil))

	return rec, alerts
}

// The period is a path segment, so a caller can invent one. Answered as a fault, a
// scanner walking the path raised one alert per request on its way to a 500.
func TestStat_UnsupportedPeriodIsRejectedWithoutAnAlert(t *testing.T) {
	rec, alerts := serveStat(t,
		stubStatGetter{err: fmt.Errorf("statService.Get: %w", service.ErrInvalidKey)}, "1yr")

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Zero(t, alerts.count, "an invented period must not raise an alert")
}

// The branch above keys on the error, not on the status, so a database that has
// gone away has to keep reaching the log.
func TestStat_FailureIsStillReported(t *testing.T) {
	rec, alerts := serveStat(t, stubStatGetter{err: errors.New("connection refused")}, "24h")

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, 1, alerts.count)
}
