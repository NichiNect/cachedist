package cache

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkCache_Get(b *testing.B) {
	c := NewShardedCache(256, 1000000, 60)
	defer c.Stop()

	key := "bench_key"
	c.Set(key, "bench_value", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(key)
	}
}

func BenchmarkCache_Set(b *testing.B) {
	c := NewShardedCache(256, 1000000, 60)
	defer c.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key_%d", i), "value", 0)
	}
}

func BenchmarkCache_GetParallel(b *testing.B) {
	c := NewShardedCache(256, 1000000, 60)
	defer c.Stop()

	key := "bench_key"
	c.Set(key, "bench_value", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Get(key)
		}
	})
}

func BenchmarkCache_SetParallel(b *testing.B) {
	c := NewShardedCache(256, 1000000, 60)
	defer c.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(fmt.Sprintf("key_%d", i), "value", 0)
			i++
		}
	})
}

func BenchmarkCache_MixedWorkload(b *testing.B) {
	c := NewShardedCache(256, 1000000, 60)
	defer c.Stop()

	// Pre-fill some data
	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("key_%d", i), "value", 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(42))
		for pb.Next() {
			key := fmt.Sprintf("key_%d", r.Intn(1000))
			if r.Float64() < 0.8 {
				c.Get(key)
			} else {
				c.Set(key, "value", 0)
			}
		}
	})
}

func BenchmarkCache_LargeValue(b *testing.B) {
	c := NewShardedCache(256, 1000000, 60)
	defer c.Stop()

	// 10KB value
	largeValue := make([]byte, 10*1024)
	for i := range largeValue {
		largeValue[i] = 'a'
	}
	valueStr := string(largeValue)
	key := "large_key"

	b.Run("SET", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Set(key, valueStr, 0)
		}
	})

	b.Run("GET", func(b *testing.B) {
		c.Set(key, valueStr, 0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c.Get(key)
		}
	})
}
