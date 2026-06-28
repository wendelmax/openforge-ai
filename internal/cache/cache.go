// Package cache provides a simple in-memory cache with TTL support.
package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// Entry holds a cached value with its creation time and TTL.
type Entry struct {
	Value     []byte
	CreatedAt time.Time
	TTL       time.Duration
}

// Cache is a thread-safe in-memory cache with TTL-based expiration.
type Cache struct {
	mu    sync.RWMutex
	store map[string]*Entry
}

// New creates a new empty Cache.
func New() *Cache {
	return &Cache{
		store: make(map[string]*Entry),
	}
}

func (c *Cache) hashKey(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Get retrieves a value by key, returning false if missing or expired.
func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.store[key]
	if !ok {
		return nil, false
	}

	if entry.TTL > 0 && time.Since(entry.CreatedAt) > entry.TTL {
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value with the given key and TTL.
func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = &Entry{
		Value:     value,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}
}

// Delete removes a value by key from the cache.
func (c *Cache) Delete(ctx context.Context, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, key)
}

// Clear removes all entries from the cache.
func (c *Cache) Clear(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]*Entry)
}
