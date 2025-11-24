package cache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryCacheInitializesStore(t *testing.T) {
	t.Parallel()

	cache := NewInMemoryCache()

	require.NotNil(t, cache, "NewInMemoryCache returned nil")
	require.NotNil(t, cache.items, "items map should be initialized")
	assert.Equal(t, 0, len(cache.items), "cache should start empty")
}

func TestSetAndGet(t *testing.T) {
	t.Parallel()

	cache := NewInMemoryCache()
	expected := struct{ Name string }{Name: "value"}

	cache.Set("key", expected)

	val, ok := cache.Get("key")
	require.True(t, ok, "expected key to be found")
	assert.Equal(t, expected, val, "unexpected value for key")
}

func TestGetMissingKey(t *testing.T) {
	t.Parallel()

	cache := NewInMemoryCache()

	val, ok := cache.Get("missing")
	require.False(t, ok, "missing key should not be found")
	assert.Nil(t, val, "missing key should return nil value")
}

func TestDeleteRemovesKey(t *testing.T) {
	t.Parallel()

	cache := NewInMemoryCache()

	cache.Set("temp", "value")
	cache.Delete("temp")

	_, ok := cache.Get("temp")
	require.False(t, ok, "deleted key should not be found")
	assert.Equal(t, 0, len(cache.items), "cache should be empty after delete")
}

func TestConcurrentAccess(t *testing.T) {
	cache := NewInMemoryCache()
	const workers = 50

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", i)
			cache.Set(key, i)

			// Mix reads with writes to exercise locks.
			cache.Get(fmt.Sprintf("key-%d", (i+1)%workers))

			if i%2 == 0 {
				cache.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < workers; i++ {
		key := fmt.Sprintf("key-%d", i)
		val, ok := cache.Get(key)
		if i%2 == 0 {
			assert.False(t, ok, "expected key %s to be deleted", key)
			continue
		}

		require.True(t, ok, "expected key %s to be present", key)
		assert.Equal(t, i, val, "unexpected value for %s", key)
	}
}
