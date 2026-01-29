package core

import (
	"sync"
	"testing"
)

type RaceDep struct{}

type RaceTarget struct {
	Dep *RaceDep `inject:""`
}

func TestInjectConcurrency(t *testing.T) {
	c := NewDIContainer()
	c.Register(&RaceDep{})

	target := &RaceTarget{}

	// Pre-warm the container to avoid resolution race masking the injection race
	if err := c.Inject(target); err != nil {
		t.Fatalf("setup inject failed: %v", err)
	}

	var wg sync.WaitGroup
	count := 100
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			// Concurrent Inject on the SAME instance
			if err := c.Inject(target); err != nil {
				// checking error is not enough, race detector should catch data race
				_ = err
			}
		}()
	}
	wg.Wait()
}
