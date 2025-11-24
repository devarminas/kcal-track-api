package cache

import "sync"

type CacheStore interface {
	Set(key string, value any)
	Get(key string) (any, bool)
	Delete(key string)
}

type inMemoryCache struct {
	mu    sync.RWMutex
	items map[string]any
}

func NewInMemoryCache() *inMemoryCache {
	return &inMemoryCache{
		items: make(map[string]any),
	}
}

func (c *inMemoryCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

func (c *inMemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, found := c.items[key]
	return val, found
}

func (c *inMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}
