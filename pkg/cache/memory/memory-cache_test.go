package memory

import (
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
		{setUpItem{"test", "value", cache.CacheLifeTimeNeverExpired}, time.Second, "garbageValue", "value"},
		{setUpItem{"test", "value", time.Second * 2}, time.Second, "garbageValue", "value"},
		{setUpItem{"test", "value", time.Second}, time.Second * 2, "", ""},
		{setUpItem{"test", "value", time.Second}, time.Second * 2, "", ""},
	}

	assert := assert.New(t)

	tcTester := func(id int, t testCase) {
		c := NewMemoryCache(cache.NewByteCodec())
		setUp(t.item, c)
		fmt.Printf("test case %d\n", id)
		time.Sleep(t.wait)

		err := c.Get(t.item.key, &t.destination)
		if t.expected == "" {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
		assert.Equal(t.expected, t.destination)
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
