package pubsub

import (
	"testing"
	"time"
)

func TestMultiTenant_CreateTenant(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Create tenant
	err := mtps.CreateTenant("tenant1", DefaultTenantQuota())
	if err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Verify tenant exists
	tenants := mtps.ListTenants()
	if len(tenants) != 1 {
		t.Errorf("Expected 1 tenant, got %d", len(tenants))
	}

	// Try to create duplicate
	err = mtps.CreateTenant("tenant1", DefaultTenantQuota())
	if err != ErrTenantAlreadyExists {
		t.Errorf("Expected ErrTenantAlreadyExists, got %v", err)
	}
}

func TestMultiTenant_DeleteTenant(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Create and delete tenant
	_ = mtps.CreateTenant("tenant1", DefaultTenantQuota())

	err := mtps.DeleteTenant("tenant1")
	if err != nil {
		t.Fatalf("Failed to delete tenant: %v", err)
	}

	// Verify tenant is gone
	tenants := mtps.ListTenants()
	if len(tenants) != 0 {
		t.Errorf("Expected 0 tenants, got %d", len(tenants))
	}

	// Try to delete non-existent tenant
	err = mtps.DeleteTenant("nonexistent")
	if err != ErrTenantNotFound {
		t.Errorf("Expected ErrTenantNotFound, got %v", err)
	}
}

func TestMultiTenant_PublishWithNamespace(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Create two tenants
	_ = mtps.CreateTenant("tenant1", DefaultTenantQuota())
	_ = mtps.CreateTenant("tenant2", DefaultTenantQuota())

	// Subscribe to tenant1's topic
	sub1, err := mtps.Subscribe("tenant1", "test", SubOptions{BufferSize: 10})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub1.Cancel()

	// Subscribe to tenant2's topic
	sub2, err := mtps.Subscribe("tenant2", "test", SubOptions{BufferSize: 10})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub2.Cancel()

	// Publish to tenant1
	msg := Message{Data: "tenant1 message"}
	_ = mtps.Publish("tenant1", "test", msg)

	time.Sleep(100 * time.Millisecond)

	// tenant1 should receive message
	select {
	case receivedMsg := <-sub1.C():
		if receivedMsg.Data != "tenant1 message" {
			t.Error("tenant1 received wrong message")
		}
	case <-time.After(time.Second):
		t.Error("tenant1 did not receive message")
	}

	// tenant2 should NOT receive message
	select {
	case <-sub2.C():
		t.Error("tenant2 should not receive tenant1's message")
	case <-time.After(100 * time.Millisecond):
		// Expected - no message
	}
}

func TestMultiTenant_PublishRateQuota(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Create tenant with low rate limit
	quota := DefaultTenantQuota()
	quota.MaxPublishRate = 5
	_ = mtps.CreateTenant("tenant1", quota)

	// Try to publish more than quota
	successCount := 0
	quotaErrors := 0

	for i := 0; i < 10; i++ {
		msg := Message{Data: i}
		err := mtps.Publish("tenant1", "test", msg)
		if err == nil {
			successCount++
		} else if err == ErrQuotaExceeded || err.Error() == "quota exceeded: publish rate limit exceeded" {
			quotaErrors++
		}
	}

	if successCount > quota.MaxPublishRate {
		t.Errorf("Published %d messages, expected at most %d", successCount, quota.MaxPublishRate)
	}

	if quotaErrors == 0 {
		t.Error("Expected quota exceeded errors")
	}
}

func TestMultiTenant_SubscriptionQuota(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Create tenant with low subscription limit
	quota := DefaultTenantQuota()
	quota.MaxSubscriptions = 2
	_ = mtps.CreateTenant("tenant1", quota)

	// Create subscriptions
	subs := make([]Subscription, 0)

	for i := 0; i < 3; i++ {
		sub, err := mtps.Subscribe("tenant1", "test", SubOptions{BufferSize: 10})
		if err == nil {
			subs = append(subs, sub)
		}
	}

	// Should only have 2 subscriptions
	if len(subs) > quota.MaxSubscriptions {
		t.Errorf("Created %d subscriptions, expected at most %d", len(subs), quota.MaxSubscriptions)
	}

	// Cleanup
	for _, sub := range subs {
		sub.Cancel()
	}
}

func TestMultiTenant_MessageSizeQuota(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Create tenant with size limit
	quota := DefaultTenantQuota()
	quota.MaxMessageSize = 100
	_ = mtps.CreateTenant("tenant1", quota)

	// Try to publish large message
	largeData := make([]byte, 200)
	msg := Message{Data: largeData}

	err := mtps.Publish("tenant1", "test", msg)
	if err == nil {
		t.Error("Expected quota exceeded error for large message")
	}
}

func TestMultiTenant_Usage(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	_ = mtps.CreateTenant("tenant1", DefaultTenantQuota())

	// Publish messages
	for i := 0; i < 5; i++ {
		msg := Message{Data: i}
		_ = mtps.Publish("tenant1", "test", msg)
	}

	time.Sleep(100 * time.Millisecond)

	// Check usage
	usage, err := mtps.GetUsage("tenant1")
	if err != nil {
		t.Fatalf("Failed to get usage: %v", err)
	}

	if usage.PublishCount.Load() != 5 {
		t.Errorf("Expected 5 publishes, got %d", usage.PublishCount.Load())
	}
}

func TestMultiTenant_UpdateQuota(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	_ = mtps.CreateTenant("tenant1", DefaultTenantQuota())

	// Update quota
	newQuota := DefaultTenantQuota()
	newQuota.MaxPublishRate = 50

	err := mtps.UpdateQuota("tenant1", newQuota)
	if err != nil {
		t.Fatalf("Failed to update quota: %v", err)
	}

	// Try to update non-existent tenant
	err = mtps.UpdateQuota("nonexistent", newQuota)
	if err != ErrTenantNotFound {
		t.Errorf("Expected ErrTenantNotFound, got %v", err)
	}
}

func TestMultiTenant_AutoCreate(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Publish without creating tenant (should auto-create)
	msg := Message{Data: "test"}
	err := mtps.Publish("auto-tenant", "test", msg)
	if err != nil {
		t.Fatalf("Failed to publish (auto-create should work): %v", err)
	}

	// Verify tenant was created
	tenants := mtps.ListTenants()
	if len(tenants) != 1 {
		t.Errorf("Expected 1 auto-created tenant, got %d", len(tenants))
	}
}

func TestMultiTenant_InvalidTenant(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Try to create tenant with empty ID
	err := mtps.CreateTenant("", DefaultTenantQuota())
	if err != ErrInvalidTenant {
		t.Errorf("Expected ErrInvalidTenant, got %v", err)
	}

	// Try to publish with empty tenant ID
	msg := Message{Data: "test"}
	err = mtps.Publish("", "test", msg)
	if err != ErrInvalidTenant {
		t.Errorf("Expected ErrInvalidTenant, got %v", err)
	}
}

func TestMultiTenant_SubscriptionLifecycle(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	_ = mtps.CreateTenant("tenant1", DefaultTenantQuota())

	// Create subscription
	sub, err := mtps.Subscribe("tenant1", "test", SubOptions{BufferSize: 10})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Check usage
	usage, _ := mtps.GetUsage("tenant1")
	if usage.ActiveSubs.Load() != 1 {
		t.Errorf("Expected 1 active subscription, got %d", usage.ActiveSubs.Load())
	}

	// Cancel subscription
	sub.Cancel()

	time.Sleep(100 * time.Millisecond)

	// Check usage updated
	usage, _ = mtps.GetUsage("tenant1")
	if usage.ActiveSubs.Load() != 0 {
		t.Errorf("Expected 0 active subscriptions after cancel, got %d", usage.ActiveSubs.Load())
	}
}

func TestMultiTenant_QuotaViolations(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Create tenant with strict quotas
	quota := DefaultTenantQuota()
	quota.MaxPublishRate = 1
	_ = mtps.CreateTenant("tenant1", quota)

	// Trigger quota violations
	for i := 0; i < 5; i++ {
		_ = mtps.Publish("tenant1", "test", Message{Data: i})
	}

	// Check violations recorded
	usage, _ := mtps.GetUsage("tenant1")
	if usage.QuotaViolations.Load() == 0 {
		t.Error("Expected quota violations to be recorded")
	}
}

func TestMultiTenant_UnlimitedQuota(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, UnlimitedQuota())
	defer mtps.Close()

	_ = mtps.CreateTenant("tenant1", UnlimitedQuota())

	// Should be able to publish many messages
	for i := 0; i < 100; i++ {
		msg := Message{Data: i}
		err := mtps.Publish("tenant1", "test", msg)
		if err != nil {
			t.Fatalf("Unexpected error with unlimited quota: %v", err)
		}
	}
}

func TestMultiTenant_Namespace(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())
	defer mtps.Close()

	// Test namespace functions
	namespaced := mtps.namespaceTopic("tenant1", "test.topic")
	expected := "_tenant_tenant1.test.topic"
	if namespaced != expected {
		t.Errorf("Expected %s, got %s", expected, namespaced)
	}

	stripped := mtps.stripNamespace("tenant1", namespaced)
	if stripped != "test.topic" {
		t.Errorf("Expected test.topic, got %s", stripped)
	}
}

func TestMultiTenant_Close(t *testing.T) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, DefaultTenantQuota())

	// Close
	err := mtps.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Double close should be safe
	err = mtps.Close()
	if err != nil {
		t.Fatalf("Double close failed: %v", err)
	}
}

func BenchmarkMultiTenant_Publish(b *testing.B) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, UnlimitedQuota())
	defer mtps.Close()

	_ = mtps.CreateTenant("bench-tenant", UnlimitedQuota())

	msg := Message{Data: "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mtps.Publish("bench-tenant", "test", msg)
	}
}

func BenchmarkMultiTenant_Subscribe(b *testing.B) {
	ps := New()
	defer ps.Close()

	mtps := NewMultiTenantPubSub(ps, UnlimitedQuota())
	defer mtps.Close()

	_ = mtps.CreateTenant("bench-tenant", UnlimitedQuota())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sub, _ := mtps.Subscribe("bench-tenant", "test", SubOptions{BufferSize: 10})
		sub.Cancel()
	}
}
