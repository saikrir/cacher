package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type entry struct {
	value any
	time  time.Time
}

type Cache struct {
	store  map[string]entry
	size   int
	ttl    time.Duration
	mu     sync.RWMutex
	cancel context.CancelFunc
}

func New(size int, ttl time.Duration) (*Cache, error) {

	if size <= 0 {
		return nil, fmt.Errorf("size should be greater than zero")
	}

	if ttl <= 0 {
		return nil, fmt.Errorf("ttl should be greater than zero")
	}

	ctx, cancel := context.WithCancel(context.Background())

	storage := make(map[string]entry)

	cache := &Cache{store: storage, size: size, ttl: ttl, cancel: cancel}
	go cache.ttlEnforcer(ctx)
	return cache, nil
}

func (c *Cache) Close() {
	c.cancel()
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLocker().Lock()
	defer c.mu.RLocker().Unlock()

	if value, ok := c.store[key]; ok {
		return value.value, true
	}
	return nil, false
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.store) == c.size { // if we are at capacity, evict one
		c.evictLRU()
	}
	newEntry := entry{
		value: value,
		time:  time.Now(),
	}
	c.store[key] = newEntry
}

func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RLocker().Unlock()

	keys := make([]string, len(c.store))
	i := 0
	for k := range c.store {
		keys[i] = k
		i++
	}
	return keys
}

func (c *Cache) evictLRU() {
	minTime, key := time.Now(), ""
	for k, v := range c.store {
		if v.time.Before(minTime) {
			minTime, key = v.time, k
		}
	}
	delete(c.store, key)
}

func (c *Cache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Println("Eviction Timer will run")
	now := time.Now()
	for k, v := range c.store {
		if now.Sub(v.time) > c.ttl {
			delete(c.store, k)
		}
	}
}

func (c *Cache) ttlEnforcer(ctx context.Context) {
	timer := time.NewTimer(c.ttl)
	for {
		select {
		case <-timer.C:
			c.evictExpired()
		case <-ctx.Done():
			return
		}
	}
}
