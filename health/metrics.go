package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

// HealthMetrics contains various health-related metrics.
type HealthMetrics struct {
	StartTime        time.Time                    `json:"start_time"`
	LastCheckTime    time.Time                    `json:"last_check_time"`
	CheckCount       int64                        `json:"check_count"`
	SuccessCount     int64                        `json:"success_count"`
	FailureCount     int64                        `json:"failure_count"`
	ComponentMetrics map[string]*ComponentMetrics `json:"component_metrics"`
}

// ComponentMetrics tracks metrics for individual components.
type ComponentMetrics struct {
	Name           string              `json:"name"`
	CheckCount     int64               `json:"check_count"`
	SuccessCount   int64               `json:"success_count"`
	FailureCount   int64               `json:"failure_count"`
	AverageLatency time.Duration       `json:"average_latency"`
	MinLatency     time.Duration       `json:"min_latency"`
	MaxLatency     time.Duration       `json:"max_latency"`
	LastCheckTime  time.Time           `json:"last_check_time"`
	LastStatus     HealthState         `json:"last_status"`
	RecentHistory  []HealthCheckRecord `json:"recent_history,omitempty"`
	HealthTrend    HealthTrend         `json:"health_trend,omitempty"`
}

// HealthCheckRecord represents a single health check execution record.
type HealthCheckRecord struct {
	Timestamp    time.Time     `json:"timestamp"`
	Duration     time.Duration `json:"duration"`
	Success      bool          `json:"success"`
	Status       HealthState   `json:"status"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// HealthTrend represents the health trend analysis.
type HealthTrend struct {
	Direction     TrendDirection `json:"direction"`      // improving/declining/stable
	Stability     float64        `json:"stability"`      // 0.0-1.0, higher is more stable
	RecentSuccess float64        `json:"recent_success"` // Recent success rate
	TrendScore    float64        `json:"trend_score"`    // -1.0 to 1.0, negative means deteriorating
}

// TrendDirection represents the direction of health trend.
type TrendDirection string

const (
	TrendImproving TrendDirection = "improving"
	TrendDeclining TrendDirection = "declining"
	TrendStable    TrendDirection = "stable"
)

// GetRecentSuccessRate calculates the success rate for recent checks.
func (cm *ComponentMetrics) GetRecentSuccessRate() float64 {
	if len(cm.RecentHistory) == 0 {
		return 0.0
	}

	successCount := 0
	for _, record := range cm.RecentHistory {
		if record.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(cm.RecentHistory))
}

// MetricsCollector collects health check metrics.
type MetricsCollector struct {
	mu        sync.RWMutex
	metrics   *HealthMetrics
	collector HealthManager
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector(manager HealthManager) *MetricsCollector {
	collector := &MetricsCollector{
		metrics: &HealthMetrics{
			StartTime:        time.Now(),
			ComponentMetrics: make(map[string]*ComponentMetrics),
		},
		collector: manager,
	}
	if manager != nil {
		_ = AttachMetrics(manager, collector)
	}
	return collector
}

// AttachMetrics attaches a metrics collector to a health manager when supported.
func AttachMetrics(manager HealthManager, collector *MetricsCollector) error {
	if manager == nil {
		return errors.New("manager cannot be nil")
	}
	if collector == nil {
		return errors.New("collector cannot be nil")
	}

	hm, ok := manager.(*healthManager)
	if !ok {
		return errors.New("manager does not support metrics attachment")
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()
	if hm.closed {
		return errManagerClosed
	}

	hm.metrics = collector
	collector.collector = manager
	return nil
}

// RecordCheck records a health check execution with enhanced metrics collection.
func (mc *MetricsCollector) RecordCheck(componentName string, duration time.Duration, success bool, status HealthState) {
	mc.RecordCheckWithError(componentName, duration, success, status, nil)
}

// RecordCheckWithError records a health check execution including error details.
func (mc *MetricsCollector) RecordCheckWithError(componentName string, duration time.Duration, success bool, status HealthState, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.CheckCount++
	mc.metrics.LastCheckTime = time.Now()

	if success {
		mc.metrics.SuccessCount++
	} else {
		mc.metrics.FailureCount++
	}

	// Update component metrics
	if _, exists := mc.metrics.ComponentMetrics[componentName]; !exists {
		mc.metrics.ComponentMetrics[componentName] = &ComponentMetrics{
			Name:          componentName,
			MinLatency:    duration,
			MaxLatency:    duration,
			LastStatus:    status,
			LastCheckTime: time.Now(),
			RecentHistory: make([]HealthCheckRecord, 0, 10), // Keep last 10 checks
		}
	}

	comp := mc.metrics.ComponentMetrics[componentName]
	comp.CheckCount++
	comp.LastCheckTime = time.Now()
	comp.LastStatus = status

	if success {
		comp.SuccessCount++
	} else {
		comp.FailureCount++
	}

	// Update latency metrics
	if comp.MinLatency == 0 || duration < comp.MinLatency {
		comp.MinLatency = duration
	}
	if duration > comp.MaxLatency {
		comp.MaxLatency = duration
	}

	// Calculate running average with improved precision
	if comp.CheckCount == 1 {
		comp.AverageLatency = duration
	} else {
		totalLatency := comp.AverageLatency * time.Duration(comp.CheckCount-1)
		comp.AverageLatency = (totalLatency + duration) / time.Duration(comp.CheckCount)
	}

	// Add to recent history (keep only last 10 records)
	record := HealthCheckRecord{
		Timestamp: time.Now(),
		Duration:  duration,
		Success:   success,
		Status:    status,
	}
	if err != nil {
		record.ErrorMessage = err.Error()
	}

	// Append new record
	comp.RecentHistory = append(comp.RecentHistory, record)
	if len(comp.RecentHistory) > 10 {
		comp.RecentHistory = comp.RecentHistory[1:] // Remove oldest
	}

	// Update health trend
	mc.updateHealthTrend(comp)
}

// updateHealthTrend calculates and updates the health trend for a component.
func (mc *MetricsCollector) updateHealthTrend(comp *ComponentMetrics) {
	if len(comp.RecentHistory) < 3 {
		comp.HealthTrend = HealthTrend{
			Direction:     TrendStable,
			Stability:     1.0,
			RecentSuccess: comp.GetRecentSuccessRate(),
			TrendScore:    0.0,
		}
		return
	}

	// Calculate recent success rate
	recentSuccess := comp.GetRecentSuccessRate()

	// Calculate trend by comparing recent vs older results
	var recentGood, olderGood int
	recentCount := len(comp.RecentHistory)
	halfPoint := recentCount / 2

	for i, record := range comp.RecentHistory {
		if i >= halfPoint {
			if record.Success {
				recentGood++
			}
		} else {
			if record.Success {
				olderGood++
			}
		}
	}

	recentRate := float64(recentGood) / float64(halfPoint)
	olderRate := float64(olderGood) / float64(recentCount-halfPoint)

	trendScore := recentRate - olderRate

	var direction TrendDirection
	if trendScore > 0.1 {
		direction = TrendImproving
	} else if trendScore < -0.1 {
		direction = TrendDeclining
	} else {
		direction = TrendStable
	}

	// Calculate stability (inverse of variance in success/failure)
	stability := mc.calculateStability(comp.RecentHistory)

	comp.HealthTrend = HealthTrend{
		Direction:     direction,
		Stability:     stability,
		RecentSuccess: recentSuccess,
		TrendScore:    trendScore,
	}
}

// calculateStability calculates the stability score based on recent history.
func (mc *MetricsCollector) calculateStability(history []HealthCheckRecord) float64 {
	if len(history) < 2 {
		return 1.0
	}

	successCount := 0
	for _, record := range history {
		if record.Success {
			successCount++
		}
	}

	// Higher stability when success rate is consistent (either high or low but stable)
	successRate := float64(successCount) / float64(len(history))
	if successRate >= 0.8 || successRate <= 0.2 {
		return 0.9 // Very stable (consistently good or bad)
	} else if successRate >= 0.6 || successRate <= 0.4 {
		return 0.7 // Moderately stable
	} else {
		return 0.3 // Unstable (mixed results)
	}
}

// GetMetrics returns the current health metrics.
func (mc *MetricsCollector) GetMetrics() HealthMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy to prevent external modification without copying the mutex.
	metricsCopy := HealthMetrics{
		StartTime:     mc.metrics.StartTime,
		LastCheckTime: mc.metrics.LastCheckTime,
		CheckCount:    mc.metrics.CheckCount,
		SuccessCount:  mc.metrics.SuccessCount,
		FailureCount:  mc.metrics.FailureCount,
	}
	metricsCopy.ComponentMetrics = make(map[string]*ComponentMetrics, len(mc.metrics.ComponentMetrics))
	for name, comp := range mc.metrics.ComponentMetrics {
		compCopy := *comp
		compCopy.RecentHistory = append([]HealthCheckRecord(nil), comp.RecentHistory...)
		metricsCopy.ComponentMetrics[name] = &compCopy
	}

	return metricsCopy
}

// GetComponentMetrics returns metrics for a specific component.
func (mc *MetricsCollector) GetComponentMetrics(componentName string) (*ComponentMetrics, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, exists := mc.metrics.ComponentMetrics[componentName]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	metricsCopy := *metrics
	metricsCopy.RecentHistory = append([]HealthCheckRecord(nil), metrics.RecentHistory...)
	return &metricsCopy, true
}

// GetSuccessRate returns the success rate of health checks as a percentage.
func (mc *MetricsCollector) GetSuccessRate() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.metrics.CheckCount == 0 {
		return 0.0
	}

	return float64(mc.metrics.SuccessCount) / float64(mc.metrics.CheckCount) * 100
}

// GetUptime returns the application uptime.
func (mc *MetricsCollector) GetUptime() time.Duration {
	mc.mu.RLock()
	start := mc.metrics.StartTime
	mc.mu.RUnlock()

	uptime := time.Since(start)
	if uptime <= 0 {
		return time.Nanosecond
	}
	return uptime
}

// Reset resets all metrics.
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = &HealthMetrics{
		StartTime:        time.Now(),
		ComponentMetrics: make(map[string]*ComponentMetrics),
	}
}

// HealthReport provides a comprehensive health report with metrics.
type HealthReport struct {
	HealthStatus HealthStatus    `json:"health_status"`
	Metrics      HealthMetrics   `json:"metrics"`
	Readiness    ReadinessStatus `json:"readiness"`
	BuildInfo    BuildInfo       `json:"build_info"`
	Components   []string        `json:"components"`
}

// GenerateReport generates a comprehensive health report.
func (mc *MetricsCollector) GenerateReport() HealthReport {
	overallHealth := mc.collector.GetOverallHealth()
	allHealth := mc.collector.GetAllHealth()
	buildInfo := GetBuildInfo()
	readiness := GetReadiness()

	components := make([]string, 0, len(allHealth))
	for name := range allHealth {
		components = append(components, name)
	}

	return HealthReport{
		HealthStatus: overallHealth,
		Metrics:      mc.GetMetrics(),
		Readiness:    readiness,
		BuildInfo:    buildInfo,
		Components:   components,
	}
}

// MetricsHandler creates a handler that exposes health metrics.
func MetricsHandler(collector *MetricsCollector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := collector.GetMetrics()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(metrics)
	})
}

// HealthReportHandler creates a handler that exposes comprehensive health report.
func HealthReportHandler(collector *MetricsCollector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		report := collector.GenerateReport()

		code := http.StatusOK
		switch report.HealthStatus.Status {
		case StatusUnhealthy:
			code = http.StatusServiceUnavailable
		case StatusDegraded:
			code = http.StatusPartialContent
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(report)
	})
}
