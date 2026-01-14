package metrics

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Dashboard provides a comprehensive performance monitoring dashboard
type Dashboard struct {
	mu             sync.RWMutex
	metrics        map[string]*MetricSeries
	startTime      time.Time
	config         DashboardConfig
	alerts         []Alert
	eventHistory   []Event
	maxHistorySize int
}

// DashboardConfig configures the dashboard behavior
type DashboardConfig struct {
	UpdateInterval   time.Duration `json:"update_interval"`
	MaxHistoryPoints int           `json:"max_history_points"`
	EnableAlerts     bool          `json:"enable_alerts"`
	EnableTracing    bool          `json:"enable_tracing"`
	EnableProfiling  bool          `json:"enable_profiling"`
}

// MetricSeries represents a time series of metric values
type MetricSeries struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Unit        string            `json:"unit"`
	Category    string            `json:"category"`
	Values      []MetricPoint     `json:"values"`
	Labels      map[string]string `json:"labels,omitempty"`
	Aggregation AggregationType   `json:"aggregation"`
}

// MetricPoint represents a single data point in a time series
type MetricPoint struct {
	Timestamp time.Time         `json:"timestamp"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// Alert represents a performance alert
type Alert struct {
	ID         string        `json:"id"`
	MetricName string        `json:"metric_name"`
	Condition  string        `json:"condition"`
	Threshold  float64       `json:"threshold"`
	Current    float64       `json:"current"`
	Severity   AlertSeverity `json:"severity"`
	Message    string        `json:"message"`
	Timestamp  time.Time     `json:"timestamp"`
	Active     bool          `json:"active"`
}

// Event represents a significant event in the system
type Event struct {
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// AggregationType defines how to aggregate metric values
type AggregationType string

const (
	AggregationSum    AggregationType = "sum"
	AggregationAvg    AggregationType = "avg"
	AggregationMax    AggregationType = "max"
	AggregationMin    AggregationType = "min"
	AggregationCount  AggregationType = "count"
	AggregationLatest AggregationType = "latest"
)

// AlertSeverity defines the severity level of an alert
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityError    AlertSeverity = "error"
	SeverityCritical AlertSeverity = "critical"
)

// NewDashboard creates a new performance dashboard
func NewDashboard(config DashboardConfig) *Dashboard {
	if config.UpdateInterval == 0 {
		config.UpdateInterval = 1 * time.Second
	}
	if config.MaxHistoryPoints == 0 {
		config.MaxHistoryPoints = 100
	}

	return &Dashboard{
		metrics:        make(map[string]*MetricSeries),
		startTime:      time.Now(),
		config:         config,
		alerts:         make([]Alert, 0),
		eventHistory:   make([]Event, 0),
		maxHistorySize: config.MaxHistoryPoints,
	}
}

// RegisterMetric registers a new metric series
func (d *Dashboard) RegisterMetric(name, description, unit, category string, labels map[string]string, agg AggregationType) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.metrics[name] = &MetricSeries{
		Name:        name,
		Description: description,
		Unit:        unit,
		Category:    category,
		Values:      make([]MetricPoint, 0),
		Labels:      labels,
		Aggregation: agg,
	}
}

// AddMetricValue adds a value to a metric series
func (d *Dashboard) AddMetricValue(name string, value float64, labels map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	metric, exists := d.metrics[name]
	if !exists {
		return
	}

	point := MetricPoint{
		Timestamp: time.Now(),
		Value:     value,
		Labels:    labels,
	}

	metric.Values = append(metric.Values, point)

	// Trim history if needed
	if len(metric.Values) > d.maxHistorySize {
		metric.Values = metric.Values[len(metric.Values)-d.maxHistorySize:]
	}

	// Check for alerts
	if d.config.EnableAlerts {
		d.checkAlerts(name, value, labels)
	}
}

// RecordEvent records a system event
func (d *Dashboard) RecordEvent(eventType, message string, metadata map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	event := Event{
		Timestamp: time.Now(),
		Type:      eventType,
		Message:   message,
		Metadata:  metadata,
	}

	d.eventHistory = append(d.eventHistory, event)

	// Trim event history
	if len(d.eventHistory) > 1000 {
		d.eventHistory = d.eventHistory[len(d.eventHistory)-1000:]
	}
}

// checkAlerts checks metric values against alert thresholds
func (d *Dashboard) checkAlerts(metricName string, value float64, labels map[string]string) {
	// Example alert rules - in practice these would be configurable
	switch metricName {
	case "http_request_duration":
		if value > 1.0 { // 1 second threshold
			d.addAlert(metricName, "Request duration too high", value, 1.0, SeverityWarning)
		}
	case "memory_usage":
		if value > 80.0 { // 80% threshold
			d.addAlert(metricName, "High memory usage", value, 80.0, SeverityError)
		}
	case "cpu_usage":
		if value > 90.0 { // 90% threshold
			d.addAlert(metricName, "Critical CPU usage", value, 90.0, SeverityCritical)
		}
	case "error_rate":
		if value > 5.0 { // 5% error rate
			d.addAlert(metricName, "High error rate", value, 5.0, SeverityError)
		}
	}
}

// addAlert creates and adds an alert
func (d *Dashboard) addAlert(metricName, message string, current, threshold float64, severity AlertSeverity) {
	alert := Alert{
		ID:         fmt.Sprintf("%s_%d", metricName, time.Now().UnixNano()),
		MetricName: metricName,
		Message:    message,
		Current:    current,
		Threshold:  threshold,
		Severity:   severity,
		Timestamp:  time.Now(),
		Active:     true,
	}

	// Check if similar alert already exists
	for i, existing := range d.alerts {
		if existing.MetricName == metricName && existing.Active {
			// Update existing alert
			d.alerts[i] = alert
			return
		}
	}

	d.alerts = append(d.alerts, alert)
}

// GetMetricStats returns statistics for a metric
func (d *Dashboard) GetMetricStats(name string) (MetricStats, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	metric, exists := d.metrics[name]
	if !exists {
		return MetricStats{}, fmt.Errorf("metric %s not found", name)
	}

	if len(metric.Values) == 0 {
		return MetricStats{}, fmt.Errorf("no data for metric %s", name)
	}

	return calculateStats(metric.Values, metric.Aggregation), nil
}

// GetDashboardSnapshot returns a complete snapshot of the dashboard state
func (d *Dashboard) GetDashboardSnapshot() DashboardSnapshot {
	d.mu.RLock()
	defer d.mu.RUnlock()

	snapshot := DashboardSnapshot{
		Timestamp:    time.Now(),
		Uptime:       time.Since(d.startTime),
		Metrics:      make(map[string]MetricSeries),
		Alerts:       make([]Alert, 0),
		Events:       make([]Event, 0),
		TotalMetrics: len(d.metrics),
	}

	// Copy metrics
	for name, metric := range d.metrics {
		// Create a copy to avoid data races
		metricCopy := *metric
		snapshot.Metrics[name] = metricCopy
	}

	// Copy active alerts
	for _, alert := range d.alerts {
		if alert.Active {
			snapshot.Alerts = append(snapshot.Alerts, alert)
		}
	}

	// Copy recent events
	if len(d.eventHistory) > 0 {
		start := len(d.eventHistory) - 50
		if start < 0 {
			start = 0
		}
		snapshot.Events = d.eventHistory[start:]
	}

	return snapshot
}

// GenerateReport generates a comprehensive performance report
func (d *Dashboard) GenerateReport() string {
	snapshot := d.GetDashboardSnapshot()

	var builder strings.Builder

	builder.WriteString("# Performance Dashboard Report\n\n")
	builder.WriteString(fmt.Sprintf("Generated: %s\n", snapshot.Timestamp.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("Uptime: %s\n\n", snapshot.Uptime.Round(time.Second)))

	// Summary
	builder.WriteString("## Summary\n\n")
	builder.WriteString(fmt.Sprintf("- Total Metrics: %d\n", snapshot.TotalMetrics))
	builder.WriteString(fmt.Sprintf("- Active Alerts: %d\n", len(snapshot.Alerts)))
	builder.WriteString(fmt.Sprintf("- Recent Events: %d\n\n", len(snapshot.Events)))

	// Metrics Overview
	builder.WriteString("## Metrics Overview\n\n")
	for name, metric := range snapshot.Metrics {
		stats, err := d.GetMetricStats(name)
		if err == nil {
			builder.WriteString(fmt.Sprintf("### %s\n", name))
			builder.WriteString(fmt.Sprintf("- Description: %s\n", metric.Description))
			builder.WriteString(fmt.Sprintf("- Unit: %s\n", metric.Unit))
			builder.WriteString(fmt.Sprintf("- Category: %s\n", metric.Category))
			builder.WriteString(fmt.Sprintf("- Latest Value: %.2f %s\n", stats.Latest, metric.Unit))
			builder.WriteString(fmt.Sprintf("- Average: %.2f %s\n", stats.Average, metric.Unit))
			builder.WriteString(fmt.Sprintf("- Max: %.2f %s\n", stats.Max, metric.Unit))
			builder.WriteString(fmt.Sprintf("- Min: %.2f %s\n", stats.Min, metric.Unit))
			builder.WriteString(fmt.Sprintf("- Count: %d\n\n", stats.Count))
		}
	}

	// Active Alerts
	if len(snapshot.Alerts) > 0 {
		builder.WriteString("## Active Alerts\n\n")
		for _, alert := range snapshot.Alerts {
			builder.WriteString(fmt.Sprintf("### [%s] %s\n", alert.Severity, alert.MetricName))
			builder.WriteString(fmt.Sprintf("- Message: %s\n", alert.Message))
			builder.WriteString(fmt.Sprintf("- Current: %.2f\n", alert.Current))
			builder.WriteString(fmt.Sprintf("- Threshold: %.2f\n", alert.Threshold))
			builder.WriteString(fmt.Sprintf("- Time: %s\n\n", alert.Timestamp.Format(time.RFC3339)))
		}
	}

	// Recent Events
	if len(snapshot.Events) > 0 {
		builder.WriteString("## Recent Events\n\n")
		for _, event := range snapshot.Events {
			builder.WriteString(fmt.Sprintf("- [%s] %s: %s\n",
				event.Timestamp.Format("15:04:05"), event.Type, event.Message))
		}
	}

	return builder.String()
}

// GenerateJSON generates a JSON representation of the dashboard
func (d *Dashboard) GenerateJSON() (string, error) {
	snapshot := d.GetDashboardSnapshot()
	jsonBytes, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// GenerateMermaid generates a Mermaid diagram of the metrics flow
func (d *Dashboard) GenerateMermaid() string {
	snapshot := d.GetDashboardSnapshot()

	var builder strings.Builder
	builder.WriteString("```mermaid\n")
	builder.WriteString("graph TD\n")
	builder.WriteString("    subgraph \"Performance Dashboard\"\n")

	// Metrics
	for name, metric := range snapshot.Metrics {
		nodeName := strings.ReplaceAll(name, " ", "_")
		stats, _ := d.GetMetricStats(name)
		label := fmt.Sprintf("%s\\n%.2f %s", name, stats.Latest, metric.Unit)
		builder.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeName, label))
	}

	// Alerts
	if len(snapshot.Alerts) > 0 {
		builder.WriteString("    subgraph \"Active Alerts\"\n")
		for _, alert := range snapshot.Alerts {
			alertNode := fmt.Sprintf("alert_%s", alert.ID)
			builder.WriteString(fmt.Sprintf("    %s[\"%s: %s\"]\n", alertNode, alert.Severity, alert.MetricName))
		}
		builder.WriteString("    end\n")
	}

	builder.WriteString("    end\n")
	builder.WriteString("```\n")

	return builder.String()
}

// DeactivateAlert deactivates an alert by ID
func (d *Dashboard) DeactivateAlert(alertID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, alert := range d.alerts {
		if alert.ID == alertID {
			d.alerts[i].Active = false
			return
		}
	}
}

// ClearMetrics clears all metric data (useful for testing)
func (d *Dashboard) ClearMetrics() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, metric := range d.metrics {
		metric.Values = make([]MetricPoint, 0)
	}
}

// ClearAlerts clears all alerts
func (d *Dashboard) ClearAlerts() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.alerts = make([]Alert, 0)
}

// GetActiveAlerts returns all active alerts
func (d *Dashboard) GetActiveAlerts() []Alert {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var active []Alert
	for _, alert := range d.alerts {
		if alert.Active {
			active = append(active, alert)
		}
	}
	return active
}

// MetricStats represents statistical information about a metric
type MetricStats struct {
	Name    string  `json:"name"`
	Latest  float64 `json:"latest"`
	Average float64 `json:"average"`
	Max     float64 `json:"max"`
	Min     float64 `json:"min"`
	Count   int     `json:"count"`
	Sum     float64 `json:"sum"`
}

// DashboardSnapshot represents a complete snapshot of the dashboard state
type DashboardSnapshot struct {
	Timestamp    time.Time               `json:"timestamp"`
	Uptime       time.Duration           `json:"uptime"`
	Metrics      map[string]MetricSeries `json:"metrics"`
	Alerts       []Alert                 `json:"alerts"`
	Events       []Event                 `json:"events"`
	TotalMetrics int                     `json:"total_metrics"`
}

// calculateStats computes statistics for a metric series
func calculateStats(values []MetricPoint, agg AggregationType) MetricStats {
	if len(values) == 0 {
		return MetricStats{}
	}

	var sum, max, min float64
	max = values[0].Value
	min = values[0].Value

	for _, point := range values {
		sum += point.Value
		if point.Value > max {
			max = point.Value
		}
		if point.Value < min {
			min = point.Value
		}
	}

	avg := sum / float64(len(values))
	latest := values[len(values)-1].Value

	return MetricStats{
		Name:    "",
		Latest:  latest,
		Average: avg,
		Max:     max,
		Min:     min,
		Count:   len(values),
		Sum:     sum,
	}
}

// DashboardIntegration provides integration with the existing metrics system
type DashboardIntegration struct {
	dashboard *Dashboard
	collector MetricsCollector
}

// NewDashboardIntegration creates a new integration
func NewDashboardIntegration(collector MetricsCollector, config DashboardConfig) *DashboardIntegration {
	dashboard := NewDashboard(config)
	return &DashboardIntegration{
		dashboard: dashboard,
		collector: collector,
	}
}

// StartCollection starts collecting metrics from the collector
func (di *DashboardIntegration) StartCollection() {
	// Register common metrics
	di.dashboard.RegisterMetric("http_request_duration", "HTTP request duration", "seconds", "performance", nil, AggregationAvg)
	di.dashboard.RegisterMetric("memory_usage", "Memory usage percentage", "percent", "resource", nil, AggregationMax)
	di.dashboard.RegisterMetric("cpu_usage", "CPU usage percentage", "percent", "resource", nil, AggregationMax)
	di.dashboard.RegisterMetric("error_rate", "Error rate", "percent", "reliability", nil, AggregationAvg)
	di.dashboard.RegisterMetric("goroutines", "Number of goroutines", "count", "resource", nil, AggregationLatest)

	// In a real implementation, this would periodically collect from the collector
	// For now, we provide manual collection methods
}

// CollectNow collects current metrics from the collector's stats
func (di *DashboardIntegration) CollectNow() {
	// Get stats from the collector
	stats := di.collector.GetStats()

	// Calculate error rate
	var errorRate float64
	if stats.TotalRecords > 0 {
		errorRate = float64(stats.ErrorRecords) / float64(stats.TotalRecords) * 100
	}

	// Add metrics to dashboard
	if errorRate > 0 {
		di.dashboard.AddMetricValue("error_rate", errorRate, nil)
	}

	// Add collector stats as metrics
	di.dashboard.AddMetricValue("collector_total_records", float64(stats.TotalRecords), nil)
	di.dashboard.AddMetricValue("collector_error_records", float64(stats.ErrorRecords), nil)
	di.dashboard.AddMetricValue("collector_active_series", float64(stats.ActiveSeries), nil)

	// Add tracing metrics if available
	if stats.TotalSpans > 0 {
		di.dashboard.AddMetricValue("tracing_total_spans", float64(stats.TotalSpans), nil)
		di.dashboard.AddMetricValue("tracing_error_spans", float64(stats.ErrorSpans), nil)
		if stats.AverageDuration > 0 {
			di.dashboard.AddMetricValue("tracing_avg_duration", float64(stats.AverageDuration.Milliseconds()), nil)
		}
	}

	// Add legacy metrics for backward compatibility
	if stats.TotalRequests > 0 {
		di.dashboard.AddMetricValue("http_total_requests", float64(stats.TotalRequests), nil)
	}
	if stats.AverageLatency > 0 {
		di.dashboard.AddMetricValue("http_avg_latency", stats.AverageLatency, nil)
	}

	// Record an event
	di.dashboard.RecordEvent("metrics_collection", "Collected metrics from collector", map[string]string{
		"total_records": fmt.Sprintf("%d", stats.TotalRecords),
		"error_records": fmt.Sprintf("%d", stats.ErrorRecords),
	})
}

// CollectFromRecord processes a single metric record
func (di *DashboardIntegration) CollectFromRecord(record MetricRecord) {
	// Map record to dashboard metrics
	switch record.Type {
	case MetricHTTPRequest:
		// HTTP request duration in seconds
		duration := record.Duration.Seconds()
		di.dashboard.AddMetricValue("http_request_duration", duration, record.Labels)
		
		// Check for errors
		if record.Error != nil {
			di.dashboard.AddMetricValue("error_rate", 1.0, record.Labels)
		}

	case MetricPubSubPublish, MetricPubSubSubscribe, MetricPubSubDeliver, MetricPubSubDrop:
		// PubSub metrics
		duration := record.Duration.Seconds()
		di.dashboard.AddMetricValue("pubsub_duration", duration, record.Labels)
		if record.Error != nil {
			di.dashboard.AddMetricValue("pubsub_errors", 1.0, record.Labels)
		}

	case MetricKVGet, MetricKVSet, MetricKVDelete:
		// KV store metrics
		duration := record.Duration.Seconds()
		di.dashboard.AddMetricValue("kv_duration", duration, record.Labels)
		if record.Error != nil {
			di.dashboard.AddMetricValue("kv_errors", 1.0, record.Labels)
		}
	}
}

// GetDashboard returns the underlying dashboard
func (di *DashboardIntegration) GetDashboard() *Dashboard {
	return di.dashboard
}

// CollectMetricsBatch processes multiple metric records at once
func (di *DashboardIntegration) CollectMetricsBatch(records []MetricRecord) {
	for _, record := range records {
		di.CollectFromRecord(record)
	}
}
