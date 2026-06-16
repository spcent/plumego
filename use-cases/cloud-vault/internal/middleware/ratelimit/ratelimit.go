package ratelimit

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
)

// Config controls the per-IP rate limiter.
type Config struct {
	// RequestsPerSecond is the sustained request rate allowed per IP.
	RequestsPerSecond float64
	// Burst is the maximum burst of requests allowed.
	Burst int
}

// Middleware returns HTTP middleware that enforces per-IP rate limiting using
// a token bucket algorithm. Requests exceeding the limit receive 429.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	if cfg.RequestsPerSecond <= 0 || cfg.Burst <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	lm := newLimiterMap(cfg)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			if !lm.allow(ip) {
				w.Header().Set("Retry-After", "1")
				_ = contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeRateLimited).
					Message("too many requests").
					Build())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type entry struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

type limiterMap struct {
	mu      sync.RWMutex
	entries map[string]*entry
	cfg     Config
}

func newLimiterMap(cfg Config) *limiterMap {
	lm := &limiterMap{
		entries: make(map[string]*entry),
		cfg:     cfg,
	}
	go lm.cleanupLoop()
	return lm
}

func (lm *limiterMap) cleanupLoop() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		lm.cleanup()
	}
}

func (lm *limiterMap) allow(ip string) bool {
	lm.mu.RLock()
	e, ok := lm.entries[ip]
	lm.mu.RUnlock()
	if !ok {
		lm.mu.Lock()
		e, ok = lm.entries[ip]
		if !ok {
			e = &entry{tokens: float64(lm.cfg.Burst), lastRefill: time.Now()}
			lm.entries[ip] = e
		}
		lm.mu.Unlock()
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(e.lastRefill).Seconds()
	e.tokens += elapsed * lm.cfg.RequestsPerSecond
	if e.tokens > float64(lm.cfg.Burst) {
		e.tokens = float64(lm.cfg.Burst)
	}
	e.lastRefill = now
	if e.tokens < 1 {
		return false
	}
	e.tokens--
	return true
}

func (lm *limiterMap) cleanup() {
	cutoff := time.Now().Add(-10 * time.Minute)
	lm.mu.Lock()
	defer lm.mu.Unlock()
	for ip, e := range lm.entries {
		e.mu.Lock()
		stale := e.lastRefill.Before(cutoff)
		e.mu.Unlock()
		if stale {
			delete(lm.entries, ip)
		}
	}
}

func extractIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		idx := strings.Index(fwd, ",")
		if idx > 0 {
			return strings.TrimSpace(fwd[:idx])
		}
		return strings.TrimSpace(fwd)
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return strings.TrimSpace(rip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
