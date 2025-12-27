package webhookout

import (
	"math/rand"
	"time"
)

type RetryDecision struct {
	Retry bool
	Next  time.Time
	Why   string
}

func ShouldRetry(status int, err error, attempt int, maxRetries int, retryOn429 bool) bool {
	if attempt >= maxRetries {
		return false
	}
	if err != nil {
		return true
	}
	if status >= 500 && status <= 599 {
		return true
	}
	if retryOn429 && status == 429 {
		return true
	}
	// 408 request timeout could be retried (optional)
	if status == 408 {
		return true
	}
	return false
}

func NextBackoff(now time.Time, attempt int, base, max time.Duration) time.Time {
	// exp backoff: base * 2^(attempt-1), with jitter [0.5, 1.5)
	if attempt < 1 {
		attempt = 1
	}
	d := base
	for i := 1; i < attempt; i++ {
		if d >= max/2 {
			d = max
			break
		}
		d *= 2
	}
	if d > max {
		d = max
	}
	jitter := 0.5 + rand.Float64() // [0.5, 1.5)
	wait := time.Duration(float64(d) * jitter)
	return now.Add(wait)
}
