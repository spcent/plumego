package log

import (
	"sync"
	"testing"
	"time"
)

// TestNewRequestIDGeneration tests basic request ID generation.
func TestNewRequestIDGeneration(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := NewRequestID()
		if len(id) != idWidth {
			t.Errorf("Expected request ID length %d, got %d", idWidth, len(id))
		}
		if ids[id] {
			t.Errorf("Duplicate request ID generated: %s", id)
		}
		ids[id] = true
	}
}

// TestDecodeRequestID tests request ID decoding functionality.
func TestDecodeRequestID(t *testing.T) {
	originalID := NewRequestID()
	unixMilli, r, seqVal, err := DecodeRequestID(originalID)
	if err != nil {
		t.Errorf("Failed to decode valid request ID: %v", err)
	}

	// Verify decoded values are within expected ranges
	if unixMilli < epochMilli {
		t.Errorf("Decoded timestamp before epoch: %d", unixMilli)
	}
	if r < 0 || r >= randMax {
		t.Errorf("Decoded random value out of range: %d", r)
	}
	if seqVal < 0 || seqVal > seqMax {
		t.Errorf("Decoded sequence value out of range: %d", seqVal)
	}
}

// TestInvalidRequestID tests error handling for invalid request IDs.
func TestInvalidRequestID(t *testing.T) {
	_, _, _, err := DecodeRequestID("")
	if err != errInvalidBase62 {
		t.Errorf("Expected errInvalidBase62 for empty string, got %v", err)
	}

	_, _, _, err = DecodeRequestID("short")
	if err != errInvalidBase62 {
		t.Errorf("Expected errInvalidBase62 for wrong length, got %v", err)
	}

	_, _, _, err = DecodeRequestID("!!!@#$%^&*()")
	if err != errInvalidBase62 {
		t.Errorf("Expected errInvalidBase62 for invalid characters, got %v", err)
	}
}

// TestConcurrentGeneration tests concurrent request ID generation for thread safety.
func TestConcurrentGeneration(t *testing.T) {
	const numGoroutines = 100
	const idsPerGoroutine = 100
	var wg sync.WaitGroup

	ids := make(map[string]bool, numGoroutines*idsPerGoroutine)
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id := NewRequestID()
				mu.Lock()
				if ids[id] {
					t.Errorf("Duplicate ID generated in concurrent test: %s", id)
				}
				ids[id] = true
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
}

// TestGeneratorStruct tests the RequestIDGenerator struct functionality.
func TestGeneratorStruct(t *testing.T) {
	gen := NewRequestIDGenerator()

	// Test that generator produces valid IDs
	for i := 0; i < 50; i++ {
		id := gen.Generate()
		if len(id) != idWidth {
			t.Errorf("Expected length %d, got %d", idWidth, len(id))
		}

		// Verify decoding works
		_, _, _, err := DecodeRequestID(id)
		if err != nil {
			t.Errorf("Failed to decode generated ID: %v", err)
		}
	}
}

func TestZeroValueGenerator(t *testing.T) {
	var gen RequestIDGenerator

	id := gen.Generate()
	if len(id) != idWidth {
		t.Fatalf("expected id length %d, got %d", idWidth, len(id))
	}

	if _, _, _, err := DecodeRequestID(id); err != nil {
		t.Fatalf("failed to decode id from zero value generator: %v", err)
	}
}

// TestTimestampOrdering tests that newer request IDs have newer timestamps.
func TestTimestampOrdering(t *testing.T) {
	time.Sleep(time.Millisecond) // Ensure different millisecond
	gen := NewRequestIDGenerator()

	// Generate IDs with small delays
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond)
		ids[i] = gen.Generate()
	}

	// Verify timestamps are increasing
	for i := 1; i < len(ids); i++ {
		ts1, _, _, _ := DecodeRequestID(ids[i-1])
		ts2, _, _, _ := DecodeRequestID(ids[i])
		if ts2 < ts1 {
			t.Errorf("Timestamp not increasing: %d >= %d", ts1, ts2)
		}
	}
}

// BenchmarkRequestIDGeneration benchmarks the request ID generation performance.
func BenchmarkRequestIDGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRequestID()
	}
}

// BenchmarkConcurrentGeneration benchmarks concurrent request ID generation.
func BenchmarkConcurrentGeneration(b *testing.B) {
	var wg sync.WaitGroup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = NewRequestID()
		}()
	}
	wg.Wait()
}

// TestRandomness tests that random values are reasonably distributed
func TestRandomness(t *testing.T) {
	const testCount = 1000
	randValues := make([]int, testCount)

	// Collect random values from decoded request IDs
	for i := 0; i < testCount; i++ {
		id := NewRequestID()
		_, r, _, err := DecodeRequestID(id)
		if err != nil {
			t.Errorf("Failed to decode ID for randomness test: %v", err)
			return
		}
		randValues[i] = r
	}

	// Check distribution (simple variance test)
	sum := 0
	for _, r := range randValues {
		sum += r
	}
	avg := float64(sum) / float64(testCount)

	// Expected average should be close to randMax/2
	expectedAvg := float64(randMax) / 2.0
	if avg < expectedAvg-100 || avg > expectedAvg+100 {
		t.Errorf("Random distribution seems skewed: average %f, expected ~%f", avg, expectedAvg)
	}
}
