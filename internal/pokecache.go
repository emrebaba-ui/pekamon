package internal

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	Val []byte
}

type Cache struct {
	data map[string]cacheEntry
	mu sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		data: make(map[string]cacheEntry),
	}
	go c.reapLoop(interval)
	return c
}

func (c *Cache) Add(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	newEntry := cacheEntry{createdAt: time.Now(), Val: value}
	c.data[key] = newEntry
}

func (c *Cache) Get(key string) (cacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.data[key]
	return entry, ok
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		defer c.mu.Unlock()
		for key, entry := range c.data {
			if time.Since(entry.createdAt) > interval {
				delete(c.data, key)
			}
			}
		}
	}

