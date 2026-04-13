package coingecko

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSingleflightDeduplicatesCoinGeckoRequests verifies that concurrent calls
// to cachePriceInUsd via sfGroup.Do result in exactly one outbound HTTP request.
func TestSingleflightDeduplicatesCoinGeckoRequests(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		time.Sleep(20 * time.Millisecond) // let goroutines pile up before responding
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"prices":[[1000000,1.0],[2000000,2.0]]}`)
	}))
	defer srv.Close()

	s := &tickerService{httpClient: srv.Client(), endpoint: srv.URL + "/"}

	const concurrency = 20
	var wg sync.WaitGroup
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.sfGroup.Do(priceTokenId, func() (any, error) { //nolint:errcheck
				return nil, s.cachePriceInUsd(priceTokenId)
			})
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), callCount.Load(), "singleflight should collapse concurrent fetches into one HTTP request")
	assert.Equal(t, 2.0, s.price(3_000_000, true), "prices should be cached after fetch")
}

// TestSingleflightSecondBatchFetchesAgain verifies that after the in-flight
// call completes, a subsequent call (new batch) does issue a new HTTP request.
func TestSingleflightSecondBatchFetchesAgain(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"prices":[[1000000,1.0],[2000000,2.0]]}`)
	}))
	defer srv.Close()

	s := &tickerService{httpClient: srv.Client(), endpoint: srv.URL + "/"}

	// first call
	_, err, _ := s.sfGroup.Do(priceTokenId, func() (any, error) {
		return nil, s.cachePriceInUsd(priceTokenId)
	})
	assert.NoError(t, err)

	// second call — previous Do has already finished, so a new request goes out
	_, err, _ = s.sfGroup.Do(priceTokenId, func() (any, error) {
		return nil, s.cachePriceInUsd(priceTokenId)
	})
	assert.NoError(t, err)

	assert.Equal(t, int32(2), callCount.Load(), "each independent call should fetch once")
}
