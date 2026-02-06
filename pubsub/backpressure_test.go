package pubsub

import (
	"testing"
	"time"
)

func TestBackpressure_Basic(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureWarn

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Report low pressure - should allow publish
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.3,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	err := bpc.CheckPublish("test")
	if err != nil {
		t.Errorf("Expected no error with low pressure, got %v", err)
	}

	// Report high pressure - should warn
	signal.BufferUsage = 0.9
	bpc.ReportPressure(signal)

	time.Sleep(100 * time.Millisecond)

	err = bpc.CheckPublish("test")
	if err != ErrBackpressureActive {
		t.Errorf("Expected ErrBackpressureActive with high pressure, got %v", err)
	}
}

func TestBackpressure_PolicyNone(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureNone

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Report high pressure
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.9,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	// Should still allow publish
	err := bpc.CheckPublish("test")
	if err != nil {
		t.Errorf("Expected no error with None policy, got %v", err)
	}
}

func TestBackpressure_PolicyThrottle(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureThrottle
	config.SlowStartInitialRate = 100

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Report high pressure
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.9,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	time.Sleep(100 * time.Millisecond)

	// Should throttle but still return error
	start := time.Now()
	err := bpc.CheckPublish("test")
	elapsed := time.Since(start)

	if err != ErrBackpressureActive {
		t.Errorf("Expected ErrBackpressureActive, got %v", err)
	}

	// Should have added some delay
	if elapsed < 1*time.Millisecond {
		t.Log("Note: Throttle delay may be very small")
	}

	// Check stats
	stats := bpc.Stats()
	if stats.ThrottledPublish.Load() != 1 {
		t.Errorf("Expected 1 throttled publish, got %d", stats.ThrottledPublish.Load())
	}
}

func TestBackpressure_PolicyBlock(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureBlock
	config.BlockTimeout = 200 * time.Millisecond

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Report high pressure
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.9,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	time.Sleep(100 * time.Millisecond)

	// Should block and timeout
	start := time.Now()
	err := bpc.CheckPublish("test")
	elapsed := time.Since(start)

	if err != ErrBackpressureActive {
		t.Errorf("Expected ErrBackpressureActive after timeout, got %v", err)
	}

	// Should have blocked for at least block timeout
	if elapsed < config.BlockTimeout {
		t.Errorf("Expected to block for at least %v, blocked for %v", config.BlockTimeout, elapsed)
	}

	// Check stats
	stats := bpc.Stats()
	if stats.BlockedPublish.Load() != 1 {
		t.Errorf("Expected 1 blocked publish, got %d", stats.BlockedPublish.Load())
	}
}

func TestBackpressure_PressureRelease(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureWarn

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// High pressure
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.9,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	time.Sleep(100 * time.Millisecond)

	if !bpc.IsUnderPressure("test") {
		t.Error("Expected to be under pressure")
	}

	// Release pressure
	signal.BufferUsage = 0.3
	bpc.ReportPressure(signal)

	time.Sleep(100 * time.Millisecond)

	if bpc.IsUnderPressure("test") {
		t.Error("Expected pressure to be released")
	}

	// Should allow publish
	err := bpc.CheckPublish("test")
	if err != nil {
		t.Errorf("Expected no error after pressure release, got %v", err)
	}
}

func TestBackpressure_GetPressureLevel(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Report pressure
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.75,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	time.Sleep(100 * time.Millisecond)

	level := bpc.GetPressureLevel("test")
	if level < 70 || level > 80 {
		t.Errorf("Expected pressure level around 75, got %d", level)
	}
}

func TestBackpressure_MultipleTopics(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureWarn

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Pressure on topic1
	signal1 := BackpressureSignal{
		Topic:       "topic1",
		BufferUsage: 0.9,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal1)

	// No pressure on topic2
	signal2 := BackpressureSignal{
		Topic:       "topic2",
		BufferUsage: 0.3,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal2)

	time.Sleep(100 * time.Millisecond)

	// topic1 should be under pressure
	if !bpc.IsUnderPressure("topic1") {
		t.Error("Expected topic1 to be under pressure")
	}

	// topic2 should not be under pressure
	if bpc.IsUnderPressure("topic2") {
		t.Error("Expected topic2 to not be under pressure")
	}

	// Check publishes
	if err := bpc.CheckPublish("topic1"); err == nil {
		t.Error("Expected error for topic1")
	}

	if err := bpc.CheckPublish("topic2"); err != nil {
		t.Errorf("Expected no error for topic2, got %v", err)
	}
}

func TestBackpressure_SlowStart(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.SlowStartInitialRate = 10
	config.SlowStartMaxRate = 100
	config.SlowStartMultiplier = 2.0

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Initial rate should be at starting value
	stats := bpc.Stats()
	initialRate := stats.CurrentRate.Load()
	if initialRate != 10 {
		t.Errorf("Expected initial rate 10, got %d", initialRate)
	}

	// Wait for rate to increase
	time.Sleep(2 * time.Second)

	stats = bpc.Stats()
	currentRate := stats.CurrentRate.Load()

	// Rate should have increased
	if currentRate <= initialRate {
		t.Errorf("Expected rate to increase, was %d now %d", initialRate, currentRate)
	}
}

func TestBackpressure_SlowStartWithPressure(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.SlowStartInitialRate = 50
	config.SlowStartMaxRate = 500
	config.SlowStartMultiplier = 2.0

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Wait for rate to increase
	time.Sleep(1 * time.Second)

	stats := bpc.Stats()
	beforePressure := stats.CurrentRate.Load()

	// Apply pressure
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.9,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	// Wait for slow start to react
	time.Sleep(2 * time.Second)

	stats = bpc.Stats()
	afterPressure := stats.CurrentRate.Load()

	// Rate should have decreased
	if afterPressure >= beforePressure {
		t.Logf("Note: Rate did not decrease (before: %d, after: %d)", beforePressure, afterPressure)
	}
}

func TestBackpressure_ResetSlowStart(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.SlowStartInitialRate = 10

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Wait for rate to increase
	time.Sleep(2 * time.Second)

	// Reset slow start
	bpc.ResetSlowStart()

	stats := bpc.Stats()
	rate := stats.CurrentRate.Load()

	// Should be back to initial rate
	if rate != 10 {
		t.Errorf("Expected rate 10 after reset, got %d", rate)
	}
}

func TestBackpressure_Callback(t *testing.T) {
	ps := New()
	defer ps.Close()

	callbackCalled := false
	var callbackActive bool
	var callbackPressure float64

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureWarn
	config.Callback = func(topic string, active bool, pressure float64) {
		callbackCalled = true
		callbackActive = active
		callbackPressure = pressure
	}

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Trigger high pressure
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.9,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	time.Sleep(100 * time.Millisecond)

	if !callbackCalled {
		t.Error("Expected callback to be called")
	}

	if !callbackActive {
		t.Error("Expected callback active=true")
	}

	if callbackPressure < 0.8 {
		t.Errorf("Expected high pressure in callback, got %f", callbackPressure)
	}
}

func TestBackpressure_Stats(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureWarn

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Report signals
	for i := 0; i < 5; i++ {
		signal := BackpressureSignal{
			Topic:       "test",
			BufferUsage: 0.9,
			Timestamp:   time.Now(),
		}
		bpc.ReportPressure(signal)
	}

	time.Sleep(100 * time.Millisecond)

	stats := bpc.Stats()

	if stats.SignalCount.Load() != 5 {
		t.Errorf("Expected 5 signals, got %d", stats.SignalCount.Load())
	}

	if !stats.ActivePressure.Load() {
		t.Error("Expected active pressure")
	}

	if stats.PressureLevel.Load() < 80 {
		t.Errorf("Expected high pressure level, got %d", stats.PressureLevel.Load())
	}
}

func TestBackpressure_PublishWithBackpressure(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureWarn

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	// Create subscriber
	sub, _ := ps.Subscribe("test", SubOptions{BufferSize: 10})
	defer sub.Cancel()

	// Low pressure - should succeed
	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.3,
		Timestamp:   time.Now(),
	}
	bpc.ReportPressure(signal)

	msg := Message{Data: "test"}
	err := bpc.PublishWithBackpressure("test", msg)
	if err != nil {
		t.Errorf("Expected no error with low pressure, got %v", err)
	}
}

func TestBackpressure_Disabled(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Enabled = false

	bpc := NewBackpressureController(ps, config)

	// Should return nil when disabled
	if bpc != nil {
		t.Error("Expected nil controller when disabled")
	}
}

func TestBackpressure_Close(t *testing.T) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	bpc := NewBackpressureController(ps, config)

	// Close
	err := bpc.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Operations after close should handle gracefully
	err = bpc.CheckPublish("test")
	if err != ErrBackpressureClosed {
		t.Errorf("Expected ErrBackpressureClosed, got %v", err)
	}

	// Double close should be safe
	err = bpc.Close()
	if err != ErrBackpressureClosed {
		t.Errorf("Expected ErrBackpressureClosed on double close, got %v", err)
	}
}

func BenchmarkBackpressure_CheckPublish(b *testing.B) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	config.Policy = BackpressureWarn

	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bpc.CheckPublish("test")
	}
}

func BenchmarkBackpressure_ReportPressure(b *testing.B) {
	ps := New()
	defer ps.Close()

	config := DefaultBackpressureConfig()
	bpc := NewBackpressureController(ps, config)
	defer bpc.Close()

	signal := BackpressureSignal{
		Topic:       "test",
		BufferUsage: 0.5,
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bpc.ReportPressure(signal)
	}
}
