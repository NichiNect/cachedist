package cache

import "time"

// Item represents a single cached item.
type Item struct {
	Key       string
	Value     []byte
	ExpiresAt time.Time
}

// IsExpired checks if the item has expired.
func (i *Item) IsExpired() bool {
	if i.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(i.ExpiresAt)
}
