package health

import (
	"errors"
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
	Direction     TrendDirection `json:"direction"`
	Stability     float64        `json:"stability"`
	RecentSuccess float64        `json:"recent_success"`
	TrendScore    float64        `json:"trend_score"`
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

// MetricsTracker records and aggregates health check metrics over time.
type MetricsTracker struct {
	mu        sync.RWMutex
	metrics   *HealthMetrics
	collector HealthChecker
}

// NewMetricsTracker creates a new health metrics tracker.
func NewMetricsTracker(manager HealthManager) *MetricsTracker {
	tracker := &MetricsTracker{
		metrics: &HealthMetrics{
			StartTime:        time.Now(),
			ComponentMetrics: make(map[string]*ComponentMetrics),
		},
	}
	if manager != nil {
		_ = AttachMetricsTracker(manager, tracker)
	}
	return tracker
}

// AttachMetricsTracker attaches a MetricsTracker to a HealthManager.
func AttachMetricsTracker(manager HealthManager, tracker *MetricsTracker) error {
	if manager == nil {
		return errors.New("manager cannot be nil")
	}
	if tracker == nil {
		return errors.New("tracker cannot be nil")
	}

	manager.SetMetricsRecorder(tracker)
	tracker.collector = manager
	return nil
}

// RecordCheck records a health check execution.
func (mc *MetricsTracker) RecordCheck(componentName string, duration time.Duration, success bool, status HealthState) {
	mc.RecordCheckWithError(componentName, duration, success, status, nil)
}

// RecordCheckWithError records a health check execution including error details.
func (mc *MetricsTracker) RecordCheckWithError(componentName string, duration time.Duration, success bool, status HealthState, err error) {
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
			RecentHistory: make([]HealthCheckRecord, 0, 10),
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

	if comp.MinLatency == 0 || duration < comp.MinLatency {
		comp.MinLatency = duration
	}
	if duration > comp.MaxLatency {
		comp.MaxLatency = duration
	}

	if comp.CheckCount == 1 {
		comp.AverageLatency = duration
	} else {
		totalLatency := comp.AverageLatency * time.Duration(comp.CheckCount-1)
		comp.AverageLatency = (totalLatency + duration) / time.Duration(comp.CheckCount)
	}

	record := HealthCheckRecord{
		Timestamp: time.Now(),
		Duration:  duration,
		Success:   success,
		Status:    status,
	}
	if err != nil {
		record.ErrorMessage = err.Error()
	}

	comp.RecentHistory = append(comp.RecentHistory, record)
	if len(comp.RecentHistory) > 10 {
		comp.RecentHistory = comp.RecentHistory[1:]
	}

	mc.updateHealthTrend(comp)
}

// updateHealthTrend calculates and updates the health trend for a component.
func (mc *MetricsTracker) updateHealthTrend(comp *ComponentMetrics) {
	if len(comp.RecentHistory) < 3 {
		comp.HealthTrend = HealthTrend{
			Direction:     TrendStable,
			Stability:     1.0,
			RecentSuccess: comp.GetRecentSuccessRate(),
			TrendScore:    0.0,
		}
		return
	}

	recentSuccess := comp.GetRecentSuccessRate()

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

	stability := mc.calculateStability(comp.RecentHistory)

	comp.HealthTrend = HealthTrend{
		Direction:     direction,
		Stability:     stability,
		RecentSuccess: recentSuccess,
		TrendScore:    trendScore,
	}
}

// calculateStability calculates the stability score based on recent history.
func (mc *MetricsTracker) calculateStability(history []HealthCheckRecord) float64 {
	if len(history) < 2 {
		return 1.0
	}

	successCount := 0
	for _, record := range history {
		if record.Success {
			successCount++
		}
	}

	successRate := float64(successCount) / float64(len(history))
	if successRate >= 0.8 || successRate <= 0.2 {
		return 0.9
	} else if successRate >= 0.6 || successRate <= 0.4 {
		return 0.7
	} else {
		return 0.3
	}
}

// GetMetrics returns the current health metrics.
func (mc *MetricsTracker) GetMetrics() HealthMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

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
func (mc *MetricsTracker) GetComponentMetrics(componentName string) (*ComponentMetrics, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, exists := mc.metrics.ComponentMetrics[componentName]
	if !exists {
		return nil, false
	}

	metricsCopy := *metrics
	metricsCopy.RecentHistory = append([]HealthCheckRecord(nil), metrics.RecentHistory...)
	return &metricsCopy, true
}

// GetSuccessRate returns the success rate of health checks as a percentage.
func (mc *MetricsTracker) GetSuccessRate() float64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.metrics.CheckCount == 0 {
		return 0.0
	}

	return float64(mc.metrics.SuccessCount) / float64(mc.metrics.CheckCount) * 100
}

// GetUptime returns the application uptime.
func (mc *MetricsTracker) GetUptime() time.Duration {
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
func (mc *MetricsTracker) Reset() {
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
func (mc *MetricsTracker) GenerateReport() HealthReport {
	overallHealth := mc.collector.GetOverallHealth()
	allHealth := mc.collector.GetAllHealth()
	readiness := mc.collector.Readiness()
	buildInfo := GetBuildInfo()

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
