package redis

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dezswap/dezswap-api/pkg/cache"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func Test_cache(t *testing.T) {
	type item struct {
		Id   int
		Name string
	}

	type setUpItem[T any] struct {
		key   string
		value T
		ttl   time.Duration
	}
	type testCase[T any] struct {
		item        setUpItem[T]
		wait        time.Duration
		destination T
		expected    *T
	}
	setUp := func(i setUpItem[item], c cache.Cache, codec cache.Codable, mock redismock.ClientMock, isTimeOuted bool) {
		encoded, err := codec.Encode(i.value)
		if err != nil {
			panic(err)
		}
		mock.ExpectSet(i.key, encoded, i.ttl).SetVal(string(encoded))
		c.Set(i.key, i.value, i.ttl)
		if isTimeOuted {
			mock.ExpectGet(i.key).RedisNil()
		} else {
			mock.ExpectGet(i.key).SetVal(string(encoded))
		}
	}

	tcs := []testCase[item]{
		// find key
		{setUpItem[item]{"test", item{1, "testItem"}, cache.CacheLifeTimeNeverExpired}, 0, item{}, &item{1, "testItem"}},
		{setUpItem[item]{"test", item{1, "testItem"}, cache.CacheLifeTimeNeverExpired}, time.Second, item{}, &item{1, "testItem"}},
		{setUpItem[item]{"test", item{1, "testItem"}, time.Second * 2}, time.Second, item{}, &item{1, "testItem"}},
		{setUpItem[item]{"test", item{1, "testItem"}, time.Second}, time.Second * 2, item{}, nil},
		{setUpItem[item]{"test", item{1, "testItem"}, time.Second}, time.Second * 2, item{}, nil},
	}

	assert := assert.New(t)

	tcTester := func(id int, t testCase[item]) {
		codec := cache.NewByteCodec()
		db, mock := redismock.NewClientMock()
		c := New(codec, db)
		setUp(t.item, c, codec, mock, t.expected == nil)
		fmt.Printf("test case %d\n", id)
		time.Sleep(t.wait)

		err := c.Get(t.item.key, &t.destination)
		if t.expected == nil {
			assert.Error(err)
		} else {
			assert.NoError(err)
			assert.Equal(*t.expected, t.destination)
		}
	}

	wg := sync.WaitGroup{}
	for id, tc := range tcs {
		wg.Add(1)
		go func(id int, tc testCase[item]) {
			defer wg.Done()
			tcTester(id, tc)
		}(id, tc)
	}
	wg.Wait()
}
