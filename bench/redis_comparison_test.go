package bench

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/NichiNect/cachedist/internal/cache"
	"github.com/redis/go-redis/v9"
)

func BenchmarkComparison(b *testing.B) {
	ctx := context.Background()

	// Initialize cachedist
	cd := cache.NewShardedCache(256, 1000000, 60)
	defer cd.Stop()

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	// Check if redis is available
	if err := rdb.Ping(ctx).Err(); err != nil {
		b.Log("Redis not available, skipping redis benchmarks. Run: docker run -p 6379:6379 redis:alpine")
	}

	b.Run("cachedist-SET", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cd.Set(fmt.Sprintf("key_%d", i), "value", 0)
		}
	})

	b.Run("cachedist-GET", func(b *testing.B) {
		cd.Set("key", "value", 0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cd.Get("key")
		}
	})

	if rdb.Ping(ctx).Err() == nil {
		b.Run("redis-SET", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rdb.Set(ctx, fmt.Sprintf("key_%d", i), "value", 0)
			}
		})

		b.Run("redis-GET", func(b *testing.B) {
			rdb.Set(ctx, "key", "value", 0)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rdb.Get(ctx, "key")
			}
		})
	}
}

func BenchmarkLatency(b *testing.B) {
	ctx := context.Background()
	cd := cache.NewShardedCache(256, 1000000, 60)
	defer cd.Stop()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	b.Run("cachedist-LATENCY", func(b *testing.B) {
		durations := make([]time.Duration, b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			cd.Set("key", "value", 0)
			durations[i] = time.Since(start)
		}
		// In a real scenario we would aggregate these for p50/p95/p99
	})

	if rdb.Ping(ctx).Err() == nil {
		b.Run("redis-LATENCY", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rdb.Set(ctx, "key", "value", 0)
			}
		})
	}
}
