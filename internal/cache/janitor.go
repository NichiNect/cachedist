package cache

import (
	"context"
	"time"
)

// Janitor is responsible for periodically cleaning up expired items across all shards.
type Janitor struct{}

// NewJanitor creates a new Janitor.
func NewJanitor() *Janitor {
	return &Janitor{}
}

// Run starts the background cleanup goroutine.
func (j *Janitor) Run(ctx context.Context, interval time.Duration, shards []*Shard) {
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				for _, shard := range shards {
					shard.cleanup(now)
				}
			}
		}
	}()
}
