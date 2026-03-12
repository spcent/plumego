package mq

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type RetryPolicy interface {
	NextDelay(task Task, err error) time.Duration
}

type ExponentialBackoff struct {
	Base   time.Duration
	Max    time.Duration
	Factor float64
	Jitter float64
}

func (b ExponentialBackoff) NextDelay(task Task, _ error) time.Duration {
	base := b.Base
	if base <= 0 {
		base = time.Second
	}
	max := b.Max
	if max <= 0 {
		max = 5 * time.Minute
	}
	factor := b.Factor
	if factor <= 1 {
		factor = 2
	}
	attempts := task.Attempts
	if attempts < 1 {
		attempts = 1
	}
	exp := math.Pow(factor, float64(attempts-1))
	delay := time.Duration(float64(base) * exp)
	if delay > max {
		delay = max
	}
	if b.Jitter > 0 {
		jitter := b.Jitter
		if jitter > 1 {
			jitter = 1
		}
		r := randFloat64()
		mult := 1 - jitter + (2 * jitter * r)
		delay = time.Duration(float64(delay) * mult)
		if delay < 0 {
			delay = base
		}
	}
	return delay
}

func DefaultRetryPolicy() RetryPolicy {
	return ExponentialBackoff{
		Base:   time.Second,
		Max:    5 * time.Minute,
		Factor: 2,
		Jitter: 0.2,
	}
}

var (
	rngMu sync.Mutex
	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func randFloat64() float64 {
	rngMu.Lock()
	defer rngMu.Unlock()
	return rng.Float64()
}
