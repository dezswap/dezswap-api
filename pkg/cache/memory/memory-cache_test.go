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
		value interface{}
		ttl   time.Duration
	}
	type testCase struct {
		item     setUpItem
		wait     time.Duration
		expected interface{}
	}
	setUp := func(i setUpItem, c cache.Cache) {
		c.Set(i.key, i.value, i.ttl)
	}

	tcs := []testCase{
		// find key
		{setUpItem{"test", "value", cache.CacheLifeTimeNeverExpired}, 0, "value"},
		{setUpItem{"test", "value", cache.CacheLifeTimeNeverExpired}, time.Second, "value"},
		{setUpItem{"test", "value", time.Second * 2}, time.Second, "value"},
		{setUpItem{"test", "value", time.Second}, time.Second * 2, nil},
		{setUpItem{"test", "value", time.Second}, time.Second * 2, nil},
	}

	assert := assert.New(t)

	tcTester := func(id int, t testCase) {
		c := NewMemoryCache()
		setUp(t.item, c)
		fmt.Printf("test case %d\n", id)
		time.Sleep(t.wait)
		actual, _ := c.Get(t.item.key)
		assert.Equal(t.expected, actual)
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

	var testStruct = struct {
		id   int
		name string
	}{1, "test"}

	// test reference
	c := NewMemoryCache()
	c.Set("test", &testStruct, cache.CacheLifeTimeNeverExpired)
	testStruct.id = 2
	testStruct.name = "test2"
	actual, _ := c.Get("test")
	assert.Equal(&testStruct, actual)

}
