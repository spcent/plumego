package pubsub

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPrometheusExporter_Basic(t *testing.T) {
	ps := New()
	defer ps.Close()

	exporter := NewPrometheusExporter(ps)

	// Publish some messages
	for i := 0; i < 10; i++ {
		msg := Message{Data: i}
		_ = ps.Publish("test.topic", msg)
	}

	// Collect metrics
	metrics := exporter.Collect()

	if metrics == "" {
		t.Error("Expected non-empty metrics")
	}

	// Should contain basic metrics
	if !strings.Contains(metrics, "messages_published_total") {
		t.Error("Expected messages_published_total metric")
	}
}

func TestPrometheusExporter_WithSubscribers(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Create subscribers
	sub1, _ := ps.Subscribe("test.a", SubOptions{BufferSize: 10})
	defer sub1.Cancel()

	sub2, _ := ps.Subscribe("test.a", SubOptions{BufferSize: 10})
	defer sub2.Cancel()

	sub3, _ := ps.Subscribe("test.b", SubOptions{BufferSize: 10})
	defer sub3.Cancel()

	exporter := NewPrometheusExporter(ps)
	metrics := exporter.Collect()

	if !strings.Contains(metrics, "subscribers_current") {
		t.Error("Expected subscribers_current metric")
	}

	// Should show 2 subscribers for test.a
	if !strings.Contains(metrics, `topic="test.a"`) {
		t.Error("Expected topic label")
	}
}

func TestPrometheusExporter_WithPersistence(t *testing.T) {
	dir := t.TempDir()

	config := PersistenceConfig{
		Enabled:           true,
		DataDir:           dir,
		DefaultDurability: DurabilityAsync,
	}

	pps, err := NewPersistent(config)
	if err != nil {
		t.Fatalf("Failed to create persistent pubsub: %v", err)
	}
	defer pps.Close()

	// Publish messages
	for i := 0; i < 5; i++ {
		msg := Message{Data: i}
		_ = pps.Publish("test", msg)
	}

	time.Sleep(100 * time.Millisecond)

	exporter := NewPrometheusExporter(pps.InProcPubSub).
		WithPersistent(pps)

	metrics := exporter.Collect()

	// Should have persistence metrics
	if !strings.Contains(metrics, "persistence_wal_writes_total") {
		t.Error("Expected persistence_wal_writes_total metric")
	}
}

func TestPrometheusExporter_WithDistributed(t *testing.T) {
	config := DefaultClusterConfig("test-node", "127.0.0.1:19001")

	dps, err := NewDistributed(config)
	if err != nil {
		t.Fatalf("Failed to create distributed pubsub: %v", err)
	}
	defer dps.Close()

	_ = dps.JoinCluster(testContext(t))
	time.Sleep(100 * time.Millisecond)

	exporter := NewPrometheusExporter(dps.InProcPubSub).
		WithDistributed(dps)

	metrics := exporter.Collect()

	// Should have cluster metrics
	if !strings.Contains(metrics, "cluster_nodes_total") {
		t.Error("Expected cluster_nodes_total metric")
	}
}

func TestPrometheusExporter_WithOrdering(t *testing.T) {
	config := DefaultOrderingConfig()
	ops := NewOrdered(config)
	defer ops.Close()

	// Publish ordered messages
	for i := 0; i < 10; i++ {
		msg := Message{Data: i}
		_ = ops.PublishOrdered("test", msg, OrderPerTopic)
	}

	time.Sleep(100 * time.Millisecond)

	exporter := NewPrometheusExporter(ops.InProcPubSub).
		WithOrdered(ops)

	metrics := exporter.Collect()

	// Should have ordering metrics
	if !strings.Contains(metrics, "ordering_publishes_total") {
		t.Error("Expected ordering_publishes_total metric")
	}
}

func TestPrometheusExporter_WithRateLimit(t *testing.T) {
	config := DefaultRateLimitConfig()
	config.GlobalQPS = 100

	rlps, err := NewRateLimited(config)
	if err != nil {
		t.Fatalf("Failed to create rate-limited pubsub: %v", err)
	}
	defer rlps.Close()

	// Publish messages
	for i := 0; i < 20; i++ {
		msg := Message{Data: i}
		_ = rlps.Publish("test", msg)
	}

	exporter := NewPrometheusExporter(rlps.InProcPubSub).
		WithRateLimited(rlps)

	metrics := exporter.Collect()

	// Should have rate limit metrics
	if !strings.Contains(metrics, "ratelimit_global_allowed_total") {
		t.Error("Expected ratelimit_global_allowed_total metric")
	}
}

func TestPrometheusExporter_HTTPHandler(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Publish some messages
	for i := 0; i < 5; i++ {
		msg := Message{Data: i}
		_ = ps.Publish("test", msg)
	}

	exporter := NewPrometheusExporter(ps)

	// Create test server
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler := exporter.Handler()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Expected text/plain content type, got %s", contentType)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}

	// Should have Prometheus format markers
	if !strings.Contains(body, "# HELP") {
		t.Error("Expected # HELP in output")
	}
	if !strings.Contains(body, "# TYPE") {
		t.Error("Expected # TYPE in output")
	}
}

func TestPrometheusExporter_CustomNamespace(t *testing.T) {
	ps := New()
	defer ps.Close()

	exporter := NewPrometheusExporter(ps).
		WithNamespace("myapp")

	metrics := exporter.Collect()

	// Metrics should have custom namespace
	if !strings.Contains(metrics, "myapp_pubsub_") {
		t.Error("Expected custom namespace in metrics")
	}
}

func TestPrometheusExporter_WithLabels(t *testing.T) {
	ps := New()
	defer ps.Close()

	exporter := NewPrometheusExporter(ps).
		WithLabels(map[string]string{
			"environment": "production",
			"region":      "us-west-2",
		})

	// Publish message
	_ = ps.Publish("test", Message{Data: "test"})

	metrics := exporter.Collect()

	// Should have global labels
	if !strings.Contains(metrics, `environment="production"`) {
		t.Error("Expected environment label")
	}
	if !strings.Contains(metrics, `region="us-west-2"`) {
		t.Error("Expected region label")
	}
}

func TestPrometheusExporter_MetricFormatting(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Create multiple topics
	topics := []string{"topic.a", "topic.b", "topic.c"}
	for _, topic := range topics {
		for i := 0; i < 5; i++ {
			_ = ps.Publish(topic, Message{Data: i})
		}
	}

	exporter := NewPrometheusExporter(ps)
	metrics := exporter.Collect()

	lines := strings.Split(metrics, "\n")

	// Should have proper Prometheus format
	hasHelp := false
	hasType := false
	hasMetric := false

	for _, line := range lines {
		if strings.HasPrefix(line, "# HELP") {
			hasHelp = true
		}
		if strings.HasPrefix(line, "# TYPE") {
			hasType = true
		}
		if strings.Contains(line, "messages_published_total") && !strings.HasPrefix(line, "#") {
			hasMetric = true

			// Verify format: metric_name{labels} value
			if !strings.Contains(line, "{") || !strings.Contains(line, "}") {
				t.Errorf("Invalid metric format: %s", line)
			}
		}
	}

	if !hasHelp {
		t.Error("Missing # HELP lines")
	}
	if !hasType {
		t.Error("Missing # TYPE lines")
	}
	if !hasMetric {
		t.Error("Missing metric lines")
	}
}

func TestPrometheusExporter_LabelEscaping(t *testing.T) {
	// Test label value escaping
	escaped := escapeLabel(`test"value\with\newline` + "\n")

	if !strings.Contains(escaped, `\"`) {
		t.Error("Should escape quotes")
	}
	if !strings.Contains(escaped, `\\`) {
		t.Error("Should escape backslashes")
	}
	if !strings.Contains(escaped, `\n`) {
		t.Error("Should escape newlines")
	}
}

func TestPrometheusExporter_MultipleMetricTypes(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, _ := ps.Subscribe("test", SubOptions{BufferSize: 10})
	defer sub.Cancel()

	for i := 0; i < 3; i++ {
		_ = ps.Publish("test", Message{Data: i})
	}

	exporter := NewPrometheusExporter(ps)
	metrics := exporter.Collect()

	// Should have both counters and gauges
	hasCounter := strings.Contains(metrics, "# TYPE plumego_pubsub_messages_published_total counter")
	hasGauge := strings.Contains(metrics, "# TYPE plumego_pubsub_subscribers_current gauge")

	if !hasCounter {
		t.Error("Expected counter metric type")
	}
	if !hasGauge {
		t.Error("Expected gauge metric type")
	}
}

func TestPrometheusExporter_AllFeatures(t *testing.T) {
	// Create a fully featured setup
	dir := t.TempDir()

	// Base pubsub
	ps := New()
	defer ps.Close()

	// Add persistence
	pconfig := PersistenceConfig{
		Enabled:           true,
		DataDir:           dir,
		DefaultDurability: DurabilityAsync,
	}
	pps, _ := NewPersistent(pconfig, WithShardCount(16))
	defer pps.Close()

	// Add rate limiting
	rlconfig := DefaultRateLimitConfig()
	rlconfig.GlobalQPS = 100
	rlps, _ := NewRateLimited(rlconfig)
	defer rlps.Close()

	// Publish to all
	for i := 0; i < 10; i++ {
		msg := Message{Data: i}
		_ = ps.Publish("test.basic", msg)
		_ = pps.Publish("test.persistent", msg)
		_ = rlps.Publish("test.ratelimited", msg)
	}

	time.Sleep(200 * time.Millisecond)

	// Export all metrics
	exporter := NewPrometheusExporter(ps).
		WithPersistent(pps).
		WithRateLimited(rlps).
		WithNamespace("integration").
		WithLabels(map[string]string{"app": "test"})

	metrics := exporter.Collect()

	// Should have metrics from all sources
	checks := []string{
		"integration_pubsub_messages_published_total",
		"integration_pubsub_persistence_wal_writes_total",
		"integration_pubsub_ratelimit_global_allowed_total",
		`app="test"`,
	}

	for _, check := range checks {
		if !strings.Contains(metrics, check) {
			t.Errorf("Missing expected content: %s", check)
		}
	}
}

func TestPrometheusExporter_HTTPMethodValidation(t *testing.T) {
	ps := New()
	defer ps.Close()

	exporter := NewPrometheusExporter(ps)

	// Test POST (should fail)
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	w := httptest.NewRecorder()

	handler := exporter.Handler()
	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Code)
	}
}

func TestPrometheusExporter_ConsistentOutput(t *testing.T) {
	ps := New()
	defer ps.Close()

	_ = ps.Publish("test", Message{Data: "test"})

	exporter := NewPrometheusExporter(ps)

	// Collect twice
	metrics1 := exporter.Collect()
	metrics2 := exporter.Collect()

	// Output should be consistent (same metrics, same order)
	if metrics1 != metrics2 {
		t.Error("Metrics output should be consistent across calls")
	}
}

// Helper function for creating test context
func testContext(t *testing.T) testContextImpl {
	return testContextImpl{t: t}
}

type testContextImpl struct {
	t *testing.T
}

func (tc testContextImpl) Deadline() (time.Time, bool) { return time.Time{}, false }
func (tc testContextImpl) Done() <-chan struct{}        { return nil }
func (tc testContextImpl) Err() error                   { return nil }
func (tc testContextImpl) Value(key any) any            { return nil }
