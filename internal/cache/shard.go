package cache

import (
	"sync"
	"time"
)

// Shard handles a segment of the overall cache, managing its own locks and LRU list.
type Shard struct {
	items    map[string]*Item
	nodes    map[string]*LRUNode
	lru      *LRUList
	mu       sync.RWMutex
	maxItems int
}

// NewShard creates a new cache shard.
func NewShard(maxItems int) *Shard {
	return &Shard{
		items:    make(map[string]*Item),
		nodes:    make(map[string]*LRUNode),
		lru:      NewLRUList(),
		maxItems: maxItems,
	}
}

// Set adds or updates an item in the shard.
func (s *Shard) Set(key string, item *Item) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If key already exists, update item and move to front
	if node, exists := s.nodes[key]; exists {
		s.items[key] = item
		s.lru.MoveToFront(node)
		return
	}

	// Add new item
	s.items[key] = item
	s.nodes[key] = s.lru.PushFront(key)

	// Evict if over capacity
	if s.maxItems > 0 && len(s.items) > s.maxItems {
		tail := s.lru.RemoveTail()
		if tail != nil {
			delete(s.items, tail.Key)
			delete(s.nodes, tail.Key)
		}
	}
}

// Get retrieves an item from the shard.
func (s *Shard) Get(key string) (*Item, bool) {
	s.mu.Lock() // Requires full lock to update LRU list
	defer s.mu.Unlock()

	item, exists := s.items[key]
	if !exists {
		return nil, false
	}

	// Lazy deletion
	if item.IsExpired() {
		s.deleteLocked(key)
		return nil, false
	}

	node := s.nodes[key]
	s.lru.MoveToFront(node)
	
	return item, true
}

// Delete removes an item from the shard.
func (s *Shard) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteLocked(key)
}

func (s *Shard) deleteLocked(key string) {
	if node, exists := s.nodes[key]; exists {
		s.lru.Remove(node)
		delete(s.nodes, key)
		delete(s.items, key)
	}
}

// cleanup removes expired items from the shard.
func (s *Shard) cleanup(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, item := range s.items {
		if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
			s.deleteLocked(key)
		}
	}
}

// getKeys returns all keys currently in the shard.
func (s *Shard) getKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.items))
	for key := range s.items {
		keys = append(keys, key)
	}
	return keys
}

// getAllItems returns all items in the shard.
func (s *Shard) getAllItems() map[string]Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make(map[string]Item, len(s.items))
	now := time.Now()
	for key, item := range s.items {
		// Only return non-expired items
		if item.ExpiresAt.IsZero() || now.Before(item.ExpiresAt) {
			items[key] = *item
		}
	}
	return items
}

// count returns the number of items in the shard.
func (s *Shard) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}
