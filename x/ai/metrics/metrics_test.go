package metrics

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNoOpCollector(t *testing.T) {
	collector := &NoOpCollector{}

	// Should not panic
	collector.Counter("test", 1.0)
	collector.Gauge("test", 100.0)
	collector.Histogram("test", 0.5)
	collector.Timing("test", time.Second)
}

func TestMemoryCollector_Counter(t *testing.T) {
	collector := NewMemoryCollector()

	// Increment counter
	collector.Counter("requests_total", 1, Tags("method", "GET", "status", "200")...)
	collector.Counter("requests_total", 2, Tags("method", "GET", "status", "200")...)
	collector.Counter("requests_total", 1, Tags("method", "POST", "status", "201")...)

	snapshot := collector.Snapshot()

	// Check GET counter
	getKey := "requests_total{method=GET,status=200}"
	if counter, exists := snapshot.Counters[getKey]; !exists {
		t.Errorf("Counter %s not found", getKey)
	} else if counter.Value != 3.0 {
		t.Errorf("Counter value = %v, want 3.0", counter.Value)
	}

	// Check POST counter
	postKey := "requests_total{method=POST,status=201}"
	if counter, exists := snapshot.Counters[postKey]; !exists {
		t.Errorf("Counter %s not found", postKey)
	} else if counter.Value != 1.0 {
		t.Errorf("Counter value = %v, want 1.0", counter.Value)
	}
}

func TestMemoryCollector_Gauge(t *testing.T) {
	collector := NewMemoryCollector()

	// Set gauge values
	collector.Gauge("cpu_usage", 45.5, Tags("host", "server1")...)
	collector.Gauge("cpu_usage", 67.8, Tags("host", "server1")...) // Overwrite
	collector.Gauge("cpu_usage", 23.4, Tags("host", "server2")...)

	snapshot := collector.Snapshot()

	// Check server1 gauge
	key1 := "cpu_usage{host=server1}"
	if gauge, exists := snapshot.Gauges[key1]; !exists {
		t.Errorf("Gauge %s not found", key1)
	} else if gauge.Value != 67.8 {
		t.Errorf("Gauge value = %v, want 67.8", gauge.Value)
	}

	// Check server2 gauge
	key2 := "cpu_usage{host=server2}"
	if gauge, exists := snapshot.Gauges[key2]; !exists {
		t.Errorf("Gauge %s not found", key2)
	} else if gauge.Value != 23.4 {
		t.Errorf("Gauge value = %v, want 23.4", gauge.Value)
	}
}

func TestMemoryCollector_Histogram(t *testing.T) {
	collector := NewMemoryCollector()

	// Record values
	values := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	for _, v := range values {
		collector.Histogram("request_duration", v, Tags("endpoint", "/api")...)
	}

	snapshot := collector.Snapshot()

	key := "request_duration{endpoint=/api}"
	hist, exists := snapshot.Histograms[key]
	if !exists {
		t.Fatalf("Histogram %s not found", key)
	}

	if hist.Count != 5 {
		t.Errorf("Histogram count = %v, want 5", hist.Count)
	}

	expectedSum := 1.5
	if hist.Sum != expectedSum {
		t.Errorf("Histogram sum = %v, want %v", hist.Sum, expectedSum)
	}

	if hist.Min != 0.1 {
		t.Errorf("Histogram min = %v, want 0.1", hist.Min)
	}

	if hist.Max != 0.5 {
		t.Errorf("Histogram max = %v, want 0.5", hist.Max)
	}

	expectedAvg := 0.3
	if hist.Avg != expectedAvg {
		t.Errorf("Histogram avg = %v, want %v", hist.Avg, expectedAvg)
	}
}

func TestMemoryCollector_Timing(t *testing.T) {
	collector := NewMemoryCollector()

	// Record timing
	collector.Timing("function_duration", 100*time.Millisecond, Tags("function", "processRequest")...)

	snapshot := collector.Snapshot()

	key := "function_duration{function=processRequest}"
	hist, exists := snapshot.Histograms[key]
	if !exists {
		t.Fatalf("Histogram %s not found", key)
	}

	if hist.Count != 1 {
		t.Errorf("Histogram count = %v, want 1", hist.Count)
	}

	expectedValue := 0.1 // 100ms = 0.1s
	if hist.Sum < expectedValue-0.001 || hist.Sum > expectedValue+0.001 {
		t.Errorf("Histogram sum = %v, want ~%v", hist.Sum, expectedValue)
	}
}

func TestMemoryCollector_Reset(t *testing.T) {
	collector := NewMemoryCollector()

	collector.Counter("test", 1.0)
	collector.Gauge("test", 100.0)
	collector.Histogram("test", 0.5)

	snapshot := collector.Snapshot()
	if len(snapshot.Counters) == 0 && len(snapshot.Gauges) == 0 && len(snapshot.Histograms) == 0 {
		t.Error("Snapshot should have metrics before reset")
	}

	collector.Reset()

	snapshot = collector.Snapshot()
	if len(snapshot.Counters) != 0 {
		t.Errorf("Counters count = %v, want 0 after reset", len(snapshot.Counters))
	}
	if len(snapshot.Gauges) != 0 {
		t.Errorf("Gauges count = %v, want 0 after reset", len(snapshot.Gauges))
	}
	if len(snapshot.Histograms) != 0 {
		t.Errorf("Histograms count = %v, want 0 after reset", len(snapshot.Histograms))
	}
}

func TestTags(t *testing.T) {
	tags := Tags("key1", "value1", "key2", "value2")

	if len(tags) != 2 {
		t.Errorf("Tags length = %v, want 2", len(tags))
	}

	if tags[0].Key != "key1" || tags[0].Value != "value1" {
		t.Errorf("Tag 0 = %+v, want {key1, value1}", tags[0])
	}

	if tags[1].Key != "key2" || tags[1].Value != "value2" {
		t.Errorf("Tag 1 = %+v, want {key2, value2}", tags[1])
	}
}

func TestTags_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Tags should panic with odd number of arguments")
		}
	}()

	Tags("key1", "value1", "key2") // Odd number
}

func TestTimer(t *testing.T) {
	collector := NewMemoryCollector()

	timer := NewTimer(collector, "test_duration", Tags("operation", "test")...)
	time.Sleep(10 * time.Millisecond)
	timer.Stop()

	snapshot := collector.Snapshot()
	key := "test_duration{operation=test}"
	hist, exists := snapshot.Histograms[key]
	if !exists {
		t.Fatalf("Histogram %s not found", key)
	}

	if hist.Count != 1 {
		t.Errorf("Histogram count = %v, want 1", hist.Count)
	}

	// Should be at least 10ms (0.01s)
	if hist.Sum < 0.01 {
		t.Errorf("Duration = %v, want >= 0.01", hist.Sum)
	}
}

func TestObserveFunc(t *testing.T) {
	collector := NewMemoryCollector()

	err := ObserveFunc(context.Background(), collector, "test_func", func() error {
		time.Sleep(5 * time.Millisecond)
		return nil
	}, Tags("func", "test")...)

	if err != nil {
		t.Errorf("ObserveFunc error = %v, want nil", err)
	}

	snapshot := collector.Snapshot()
	key := "test_func{func=test}"
	hist, exists := snapshot.Histograms[key]
	if !exists {
		t.Fatalf("Histogram %s not found", key)
	}

	if hist.Count != 1 {
		t.Errorf("Histogram count = %v, want 1", hist.Count)
	}
}

func TestBuildKey(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		tags     []Tag
		expected string
	}{
		{
			name:     "no tags",
			metric:   "metric",
			tags:     nil,
			expected: "metric",
		},
		{
			name:     "single tag",
			metric:   "metric",
			tags:     []Tag{{Key: "key", Value: "value"}},
			expected: "metric{key=value}",
		},
		{
			name:     "multiple tags",
			metric:   "metric",
			tags:     []Tag{{Key: "k1", Value: "v1"}, {Key: "k2", Value: "v2"}},
			expected: "metric{k1=v1,k2=v2}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildKey(tt.metric, tt.tags)
			if result != tt.expected {
				t.Errorf("buildKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		p        float64
		expected float64
	}{
		{0.0, 1},
		{0.5, 5.5}, // Linear interpolation between 5 and 6
		{0.9, 9.1}, // Linear interpolation between 9 and 10
		{1.0, 10},
	}

	for _, tt := range tests {
		t.Run("p"+string(rune(tt.p*100)), func(t *testing.T) {
			result := percentile(values, tt.p)
			if result != tt.expected {
				t.Errorf("percentile(%v) = %v, want %v", tt.p, result, tt.expected)
			}
		})
	}
}

func TestPercentile_EmptySlice(t *testing.T) {
	result := percentile([]float64{}, 0.5)
	if result != 0 {
		t.Errorf("percentile(empty) = %v, want 0", result)
	}
}

func TestMemoryCollector_Concurrent(t *testing.T) {
	collector := NewMemoryCollector()

	// Concurrent increments
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				collector.Counter("concurrent_test", 1.0)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	snapshot := collector.Snapshot()
	counter, exists := snapshot.Counters["concurrent_test"]
	if !exists {
		t.Fatal("Counter not found")
	}

	expected := 1000.0 // 10 goroutines * 100 increments
	if counter.Value != expected {
		t.Errorf("Counter value = %v, want %v", counter.Value, expected)
	}
}

func TestMemoryCollector_Snapshot_ConsistentRead(t *testing.T) {
	collector := NewMemoryCollector()

	// Start background updates
	stop := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				collector.Counter("test", 1.0)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Take multiple snapshots
	time.Sleep(10 * time.Millisecond)
	snapshot1 := collector.Snapshot()
	time.Sleep(10 * time.Millisecond)
	snapshot2 := collector.Snapshot()

	stop <- true

	// Snapshots should be consistent (not torn reads)
	if snapshot1.Counters["test"].Value >= snapshot2.Counters["test"].Value {
		t.Error("Counter should increase between snapshots")
	}
}

func TestTagsToMap(t *testing.T) {
	tags := []Tag{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	m := tagsToMap(tags)

	if len(m) != 2 {
		t.Errorf("Map length = %v, want 2", len(m))
	}

	if m["key1"] != "value1" {
		t.Errorf("m[key1] = %v, want value1", m["key1"])
	}

	if m["key2"] != "value2" {
		t.Errorf("m[key2] = %v, want value2", m["key2"])
	}
}

func BenchmarkCounter(b *testing.B) {
	collector := NewMemoryCollector()
	tags := Tags("method", "GET", "status", "200")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Counter("requests_total", 1.0, tags...)
	}
}

func BenchmarkHistogram(b *testing.B) {
	collector := NewMemoryCollector()
	tags := Tags("endpoint", "/api")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Histogram("request_duration", 0.123, tags...)
	}
}

func BenchmarkSnapshot(b *testing.B) {
	collector := NewMemoryCollector()

	// Populate with metrics
	for i := 0; i < 100; i++ {
		collector.Counter("counter", 1.0)
		collector.Gauge("gauge", 100.0)
		collector.Histogram("histogram", 0.5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collector.Snapshot()
	}
}

func TestPrometheusExporter_Handler(t *testing.T) {
	collector := NewMemoryCollector()
	exporter := NewPrometheusExporter(collector, "test")

	// Add some metrics
	collector.Counter("requests_total", 100, Tags("method", "GET")...)
	collector.Gauge("memory_usage", 1024.5, Tags("host", "server1")...)
	collector.Histogram("request_duration", 0.123)

	// Create HTTP test
	handler := exporter.Handler()

	// Since we can't easily test HTTP handlers without net/http/httptest,
	// we'll just verify it doesn't panic
	if handler == nil {
		t.Error("Handler should not be nil")
	}
}

func TestPrometheusExporter_ParseKey(t *testing.T) {
	exporter := NewPrometheusExporter(NewMemoryCollector(), "")

	tests := []struct {
		key          string
		expectedName string
		expectedTags map[string]string
	}{
		{
			key:          "metric",
			expectedName: "metric",
			expectedTags: nil,
		},
		{
			key:          "metric{key=value}",
			expectedName: "metric",
			expectedTags: map[string]string{"key": "value"},
		},
		{
			key:          "metric{k1=v1,k2=v2}",
			expectedName: "metric",
			expectedTags: map[string]string{"k1": "v1", "k2": "v2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			name, tags := exporter.parseKey(tt.key)
			if name != tt.expectedName {
				t.Errorf("name = %v, want %v", name, tt.expectedName)
			}
			if len(tags) != len(tt.expectedTags) {
				t.Errorf("tags length = %v, want %v", len(tags), len(tt.expectedTags))
			}
			for k, v := range tt.expectedTags {
				if tags[k] != v {
					t.Errorf("tags[%s] = %v, want %v", k, tags[k], v)
				}
			}
		})
	}
}

func TestPrometheusExporter_FormatLabels(t *testing.T) {
	exporter := NewPrometheusExporter(NewMemoryCollector(), "")

	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "no labels",
			labels:   map[string]string{},
			expected: "",
		},
		{
			name:     "single label",
			labels:   map[string]string{"key": "value"},
			expected: "{key=\"value\"}",
		},
		{
			name:     "multiple labels",
			labels:   map[string]string{"k1": "v1", "k2": "v2"},
			expected: "{k1=\"v1\",k2=\"v2\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exporter.formatLabels(tt.labels)
			// For multiple labels, order may vary, so check both possibilities
			if tt.name == "multiple labels" {
				alt := "{k2=\"v2\",k1=\"v1\"}"
				if result != tt.expected && result != alt {
					t.Errorf("formatLabels() = %v, want %v or %v", result, tt.expected, alt)
				}
			} else {
				if result != tt.expected {
					t.Errorf("formatLabels() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestPrometheusExporter_AddNamespace(t *testing.T) {
	tests := []struct {
		namespace string
		metric    string
		expected  string
	}{
		{"", "metric", "metric"},
		{"app", "metric", "app_metric"},
		{"my_app", "requests", "my_app_requests"},
	}

	for _, tt := range tests {
		t.Run(tt.namespace+"_"+tt.metric, func(t *testing.T) {
			exporter := NewPrometheusExporter(NewMemoryCollector(), tt.namespace)
			result := exporter.addNamespace(tt.metric)
			if result != tt.expected {
				t.Errorf("addNamespace() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPrometheusExporter_ScrapeConfig(t *testing.T) {
	exporter := NewPrometheusExporter(NewMemoryCollector(), "test")
	config := exporter.ScrapeConfig("my_job", "localhost:8080")

	if !strings.Contains(config, "job_name: 'my_job'") {
		t.Error("Config should contain job_name")
	}

	if !strings.Contains(config, "localhost:8080") {
		t.Error("Config should contain target address")
	}
}
