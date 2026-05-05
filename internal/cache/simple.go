package cache

import (
	"sync"
	"time"
)

// SimpleCache is a basic thread-safe map cache.
type SimpleCache struct {
	items map[string]*Item
	mu    sync.RWMutex
}

// NewSimpleCache creates a new SimpleCache instance.
func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		items: make(map[string]*Item),
	}
}

// Set adds or updates an item in the cache.
func (c *SimpleCache) Set(key, value string, ttlSeconds int) error {
	var expiresAt time.Time
	if ttlSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &Item{
		Key:       key,
		Value:     []byte(value),
		ExpiresAt: expiresAt,
	}
	
	return nil
}

// Get retrieves an item from the cache.
func (c *SimpleCache) Get(key string) (string, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return "", false
	}

	if item.IsExpired() {
		c.Delete(key)
		return "", false
	}

	return string(item.Value), true
}

// Delete removes an item from the cache.
func (c *SimpleCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Stats returns internal statistics about the cache.
func (c *SimpleCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return map[string]interface{}{
		"item_count": len(c.items),
	}
}

// Keys returns all keys currently in the cache.
func (c *SimpleCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}
