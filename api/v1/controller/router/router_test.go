package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	rs "github.com/dezswap/dezswap-api/api/v1/service/router"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// stubRouter counts the lookups that got past the handler. The route is not cached,
// so each one of those is a query of its own.
type stubRouter struct{ lookups int }

func (s *stubRouter) RoutesOfToken(string, int, bool) ([]rs.Route, error) {
	s.lookups++
	return nil, nil
}

func (s *stubRouter) Routes(string, string, int) ([]rs.Route, error) {
	s.lookups++
	return nil, nil
}

func serveRoutes(t *testing.T, query string) (*httptest.ResponseRecorder, *stubRouter) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	service := &stubRouter{}
	engine := gin.New()
	InitRouterController(service, engine.Group("/routes"), logging.Discard)

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/routes"+query, nil))

	return rec, service
}

// An unknown address answers 200 with an empty list, so nothing upstream tells these
// apart from a real lookup and every one of them would be paid for in full.
func TestRoutes_AnAddressThatCouldNotBeOneNeverReachesTheService(t *testing.T) {
	for _, address := range []string{
		"'; select 1--",
		"../../etc/passwd",
		"a",
		strings.Repeat("x", 200),
	} {
		for _, query := range []string{
			"?from=" + url.QueryEscape(address) + "&to=xpla1def",
			"?from=xpla1abc&to=" + url.QueryEscape(address),
		} {
			rec, service := serveRoutes(t, query)

			require.Equalf(t, http.StatusBadRequest, rec.Code, "%s", query)
			require.Zerof(t, service.lookups, "%s reached the service", query)
		}
	}
}

func TestRoutes_HopCountOutsideItsBoundsIsRejected(t *testing.T) {
	for _, query := range []string{
		"?from=xpla1abc&hopCount=-1",
		fmt.Sprintf("?from=xpla1abc&hopCount=%d", maxHopCount+1),
		"?from=xpla1abc&hopCount=notanumber",
	} {
		rec, service := serveRoutes(t, query)

		require.Equalf(t, http.StatusBadRequest, rec.Code, "%s", query)
		require.Zerof(t, service.lookups, "%s reached the service", query)
	}
}

// The checks above must not stand between a caller and the answer it came for.
func TestRoutes_AnAddressIsStillServed(t *testing.T) {
	for _, query := range []string{
		"?from=xpla1abc&hopCount=3",
		"?to=ibc/ABCDEF0123456789&hopCount=" + fmt.Sprint(maxHopCount),
		"?from=xerc20:1234abcd&to=axpla&hopCount=2",
		// hopCount is optional, and its absence is not a malformed one.
		"?from=xpla1abc",
	} {
		rec, service := serveRoutes(t, query)

		require.Equalf(t, http.StatusOK, rec.Code, "%s", query)
		require.Equalf(t, 1, service.lookups, "%s", query)
	}
}
