package cache

import (
	"context"
	"hash/fnv"
	"time"
)

// Cache is the interface that all cache implementations must satisfy.
type Cache interface {
	// Set adds or updates an item in the cache. ttlSeconds = 0 means no expiration.
	Set(key, value string, ttlSeconds int) error
	
	// Get retrieves an item from the cache. Returns the value and a boolean indicating if it was found.
	Get(key string) (string, bool)
	
	// Delete removes an item from the cache.
	Delete(key string)
	
	// Stats returns internal statistics about the cache.
	Stats() map[string]interface{}
	
	// Keys returns all keys currently in the cache.
	Keys() []string
	
	// GetAllItems returns all items with their values and expirations.
	GetAllItems() map[string]Item
}

// ShardedCache implements the Cache interface with multiple shards to reduce lock contention.
type ShardedCache struct {
	shards     []*Shard
	numShards  uint32
	janitor    *Janitor
	cancelFunc context.CancelFunc
}

// NewShardedCache creates a new ShardedCache instance.
func NewShardedCache(numShards int, maxItemsPerNode int, ttlCleanupInterval int) *ShardedCache {
	if numShards <= 0 {
		numShards = 256
	}
	
	maxItemsPerShard := maxItemsPerNode / numShards
	if maxItemsPerShard < 1 {
		maxItemsPerShard = 1
	}

	shards := make([]*Shard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = NewShard(maxItemsPerShard)
	}

	ctx, cancel := context.WithCancel(context.Background())
	janitor := NewJanitor()
	
	interval := time.Duration(ttlCleanupInterval) * time.Second
	janitor.Run(ctx, interval, shards)

	return &ShardedCache{
		shards:     shards,
		numShards:  uint32(numShards),
		janitor:    janitor,
		cancelFunc: cancel,
	}
}

// Stop halts the background janitor goroutine.
func (c *ShardedCache) Stop() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
}

// getShard calculates the FNV-1a hash of the key to determine which shard to use.
func (c *ShardedCache) getShard(key string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	shardIdx := h.Sum32() % c.numShards
	return c.shards[shardIdx]
}

// Set adds or updates an item in the appropriate shard.
func (c *ShardedCache) Set(key, value string, ttlSeconds int) error {
	var expiresAt time.Time
	if ttlSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}

	item := &Item{
		Key:       key,
		Value:     []byte(value),
		ExpiresAt: expiresAt,
	}

	shard := c.getShard(key)
	shard.Set(key, item)
	return nil
}

// Get retrieves an item from the appropriate shard.
func (c *ShardedCache) Get(key string) (string, bool) {
	shard := c.getShard(key)
	item, found := shard.Get(key)
	if !found {
		return "", false
	}
	return string(item.Value), true
}

// Delete removes an item from the appropriate shard.
func (c *ShardedCache) Delete(key string) {
	shard := c.getShard(key)
	shard.Delete(key)
}

// Stats aggregates statistics from all shards.
func (c *ShardedCache) Stats() map[string]interface{} {
	totalItems := 0
	for _, shard := range c.shards {
		totalItems += shard.count()
	}
	
	return map[string]interface{}{
		"item_count": totalItems,
		"shards":     c.numShards,
	}
}

// Keys aggregates all keys from all shards.
func (c *ShardedCache) Keys() []string {
	var allKeys []string
	for _, shard := range c.shards {
		allKeys = append(allKeys, shard.getKeys()...)
	}
	return allKeys
}

// GetAllItems aggregates all items from all shards.
func (c *ShardedCache) GetAllItems() map[string]Item {
	allItems := make(map[string]Item)
	for _, shard := range c.shards {
		for k, v := range shard.getAllItems() {
			allItems[k] = v
		}
	}
	return allItems
}
