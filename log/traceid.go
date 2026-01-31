package glog

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	// Custom epoch to reduce timestamp bits: 2020-01-01T00:00:00Z
	epochMilli int64 = 1577836800000

	// Bit layout configuration
	tsBits   = 40
	randBits = 14
	seqBits  = 10

	// Maximum values for each component
	seqMax   = (1 << seqBits) - 1 // 1023
	randMax  = 10000              // 0..9999
	randMask = (1 << randBits) - 1

	// Bit shifts for encoding
	shiftRand = seqBits
	shiftTS   = randBits + seqBits

	// Fixed output width
	idWidth = 12

	// Random pool size for fast generation.
	randomPoolSize = 4096
)

// Error definitions
var (
	errInvalidBase62 = errors.New("invalid base62 trace id")
	errOutOfRange    = errors.New("trace id out of range")
)

// TraceIDGenerator encapsulates the state for trace ID generation
type TraceIDGenerator struct {
	rng        *rand.Rand
	lastMs     atomic.Int64
	seq        atomic.Int32
	randomPool []int32
	poolIdx    atomic.Int32
	poolMu     sync.Mutex
}

// NewTraceIDGenerator creates a new trace ID generator with pre-generated random pool
func NewTraceIDGenerator() *TraceIDGenerator {
	g := &TraceIDGenerator{
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		randomPool: make([]int32, randomPoolSize), // Pre-generate random numbers
	}

	// Pre-generate random numbers to avoid runtime generation overhead
	for i := range g.randomPool {
		g.randomPool[i] = int32(g.rng.Intn(randMax))
	}

	return g
}

// Global generator instance for backward compatibility
var globalGen = NewTraceIDGenerator()

// NewTraceID generates a fixed 12-char Base62 trace id using global generator
// Reversible: can be decoded back to (unixMilli, rand[0..9999], seq[0..1023])
func NewTraceID() string {
	return globalGen.Generate()
}

// Generate creates a new trace ID using optimized atomic operations
func (g *TraceIDGenerator) Generate() string {
	g.ensureInitialized()
	for {
		nowMs := time.Now().UnixMilli()
		lastMs := g.lastMs.Load()
		if nowMs < lastMs {
			nowMs = lastMs
		}

		if nowMs == lastMs {
			// Same millisecond, try to increment sequence
			for {
				currentSeq := g.seq.Load()
				if currentSeq < int32(seqMax) {
					// Try to atomically increment sequence
					if g.seq.CompareAndSwap(currentSeq, currentSeq+1) {
						return g.buildID(nowMs, int(currentSeq+1))
					}
					// CAS failed, retry
					continue
				}
				// Sequence exhausted, wait for next millisecond
				time.Sleep(time.Microsecond * 100) // Short sleep to reduce CPU usage
				break
			}
			continue
		}

		// New millisecond, reset sequence to 0
		g.lastMs.Store(nowMs)
		g.seq.Store(0)
		return g.buildID(nowMs, 0)
	}
}

// buildID constructs the trace ID from components using optimized random number access
func (g *TraceIDGenerator) buildID(unixMs int64, seqVal int) string {
	r := g.randomValue()

	// Encode timestamp delta
	delta := unixMs - epochMilli
	if delta < 0 {
		delta = 0
	}

	// Encode into 64-bit value
	ts := uint64(delta) & ((uint64(1) << tsBits) - 1)
	v := (ts << shiftTS) | (uint64(r&randMask) << shiftRand) | uint64(seqVal)

	return encodeBase62Fixed(v, idWidth)
}

func (g *TraceIDGenerator) ensureInitialized() {
	if g.rng != nil && len(g.randomPool) > 0 {
		return
	}

	g.poolMu.Lock()
	defer g.poolMu.Unlock()

	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	if len(g.randomPool) == 0 {
		g.randomPool = make([]int32, randomPoolSize)
	}
	for i := range g.randomPool {
		g.randomPool[i] = int32(g.rng.Intn(randMax))
	}
}

func (g *TraceIDGenerator) randomValue() int {
	g.ensureInitialized()

	poolSize := len(g.randomPool)
	if poolSize == 0 {
		return g.randomFallback()
	}

	idx := uint32(g.poolIdx.Add(1) - 1)
	pos := int(idx % uint32(poolSize))

	if pos == 0 && idx != 0 {
		g.refreshRandomPool()
	}

	return int(g.randomPool[pos])
}

func (g *TraceIDGenerator) refreshRandomPool() {
	g.poolMu.Lock()
	defer g.poolMu.Unlock()

	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	for i := range g.randomPool {
		g.randomPool[i] = int32(g.rng.Intn(randMax))
	}
}

func (g *TraceIDGenerator) randomFallback() int {
	g.poolMu.Lock()
	defer g.poolMu.Unlock()

	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return g.rng.Intn(randMax)
}

// DecodeTraceID reverses a 12-char Base62 trace id into components
// Returns: unixMilli, randValue, seqValue, error
func DecodeTraceID(id string) (unixMilli int64, r int, seqVal int, err error) {
	if len(id) != idWidth {
		return 0, 0, 0, errInvalidBase62
	}

	v, err := decodeBase62(id)
	if err != nil {
		return 0, 0, 0, err
	}

	// Decode components using bit operations
	seqVal = int(v & uint64(seqMax))
	r = int((v >> shiftRand) & uint64(randMask))
	ts := (v >> shiftTS) & ((uint64(1) << tsBits) - 1)

	// Validate decoded values
	if r < 0 || r >= randMax {
		return 0, 0, 0, errOutOfRange
	}

	unixMilli = int64(ts) + epochMilli
	return unixMilli, r, seqVal, nil
}

// encodeBase62Fixed converts uint64 to fixed-width Base62 string
func encodeBase62Fixed(v uint64, width int) string {
	buf := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		buf[i] = base62Alphabet[v%62]
		v /= 62
	}
	return string(buf)
}

// decodeBase62 converts Base62 string to uint64
func decodeBase62(s string) (uint64, error) {
	var v uint64
	for i := 0; i < len(s); i++ {
		d, ok := base62Value(s[i])
		if !ok {
			return 0, errInvalidBase62
		}
		v = v*62 + uint64(d)
	}
	return v, nil
}

// base62Value converts Base62 character to its numeric value
func base62Value(c byte) (int, bool) {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0'), true
	case c >= 'A' && c <= 'Z':
		return int(c-'A') + 10, true
	case c >= 'a' && c <= 'z':
		return int(c-'a') + 36, true
	default:
		return 0, false
	}
}
