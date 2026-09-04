package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// stubStatusService reports whatever each dependency was set to fail with. The
// optional channels let a test observe both checks starting before either one is
// allowed to finish.
type stubStatusService struct {
	db      error
	cache   error
	started chan<- string
	release <-chan struct{}
	// dbPanic stands in for a driver that faults rather than returning an error.
	dbPanic any
}

func (s stubStatusService) waitForRelease(name string) {
	if s.started != nil {
		s.started <- name
	}
	if s.release != nil {
		<-s.release
	}
}

func (s stubStatusService) CheckDB(context.Context) error {
	s.waitForRelease("db")
	if s.dbPanic != nil {
		panic(s.dbPanic)
	}
	return s.db
}

func (s stubStatusService) CheckCache(context.Context) error {
	s.waitForRelease("cache")
	return s.cache
}

func newHealthEngine(service stubStatusService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	InitStatusController(service, engine.Group(""), "v-test", logging.Discard)
	return engine
}

func serveHealth(t *testing.T, service stubStatusService) (*httptest.ResponseRecorder, HealthResponse) {
	t.Helper()

	rec := httptest.NewRecorder()
	newHealthEngine(service).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	var res HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	return rec, res
}

// dependencyStatus is what the response reported for one dependency by name.
func dependencyStatus(t *testing.T, res HealthResponse, name string) string {
	t.Helper()

	for _, d := range res.Dependencies {
		if d.Name == name {
			return d.Status
		}
	}

	require.Failf(t, "dependency missing", "%q is not in the response", name)
	return ""
}

func TestHealth_AllDependenciesUp(t *testing.T) {
	rec, res := serveHealth(t, stubStatusService{})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, healthOk, res.Status)
	require.Equal(t, healthOk, dependencyStatus(t, res, "db"))
	require.Equal(t, healthOk, dependencyStatus(t, res, "cache"))
}

// The endpoint backs a readiness probe, so answering anything but 2xx here would
// take every instance out of rotation at once over a dependency none of them needs
// to serve a request.
func TestHealth_CacheFailureStaysReady(t *testing.T) {
	rec, res := serveHealth(t, stubStatusService{cache: errors.New("connection refused")})

	require.Equal(t, http.StatusOK, rec.Code, "a cache outage must not make the instance unready")
	require.Equal(t, healthDegraded, res.Status)
	require.Equal(t, healthOk, dependencyStatus(t, res, "db"))
	require.Equal(t, healthError, dependencyStatus(t, res, "cache"))
}

// The database is the one dependency nothing can be served without, and a probe
// acts on the status code alone -- the body saying "unhealthy" reaches no one.
func TestHealth_DbFailureIsUnready(t *testing.T) {
	rec, res := serveHealth(t, stubStatusService{db: errors.New("connection refused")})

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.Equal(t, healthUnhealthy, res.Status)
	require.Equal(t, healthError, dependencyStatus(t, res, "db"))
	require.Equal(t, healthOk, dependencyStatus(t, res, "cache"))
}

// A critical failure has to hold whichever order the results come back in, so that
// a degraded dependency cannot write over it.
func TestHealth_BothFailingReportsUnhealthy(t *testing.T) {
	rec, res := serveHealth(t, stubStatusService{
		db:    errors.New("db is gone"),
		cache: errors.New("cache is gone"),
	})

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.Equal(t, healthUnhealthy, res.Status)
}

// The endpoint is unauthenticated, and a driver names the user, database, host and
// port it failed to reach. None of that belongs in a response anyone can ask for.
func TestHealth_FailureDetailStaysOutOfTheResponse(t *testing.T) {
	const detail = "user=app database=dezswap host=10.0.1.5 port=5432: connection refused"

	rec, res := serveHealth(t, stubStatusService{
		db:    errors.New(detail),
		cache: errors.New(detail),
	})

	require.Equal(t, healthUnhealthy, res.Status)
	require.NotContains(t, rec.Body.String(), detail)
	require.NotContains(t, rec.Body.String(), "10.0.1.5")
	require.NotContains(t, rec.Body.String(), "user=app")
}

// The checks run on their own goroutines, past the reach of gin's recovery
// middleware. A panic left to escape one would take the whole process down rather
// than fail the single request it happened on -- a nil database handle is enough to
// cause it.
func TestHealth_PanickingCheckFailsTheDependencyNotTheProcess(t *testing.T) {
	rec, res := serveHealth(t, stubStatusService{dbPanic: "nil pointer in driver"})

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	require.Equal(t, healthUnhealthy, res.Status)
	require.Equal(t, healthError, dependencyStatus(t, res, "db"))
	require.Equal(t, healthOk, dependencyStatus(t, res, "cache"))
	require.NotContains(t, rec.Body.String(), "nil pointer in driver")
}

// If checks run one after another, the first one blocks on release and the second
// one can never announce that it started. Requiring both announcements before
// release proves overlap without relying on scheduler-sensitive elapsed time.
func TestHealth_ChecksRunConcurrently(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()

	engine := newHealthEngine(stubStatusService{
		started: started,
		release: release,
	})
	rec := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
		close(done)
	}()

	seen := make(map[string]bool, 2)
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	for len(seen) < 2 {
		select {
		case name := <-started:
			seen[name] = true
		case <-timer.C:
			t.Fatal("both dependency checks did not start before either was released")
		}
	}

	close(release)
	released = true
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("health request did not complete after dependency checks were released")
	}

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, seen["db"])
	require.True(t, seen["cache"])
}
