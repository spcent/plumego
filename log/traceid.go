package log

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	mathrand "math/rand"
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

	// Maximum values for each component.
	// randMask covers the full 14-bit random field (0..16383).
	// randMax is derived from randMask so generation and decoding are consistent.
	seqMax   = (1 << seqBits) - 1  // 1023
	randMask = (1 << randBits) - 1 // 16383
	randMax  = randMask + 1        // 16384 – upper bound for rand.Intn, fills all 14 bits

	// Bit shifts for encoding
	shiftRand = seqBits
	shiftTS   = randBits + seqBits

	// Fixed output width
	idWidth = 12

	// Random pool size for fast generation.
	randomPoolSize = 4096

	// maxSeqRetries caps the busy-wait loop when sequence is exhausted within a
	// millisecond. After this many retries we fall back to a seq=0 ID using the
	// current timestamp, accepting a very small duplicate risk rather than blocking
	// indefinitely under extreme load.
	maxSeqRetries = 20
)

// Error definitions
var (
	errInvalidBase62 = errors.New("invalid base62 request id")
	errOutOfRange    = errors.New("request id out of range")
)

// RequestIDGenerator encapsulates the state for request ID generation.
type RequestIDGenerator struct {
	rng        *mathrand.Rand
	stateMu    sync.Mutex
	lastMs     int64
	seq        int32
	randomPool []int32
	poolIdx    atomic.Int32
	poolMu     sync.RWMutex
}

// cryptoSeed returns a cryptographically secure seed for math/rand.
func cryptoSeed() int64 {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// Fallback to time-based seed if crypto/rand fails (should not happen)
		return time.Now().UnixNano()
	}
	return int64(binary.LittleEndian.Uint64(buf[:]))
}

// NewRequestIDGenerator creates a new request ID generator with pre-generated random pool.
func NewRequestIDGenerator() *RequestIDGenerator {
	g := &RequestIDGenerator{
		rng:        mathrand.New(mathrand.NewSource(cryptoSeed())),
		randomPool: make([]int32, randomPoolSize), // Pre-generate random numbers
	}

	// Pre-generate random numbers to avoid runtime generation overhead
	for i := range g.randomPool {
		g.randomPool[i] = int32(g.rng.Intn(randMax))
	}

	return g
}

var globalGen = NewRequestIDGenerator()

// NewRequestID generates a fixed 12-char Base62 request id using the global generator.
// Reversible: can be decoded back to (unixMilli, rand[0..9999], seq[0..1023]).
func NewRequestID() string {
	return globalGen.Generate()
}

// Generate creates a new trace ID with monotonic timestamp/sequence coordination.
// If the per-millisecond sequence is exhausted, it retries up to maxSeqRetries times
// (each retry sleeps 100µs). After that it falls back to seq=0 on the next available
// timestamp to avoid blocking indefinitely under burst load.
func (g *RequestIDGenerator) Generate() string {
	g.ensureInitialized()
	for attempt := 0; attempt < maxSeqRetries; attempt++ {
		nowMs := time.Now().UnixMilli()
		g.stateMu.Lock()
		if nowMs < g.lastMs {
			nowMs = g.lastMs
		}

		if nowMs == g.lastMs {
			// Same millisecond: try to increment sequence.
			if g.seq < int32(seqMax) {
				g.seq++
				seq := g.seq
				g.stateMu.Unlock()
				return g.buildID(nowMs, int(seq))
			}
			// Sequence exhausted; yield and retry.
			g.stateMu.Unlock()
			time.Sleep(time.Microsecond * 100)
			continue
		}

		// New millisecond: reset sequence.
		g.lastMs = nowMs
		g.seq = 0
		g.stateMu.Unlock()
		return g.buildID(nowMs, 0)
	}

	// Fallback after maxSeqRetries: use current time with seq=0.
	// Duplicate risk is negligible; blocking is not acceptable.
	nowMs := time.Now().UnixMilli()
	g.stateMu.Lock()
	if nowMs < g.lastMs {
		nowMs = g.lastMs
	}
	g.lastMs = nowMs
	g.seq = 0
	g.stateMu.Unlock()
	return g.buildID(nowMs, 0)
}

// buildID constructs the trace ID from components using optimized random number access
func (g *RequestIDGenerator) buildID(unixMs int64, seqVal int) string {
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

func (g *RequestIDGenerator) ensureInitialized() {
	if g.rng != nil && len(g.randomPool) > 0 {
		return
	}

	g.poolMu.Lock()
	defer g.poolMu.Unlock()

	if g.rng == nil {
		g.rng = mathrand.New(mathrand.NewSource(cryptoSeed()))
	}
	if len(g.randomPool) == 0 {
		g.randomPool = make([]int32, randomPoolSize)
	}
	for i := range g.randomPool {
		g.randomPool[i] = int32(g.rng.Intn(randMax))
	}
}

func (g *RequestIDGenerator) randomValue() int {
	g.ensureInitialized()

	g.poolMu.RLock()
	poolSize := len(g.randomPool)
	g.poolMu.RUnlock()
	if poolSize == 0 {
		return g.randomFallback()
	}

	idx := uint32(g.poolIdx.Add(1) - 1)
	pos := int(idx % uint32(poolSize))

	if pos == 0 && idx != 0 {
		g.refreshRandomPool()
	}

	g.poolMu.RLock()
	v := int(g.randomPool[pos])
	g.poolMu.RUnlock()
	return v
}

func (g *RequestIDGenerator) refreshRandomPool() {
	g.poolMu.Lock()
	defer g.poolMu.Unlock()

	if g.rng == nil {
		g.rng = mathrand.New(mathrand.NewSource(cryptoSeed()))
	}
	for i := range g.randomPool {
		g.randomPool[i] = int32(g.rng.Intn(randMax))
	}
}

func (g *RequestIDGenerator) randomFallback() int {
	g.poolMu.Lock()
	defer g.poolMu.Unlock()

	if g.rng == nil {
		g.rng = mathrand.New(mathrand.NewSource(cryptoSeed()))
	}
	return g.rng.Intn(randMax)
}

// DecodeRequestID reverses a 12-char Base62 request id into components.
// Returns: unixMilli, randValue, seqValue, error.
func DecodeRequestID(id string) (unixMilli int64, r int, seqVal int, err error) {
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

	// Validate decoded values: rand component must fit within the 14-bit mask.
	if r < 0 || r > randMask {
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
