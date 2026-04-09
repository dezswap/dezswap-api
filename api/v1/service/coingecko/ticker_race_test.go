package coingecko

import (
	"sync"
	"testing"
)

// TestCachedPricesNoRace verifies that concurrent reads and writes to
// cachedPrices do not trigger the race detector.
func TestCachedPricesNoRace(t *testing.T) {
	s := &tickerService{}

	sample := [][priceInfoLength]float64{
		{1_000_000, 1.0},
		{2_000_000, 2.0},
	}

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)

		// writer: simulates cachePriceInUsd replacing the slice
		go func() {
			defer wg.Done()
			s.mu.Lock()
			s.cachedPrices = sample
			s.mu.Unlock()
		}()

		// reader: simulates price() iterating the slice
		go func() {
			defer wg.Done()
			_ = s.price(0, true)
		}()
	}
	wg.Wait()
}
