package router

import (
	"fmt"
	"sync"
	"testing"
)

func TestRouteCacheConcurrentAccess(t *testing.T) {
	cache := NewRouteCache(64)
	for i := 0; i < 64; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), &MatchResult{})
	}

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("key-%d", (id+j)%128)
				cache.Get(key)
			}
		}(i)
	}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				key := fmt.Sprintf("key-%d", (id*500+j)%128)
				cache.Set(key, &MatchResult{})
			}
		}(i)
	}

	wg.Wait()

	if size := cache.Size(); size > 64 {
		t.Fatalf("cache size exceeded capacity: %d", size)
	}
}
