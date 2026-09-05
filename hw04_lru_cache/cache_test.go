package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(10)

		for i := 0; i < 10; i++ {
			key := Key(strconv.Itoa('a' + i))
			c.Set(key, i)
		}

		c.Clear()

		for i := 0; i < 10; i++ {
			key := Key(strconv.Itoa('a' + i))
			item, inCache := c.Get(key)

			require.False(t, inCache, "item in cache: %+v", item)
		}
	})

	t.Run("small cache", func(t *testing.T) {
		c := NewCache(2)

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		value, inCache := c.Get("a")
		require.Nil(t, value, "value must be nil")
		require.False(t, inCache, "key value in cache")
	})

	t.Run("clear old items", func(t *testing.T) {
		c := NewCache(3)

		c.Set("old", 1)
		c.Set("b", 2)
		c.Set("c", 3) // <- now c - b - old

		c.Set("old", 10) // <- now old - c - b
		c.Set("b", 20)   // <- now b - old - c
		c.Get("c")       // <- now c - b - old

		c.Set("new", 4)

		value, inCache := c.Get("old")
		require.Nil(t, value, "old value still in cache")
		require.False(t, inCache, "old value still in cache")
	})
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
