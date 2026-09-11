package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func Test_cache(t *testing.T) {
	type setUpItem struct {
		key   string
		value string
		ttl   time.Duration
	}
	type testCase struct {
		item        setUpItem
		wait        time.Duration
		destination string
		expected    string
	}
	setUp := func(i setUpItem, c cache.Cache) {
		c.Set(i.key, i.value, i.ttl)
	}

	tcs := []testCase{
		// find key
		{setUpItem{"test", "value", cache.CacheLifeTimeNeverExpired}, 0, "garbageValue", "value"},
		{setUpItem{"test", "value", cache.CacheLifeTimeNeverExpired}, time.Millisecond * 50, "garbageValue", "value"},
		{setUpItem{"test", "value", time.Millisecond * 100}, time.Millisecond * 50, "garbageValue", "value"},
		{setUpItem{"test", "value", time.Millisecond * 50}, time.Millisecond * 150, "", ""},
		{setUpItem{"test", "value", time.Millisecond * 50}, time.Millisecond * 150, "", ""},
	}

	tcTester := func(id int, tc testCase) {
		c := NewMemoryCache(context.Background(), cache.NewByteCodec())
		setUp(tc.item, c)
		fmt.Printf("test case %d\n", id)
		time.Sleep(tc.wait)

		err := c.Get(tc.item.key, &tc.destination)
		if tc.expected == "" {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, tc.expected, tc.destination)
	}

	wg := sync.WaitGroup{}

	for id, tc := range tcs {
		wg.Add(1)
		go func(id int, tc testCase) {
			defer wg.Done()
			tcTester(id, tc)
		}(id, tc)
	}
	wg.Wait()
}

func Test_deleteExpired(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := NewMemoryCache(ctx, cache.NewByteCodec()).(*memoryCacheImpl)

	// Set one item that expires soon and one that does not
	c.Set("expires", "value", time.Millisecond*10)
	c.Set("permanent", "value", cache.CacheLifeTimeNeverExpired)

	// Wait for the item to expire
	time.Sleep(time.Millisecond * 20)

	// Manually trigger cleanup
	c.deleteExpired()

	c.RLock()
	_, hasExpired := c.store["expires"]
	_, hasPermanent := c.store["permanent"]
	c.RUnlock()

	assert.False(hasExpired, "expired key should have been removed")
	assert.True(hasPermanent, "permanent key should remain")
}

// Keys are built from the request, so a caller varying one writes an entry per
// request. Bounded only by expiry, those grow the process until it is killed well
// inside a single TTL.
func Test_maxEntriesBoundsTheStore(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const maxEntries = 8
	c := NewMemoryCacheWithLimit(ctx, cache.NewByteCodec(), maxEntries).(*memoryCacheImpl)

	for i := range maxEntries * 4 {
		assert.NoError(c.Set(fmt.Sprintf("/v1/pools?cacheBuster=%d", i), "value", time.Minute))
	}

	c.RLock()
	defer c.RUnlock()

	assert.Equal(maxEntries, len(c.store))
	// The list is what eviction walks, so an entry left in it after its key was
	// dropped would be evicted forever without freeing anything.
	assert.Equal(maxEntries, c.order.Len())
}

// Eviction takes the oldest write, so a flood costs the entries written before it
// and not the ones it wrote.
func Test_evictionTakesTheOldestWrite(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := NewMemoryCacheWithLimit(ctx, cache.NewByteCodec(), 2)
	assert.NoError(c.Set("first", "value", time.Minute))
	assert.NoError(c.Set("second", "value", time.Minute))
	assert.NoError(c.Set("third", "value", time.Minute))

	var value string
	assert.ErrorIs(c.Get("first", &value), cache.ErrCacheMiss)
	assert.NoError(c.Get("second", &value))
	assert.NoError(c.Get("third", &value))
}

// Rewriting a key is what a route does on every miss, and must not count against
// the bound as a new entry would.
func Test_rewritingAKeyDoesNotGrowTheStore(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := NewMemoryCacheWithLimit(ctx, cache.NewByteCodec(), 4).(*memoryCacheImpl)
	for range 20 {
		assert.NoError(c.Set("/v1/pools", "value", time.Minute))
	}

	var value string
	assert.NoError(c.Get("/v1/pools", &value))
	assert.Equal("value", value)

	c.RLock()
	defer c.RUnlock()

	assert.Equal(1, len(c.store))
	assert.Equal(1, c.order.Len())
}
