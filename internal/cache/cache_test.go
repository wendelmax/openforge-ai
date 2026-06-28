package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New()
	ctx := context.Background()

	c.Set(ctx, "key1", []byte("value1"), 0)
	val, ok := c.Get(ctx, "key1")
	assert.True(t, ok)
	assert.Equal(t, []byte("value1"), val)
}

func TestCache_Miss(t *testing.T) {
	c := New()
	ctx := context.Background()

	_, ok := c.Get(ctx, "nonexistent")
	assert.False(t, ok)
}

func TestCache_TTLExpiry(t *testing.T) {
	c := New()
	ctx := context.Background()

	c.Set(ctx, "key1", []byte("value1"), 50*time.Millisecond)
	val, ok := c.Get(ctx, "key1")
	assert.True(t, ok)
	assert.Equal(t, []byte("value1"), val)

	time.Sleep(100 * time.Millisecond)
	_, ok = c.Get(ctx, "key1")
	assert.False(t, ok)
}

func TestCache_Delete(t *testing.T) {
	c := New()
	ctx := context.Background()

	c.Set(ctx, "key1", []byte("value1"), 0)
	c.Delete(ctx, "key1")
	_, ok := c.Get(ctx, "key1")
	assert.False(t, ok)
}

func TestCache_Clear(t *testing.T) {
	c := New()
	ctx := context.Background()

	c.Set(ctx, "key1", []byte("value1"), 0)
	c.Set(ctx, "key2", []byte("value2"), 0)
	c.Clear(ctx)

	_, ok := c.Get(ctx, "key1")
	assert.False(t, ok)
	_, ok = c.Get(ctx, "key2")
	assert.False(t, ok)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New()
	ctx := context.Background()

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(i int) {
			key := string(rune('a' + i%26))
			c.Set(ctx, key, []byte{byte(i)}, 0)
			c.Get(ctx, key)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestCache_HashKey(t *testing.T) {
	c := New()
	h1 := c.hashKey("hello", "world")
	h2 := c.hashKey("hello", "world")
	h3 := c.hashKey("hello", "there")

	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
}
