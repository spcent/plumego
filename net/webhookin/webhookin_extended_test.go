package webhookin

import (
	"testing"
	"time"
)

func TestDeduperWithSweepInterval(t *testing.T) {
	d := NewDeduper(100 * time.Millisecond)
	d.WithSweepInterval(50 * time.Millisecond)

	// Add some keys
	d.SeenBefore("key1")
	d.SeenBefore("key2")

	// Wait for sweep interval
	time.Sleep(60 * time.Millisecond)

	// Add a new key to trigger cleanup
	d.SeenBefore("key3")

	// Count should be 3 (all still valid)
	if count := d.Count(); count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}

	// Wait for TTL to expire
	time.Sleep(120 * time.Millisecond)

	// Add a new key to trigger cleanup
	d.SeenBefore("key4")

	// Count should be 1 (only key4)
	if count := d.Count(); count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}

func TestDeduperClear(t *testing.T) {
	d := NewDeduper(10 * time.Millisecond)

	// Add some keys
	d.SeenBefore("key1")
	d.SeenBefore("key2")
	d.SeenBefore("key3")

	// Wait for all to expire
	time.Sleep(20 * time.Millisecond)

	// Clear should remove all expired entries
	d.Clear()

	if count := d.Count(); count != 0 {
		t.Errorf("expected 0 entries after clear, got %d", count)
	}
}

func TestDeduperClearWithExpired(t *testing.T) {
	d := NewDeduper(10 * time.Millisecond)

	// Add a key
	d.SeenBefore("key1")

	// Wait for it to expire
	time.Sleep(20 * time.Millisecond)

	// Add another key
	d.SeenBefore("key2")

	// Clear should remove expired
	d.Clear()

	// Count should be 1 (only key2)
	if count := d.Count(); count != 1 {
		t.Errorf("expected 1 entry after clear, got %d", count)
	}
}

func TestDeduperCount(t *testing.T) {
	d := NewDeduper(1 * time.Hour)

	// Initially 0
	if count := d.Count(); count != 0 {
		t.Errorf("expected 0 entries initially, got %d", count)
	}

	// Add 3 keys
	d.SeenBefore("key1")
	d.SeenBefore("key2")
	d.SeenBefore("key3")

	if count := d.Count(); count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}

	// Add same keys again (should not increase count)
	d.SeenBefore("key1")
	d.SeenBefore("key2")

	if count := d.Count(); count != 3 {
		t.Errorf("expected 3 entries (no duplicates), got %d", count)
	}
}

func TestDeduperExpiredEntries(t *testing.T) {
	d := NewDeduper(10 * time.Millisecond)

	// Add a key
	d.SeenBefore("key1")

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Key should not be seen as duplicate
	if d.SeenBefore("key1") {
		t.Error("expected key1 to be expired, but was seen as duplicate")
	}

	// Count should be 1 (new entry)
	if count := d.Count(); count != 1 {
		t.Errorf("expected 1 entry after re-adding expired key, got %d", count)
	}
}

func TestDeduperConcurrentAccess(t *testing.T) {
	d := NewDeduper(1 * time.Second)

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := "key" + string(rune(id))
				d.SeenBefore(key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and count should be reasonable
	count := d.Count()
	if count < 0 || count > 10 {
		t.Errorf("unexpected count: %d", count)
	}
}

func TestDeduperDefaultTTL(t *testing.T) {
	d := NewDeduper(0) // Should default to 10 minutes

	if d.ttl != 10*time.Minute {
		t.Errorf("expected default TTL of 10 minutes, got %v", d.ttl)
	}
}

func TestDeduperZeroSweepInterval(t *testing.T) {
	d := NewDeduper(1 * time.Hour)
	initialInterval := d.sweepInterval

	d.WithSweepInterval(0)

	// Should not change
	if d.sweepInterval != initialInterval {
		t.Errorf("sweep interval should not change with zero value")
	}
}

func TestDeduperNegativeSweepInterval(t *testing.T) {
	d := NewDeduper(1 * time.Hour)
	initialInterval := d.sweepInterval

	d.WithSweepInterval(-1 * time.Minute)

	// Should not change
	if d.sweepInterval != initialInterval {
		t.Errorf("sweep interval should not change with negative value")
	}
}

func TestDeduperDoubleCheckMechanism(t *testing.T) {
	d := NewDeduper(1 * time.Hour)

	// First call
	seen1 := d.SeenBefore("key1")
	if seen1 {
		t.Error("first call should return false")
	}

	// Second call (should be seen)
	seen2 := d.SeenBefore("key1")
	if !seen2 {
		t.Error("second call should return true")
	}

	// Third call (should still be seen)
	seen3 := d.SeenBefore("key1")
	if !seen3 {
		t.Error("third call should return true")
	}
}

func TestDeduperSweepCleansExpired(t *testing.T) {
	d := NewDeduper(10 * time.Millisecond)
	d.WithSweepInterval(5 * time.Millisecond)

	// Add keys
	d.SeenBefore("key1")
	d.SeenBefore("key2")

	// Wait for keys to expire
	time.Sleep(20 * time.Millisecond)

	// Trigger sweep by adding new key
	d.SeenBefore("key3")

	// Wait for sweep interval
	time.Sleep(10 * time.Millisecond)

	// Add another key to ensure sweep happens
	d.SeenBefore("key4")

	// Count should be 2 (key3 and key4), but allow for timing variations
	count := d.Count()
	if count < 1 || count > 3 {
		t.Errorf("unexpected entry count after sweep: %d", count)
	}
}

func TestDeduperZeroTTL(t *testing.T) {
	// Test with zero TTL (should default to 10 minutes)
	d := NewDeduper(0)

	if d.ttl != 10*time.Minute {
		t.Errorf("expected default 10 minute TTL, got %v", d.ttl)
	}

	// Should work normally
	if d.SeenBefore("test") {
		t.Error("first call should return false")
	}
	if !d.SeenBefore("test") {
		t.Error("second call should return true")
	}
}

func TestDeduperNegativeTTL(t *testing.T) {
	// Test with negative TTL (should default to 10 minutes)
	d := NewDeduper(-1 * time.Hour)

	if d.ttl != 10*time.Minute {
		t.Errorf("expected default 10 minute TTL, got %v", d.ttl)
	}
}

func TestDeduperLongSweepInterval(t *testing.T) {
	// Test that sweep doesn't happen too frequently
	d := NewDeduper(1 * time.Hour)
	d.WithSweepInterval(10 * time.Second)

	// Add some keys
	for i := 0; i < 100; i++ {
		d.SeenBefore("key" + string(rune(i)))
	}

	// Count should be 100
	if count := d.Count(); count != 100 {
		t.Errorf("expected 100 entries, got %d", count)
	}

	// Add more keys quickly (should not trigger sweep)
	for i := 100; i < 150; i++ {
		d.SeenBefore("key" + string(rune(i)))
	}

	// Count should be 150 (no sweep happened yet)
	if count := d.Count(); count != 150 {
		t.Errorf("expected 150 entries, got %d", count)
	}
}
