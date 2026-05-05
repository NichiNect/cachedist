package cache

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
}
