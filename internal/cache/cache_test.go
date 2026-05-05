package cache

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestShardedCache_SetAndGet(t *testing.T) {
	t.Parallel()
	
	c := NewShardedCache(16, 100, 30)
	defer c.Stop()

	c.Set("foo", "bar", 0)
	
	val, found := c.Get("foo")
	if !found {
		t.Errorf("Expected to find key 'foo'")
	}
	if val != "bar" {
		t.Errorf("Expected value 'bar', got '%s'", val)
	}
}

func TestShardedCache_Delete(t *testing.T) {
	t.Parallel()
	
	c := NewShardedCache(16, 100, 30)
	defer c.Stop()

	c.Set("foo", "bar", 0)
	c.Delete("foo")
	
	_, found := c.Get("foo")
	if found {
		t.Errorf("Expected key 'foo' to be deleted")
	}
}

func TestTTL_ExpiredKeyNotReturned(t *testing.T) {
	t.Parallel()
	
	c := NewShardedCache(16, 100, 1) // 1 sec TTL cleanup
	defer c.Stop()

	c.Set("foo", "bar", 1) // 1 second TTL
	
	_, found := c.Get("foo")
	if !found {
		t.Errorf("Expected to find key 'foo' immediately")
	}
	
	time.Sleep(1100 * time.Millisecond) // wait for expiry
	
	_, found = c.Get("foo")
	if found {
		t.Errorf("Expected key 'foo' to be expired")
	}
}

func TestLRU_EvictsLeastRecent(t *testing.T) {
	t.Parallel()

	// small cache to test eviction (16 max items total, 4 shards -> 4 items per shard)
	c := NewShardedCache(4, 16, 30)
	defer c.Stop()

	// Keys designed to hash to the same shard would be ideal, but let's just 
	// insert enough items to force evictions across all shards.
	for i := 0; i < 50; i++ {
		c.Set(fmt.Sprintf("key-%d", i), "value", 0)
	}

	stats := c.Stats()
	if count, ok := stats["item_count"].(int); ok {
		if count > 16 {
			t.Errorf("Expected max 16 items, got %d", count)
		}
	} else {
		t.Errorf("Failed to read item_count from stats")
	}
}

func TestShardedCache_ParallelReadWrite(t *testing.T) {
	t.Parallel()

	c := NewShardedCache(32, 10000, 30)
	defer c.Stop()

	var wg sync.WaitGroup
	workers := 100
	opsPerWorker := 100

	wg.Add(workers * 2)

	// Writers
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				key := "key-" + strconv.Itoa(workerID) + "-" + strconv.Itoa(j)
				c.Set(key, "value", 0)
			}
		}(i)
	}

	// Readers
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				key := "key-" + strconv.Itoa(workerID) + "-" + strconv.Itoa(j)
				c.Get(key) // Might not be written yet, that's fine
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkCache_Set(b *testing.B) {
	c := NewShardedCache(256, 1000000, 30)
	defer c.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key-" + strconv.Itoa(i)
			c.Set(key, "value", 0)
			i++
		}
	})
}

func BenchmarkCache_Get(b *testing.B) {
	c := NewShardedCache(256, 1000000, 30)
	defer c.Stop()

	// Pre-fill some keys
	for i := 0; i < 1000; i++ {
		c.Set("key-"+strconv.Itoa(i), "value", 0)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key-" + strconv.Itoa(i%1000)
			c.Get(key)
			i++
		}
	})
}
