package cache

import (
	"testing"
	"time"
)

func TestSimpleCache_SetAndGet(t *testing.T) {
	t.Parallel()
	
	c := NewSimpleCache()
	c.Set("foo", "bar", 0)
	
	val, found := c.Get("foo")
	if !found {
		t.Errorf("Expected to find key 'foo'")
	}
	if val != "bar" {
		t.Errorf("Expected value 'bar', got '%s'", val)
	}
}

func TestSimpleCache_Delete(t *testing.T) {
	t.Parallel()
	
	c := NewSimpleCache()
	c.Set("foo", "bar", 0)
	c.Delete("foo")
	
	_, found := c.Get("foo")
	if found {
		t.Errorf("Expected key 'foo' to be deleted")
	}
}

func TestSimpleCache_TTLExpires(t *testing.T) {
	t.Parallel()
	
	c := NewSimpleCache()
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

func BenchmarkSimpleCache_Set(b *testing.B) {
	c := NewSimpleCache()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set("key", "value", 0)
	}
}

func BenchmarkSimpleCache_Get(b *testing.B) {
	c := NewSimpleCache()
	c.Set("key", "value", 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get("key")
	}
}
