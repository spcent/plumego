package glog

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

const (
	base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	// Custom epoch to reduce timestamp bits.
	// 2020-01-01T00:00:00Z in milliseconds.
	epochMilli int64 = 1577836800000

	// Bit layout
	tsBits   = 40
	randBits = 14
	seqBits  = 10

	seqMax  = (1 << seqBits) - 1 // 1023
	randMax = 10000              // 0..9999

	// v = (ts << (randBits+seqBits)) | (rand << seqBits) | seq
	shiftRand = seqBits
	shiftTS   = randBits + seqBits

	// Fixed output width
	idWidth = 12
)

var (
	errInvalidBase62 = errors.New("invalid base62 trace id")
	errOutOfRange    = errors.New("trace id out of range")

	rng   = rand.New(rand.NewSource(time.Now().UnixNano()))
	rngMu sync.Mutex

	genMu  sync.Mutex
	lastMs int64
	seq    int
)

// NewTraceID generates a fixed 12-char Base62 trace id.
// Reversible: can be decoded back to (unixMilli, rand[0..9999], seq[0..1023]).
func NewTraceID() string {
	genMu.Lock()
	defer genMu.Unlock()

	for {
		nowMs := time.Now().UnixMilli()
		if nowMs == lastMs {
			if seq < seqMax {
				seq++
				return buildID(nowMs, seq)
			}
			// seq exhausted in the same millisecond; wait for next millisecond
			time.Sleep(time.Millisecond)
			continue
		}

		// new millisecond
		lastMs = nowMs
		seq = 0
		return buildID(nowMs, seq)
	}
}

func buildID(unixMs int64, seqVal int) string {
	// rand in [0..9999]
	rngMu.Lock()
	r := rng.Intn(randMax)
	rngMu.Unlock()

	// timestamp delta
	delta := unixMs - epochMilli
	if delta < 0 {
		// if system clock is earlier than epoch, clamp to 0
		delta = 0
	}

	// fit into 40 bits
	ts := uint64(delta) & ((uint64(1) << tsBits) - 1)
	v := (ts << shiftTS) | (uint64(r) << shiftRand) | uint64(seqVal)

	return encodeBase62Fixed(v, idWidth)
}

// DecodeTraceID reverses a 12-char Base62 trace id into:
// - unixMilli (int64)
// - rand (0..9999)
// - seq (0..1023)
func DecodeTraceID(id string) (unixMilli int64, r int, seqVal int, err error) {
	if len(id) != idWidth {
		return 0, 0, 0, errInvalidBase62
	}

	v, err := decodeBase62(id)
	if err != nil {
		return 0, 0, 0, err
	}

	seqVal = int(v & uint64(seqMax))
	r = int((v >> shiftRand) & ((uint64(1) << randBits) - 1))
	ts := (v >> shiftTS) & ((uint64(1) << tsBits) - 1)

	// Validate ranges (rand was stored in 14 bits; ensure still in [0..9999])
	if r < 0 || r >= randMax {
		return 0, 0, 0, errOutOfRange
	}

	unixMilli = int64(ts) + epochMilli
	return unixMilli, r, seqVal, nil
}

func encodeBase62Fixed(v uint64, width int) string {
	// left padded with '0'
	buf := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		buf[i] = base62Alphabet[v%62]
		v /= 62
	}
	return string(buf)
}

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
