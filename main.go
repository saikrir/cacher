package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/saikrir/cache/cache"
)

func main() {
	fmt.Println("Hello World")
	keyFmt := "key-%02d"
	keyName := func(i int) string { return fmt.Sprintf(keyFmt, i) }

	size := 5
	ttl := 10 * time.Millisecond
	log.Printf("info: creating cache: size=%d, ttl=%v", size, ttl)
	c, err := cache.New(size, ttl)
	if err != nil {
		log.Printf("error: can't create - %s", err)
		return
	}
	log.Printf("info: OK")

	log.Printf("info: checking TTL")
	key, val := keyName(1), 3
	c.Set(key, val)
	v, ok := c.Get(key)
	if !ok || v != val {
		log.Printf("error: %q: got %v (ok=%v)", key, v, ok)
		return
	}

	// Let key expire
	time.Sleep(2 * ttl)
	_, ok = c.Get(key)
	if ok {
		log.Printf("error: %q: got value after TTL", key)
		return
	}
	log.Printf("info: OK")

	log.Printf("info: checking overflow")
	n := size * 2
	for i := 0; i < n; i++ {
		c.Set(keyName(i), i)
	}
	_, ok = c.Get(keyName(1))
	if ok {
		log.Printf("error: %q: got value after overflow", key)
		return
	}
	_, ok = c.Get(keyName(n - 1))
	if !ok {
		log.Printf("error: %q: not found", key)
		return
	}
	log.Printf("info: OK")

	numGr := size * 3
	count := 1000
	log.Printf("info: checking concurrency (%d goroutines, %d loops each)", numGr, count)

	var wg sync.WaitGroup
	wg.Add(numGr)
	for i := 0; i < numGr; i++ {
		key := keyName(i)
		go func() {
			defer wg.Done()
			for i := 0; i < count; i++ {
				time.Sleep(time.Microsecond)
				c.Set(key, i)
			}
		}()
	}
	wg.Wait()
	log.Printf("info: OK")
}
