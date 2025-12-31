package health

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// BuildInfo describes the version metadata of the running binary.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

// These variables can be overridden at build time using -ldflags.
var (
	// Version reports the application version.
	Version = version
	// Commit reports the git commit hash used for the build.
	Commit = commit
	// BuildTime reports when the binary was built.
	BuildTime = buildTime
)

var readiness atomic.Value

func init() {
	readiness.Store(ReadinessStatus{Ready: false, Reason: "starting", Timestamp: time.Now()})
}

// ReadinessStatus describes whether the application is ready to serve traffic.
type ReadinessStatus struct {
	Ready      bool            `json:"ready"`
	Reason     string          `json:"reason,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	Components map[string]bool `json:"components,omitempty"`
}

// HealthChangeListener defines the interface for health state change notifications.
type HealthChangeListener interface {
	OnHealthStateChanged(component string, oldState, newState HealthState, duration time.Duration)
}

// HealthHistoryQuery represents a query for filtering health history.
type HealthHistoryQuery struct {
	StartTime *time.Time   `json:"start_time,omitempty"` // 开始时间
	EndTime   *time.Time   `json:"end_time,omitempty"`   // 结束时间
	State     *HealthState `json:"state,omitempty"`      // 健康状态过滤
	Component string       `json:"component,omitempty"`  // 组件名称过滤
	Limit     int          `json:"limit,omitempty"`      // 返回记录数量限制
	Offset    int          `json:"offset,omitempty"`     // 偏移量（分页）
}

// HealthHistoryQueryResult represents the result of a health history query.
type HealthHistoryQueryResult struct {
	Entries []HealthHistoryEntry `json:"entries"`
	Total   int                  `json:"total"`    // 总记录数
	Limit   int                  `json:"limit"`    // 使用的限制
	Offset  int                  `json:"offset"`   // 使用的偏移
	HasMore bool                 `json:"has_more"` // 是否有更多记录
}

// HealthCheckConfig holds configuration for health checks.
type HealthCheckConfig struct {
	Enabled             bool          `json:"enabled"`
	Timeout             time.Duration `json:"timeout"`
	RetryCount          int           `json:"retry_count"`
	RetryDelay          time.Duration `json:"retry_delay"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	EnableHistory       bool          `json:"enable_history"`
	MaxHistoryEntries   int           `json:"max_history_entries"`
	HistoryRetention    time.Duration `json:"history_retention"`    // 历史记录保留时间
	AutoCleanupEnabled  bool          `json:"auto_cleanup_enabled"` // 自动清理启用标志
	CleanupInterval     time.Duration `json:"cleanup_interval"`     // 清理间隔
}

// GetBuildInfo returns the current build metadata.
func GetBuildInfo() BuildInfo {
	return BuildInfo{Version: Version, Commit: Commit, BuildTime: BuildTime}
}

// SetReady marks the application as ready to serve traffic.
func SetReady() {
	readiness.Store(ReadinessStatus{Ready: true})
}

// SetNotReady marks the application as not ready and records the reason.
func SetNotReady(reason string) {
	readiness.Store(ReadinessStatus{Ready: false, Reason: reason})
}

// GetReadiness returns the current readiness status.
func GetReadiness() ReadinessStatus {
	if status, ok := readiness.Load().(ReadinessStatus); ok {
		return status
	}
	return ReadinessStatus{Ready: false, Reason: "unknown"}
}

// RegisterComponent registers a health check component.
func (hm *HealthManager) RegisterComponent(checker ComponentChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.components[checker.Name()] = checker
	hm.health[checker.Name()] = &ComponentHealth{
		HealthStatus: HealthStatus{
			Status:    StatusUnhealthy,
			Timestamp: time.Now(),
			Message:   "component registered",
		},
		Enabled: true,
	}
}

// UnregisterComponent removes a health check component.
func (hm *HealthManager) UnregisterComponent(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.components, name)
	delete(hm.health, name)
}

// GetComponentHealth returns the health status of a specific component.
func (hm *HealthManager) GetComponentHealth(name string) (*ComponentHealth, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	health, exists := hm.health[name]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	healthCopy := *health
	healthCopy.HealthStatus.Details = make(map[string]any, len(health.HealthStatus.Details))
	for k, v := range health.HealthStatus.Details {
		healthCopy.HealthStatus.Details[k] = v
	}

	return &healthCopy, true
}

// GetAllHealth returns the health status of all components.
func (hm *HealthManager) GetAllHealth() map[string]*ComponentHealth {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result := make(map[string]*ComponentHealth)
	for name, health := range hm.health {
		healthCopy := *health
		healthCopy.HealthStatus.Details = make(map[string]any, len(health.HealthStatus.Details))
		for k, v := range health.HealthStatus.Details {
			healthCopy.HealthStatus.Details[k] = v
		}
		result[name] = &healthCopy
	}

	return result
}

// CheckComponent performs a health check for a specific component.
func (hm *HealthManager) CheckComponent(ctx context.Context, name string) error {
	hm.mu.RLock()
	checker, exists := hm.components[name]
	health, healthExists := hm.health[name]
	hm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("component %s not found", name)
	}

	if !healthExists {
		return fmt.Errorf("health status for component %s not found", name)
	}

	start := time.Now()
	err := checker.Check(ctx)
	duration := time.Since(start)

	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Update health status
	if err != nil {
		health.Status = StatusUnhealthy
		health.Message = err.Error()
	} else {
		health.Status = StatusHealthy
		health.Message = "component is healthy"
	}

	health.Timestamp = time.Now()
	health.Duration = duration
	if health.Details == nil {
		health.Details = make(map[string]any)
	}
	health.Details["last_check"] = health.Timestamp
	health.Details["check_duration"] = duration.String()

	return err
}

// CheckAllComponents performs health checks for all registered components with optimized performance.
func (hm *HealthManager) CheckAllComponents(ctx context.Context) {
	hm.mu.RLock()
	components := make([]string, 0, len(hm.components))
	for name := range hm.components {
		components = append(components, name)
	}
	hm.mu.RUnlock()

	if len(components) == 0 {
		return
	}

	start := time.Now()

	// Use concurrency for better performance with multiple components
	maxConcurrency := hm.getOptimalConcurrency(len(components))
	sem := make(chan struct{}, maxConcurrency)
	results := make(chan componentCheckResult, len(components))

	var wg sync.WaitGroup

	// Launch concurrent health checks
	for _, name := range components {
		wg.Add(1)
		go func(componentName string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			// Perform health check with timeout
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			err := hm.CheckComponent(checkCtx, componentName)
			duration := time.Since(start)

			results <- componentCheckResult{
				Name:     componentName,
				Error:    err,
				Duration: duration,
			}
		}(name)
	}

	// Wait for all checks to complete
	wg.Wait()
	close(results)

	// Collect results
	var failedComponents []string
	allHealthy := true

	for result := range results {
		if result.Error != nil {
			allHealthy = false
			failedComponents = append(failedComponents, result.Name)
		}
	}

	duration := time.Since(start)

	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.lastCheck = time.Now()

	// Update overall readiness based on component health
	if allHealthy {
		hm.readiness.Ready = true
		hm.readiness.Reason = ""
	} else {
		hm.readiness.Ready = false
		hm.readiness.Reason = fmt.Sprintf("components failed: %v", failedComponents)
	}

	// Add to history
	entry := HealthHistoryEntry{
		Timestamp:  time.Now(),
		State:      getOverallState(hm.health),
		Message:    hm.readiness.Reason,
		Components: components,
		Duration:   duration,
	}

	hm.history = append(hm.history, entry)
	// Keep only last 100 entries
	if len(hm.history) > 100 {
		hm.history = hm.history[len(hm.history)-100:]
	}
}

// componentCheckResult represents the result of a single component health check.
type componentCheckResult struct {
	Name     string
	Error    error
	Duration time.Duration
}

// getOptimalConcurrency determines the optimal number of concurrent checks based on component count.
func (hm *HealthManager) getOptimalConcurrency(componentCount int) int {
	// Use reasonable defaults based on component count
	switch {
	case componentCount <= 5:
		return componentCount // Check all concurrently for small number
	case componentCount <= 20:
		return 5 // Limit to 5 concurrent checks for medium number
	default:
		return 8 // Limit to 8 concurrent checks for large number
	}
}

// CheckAllComponentsWithRetry performs health checks with retry logic for failed components.
func (hm *HealthManager) CheckAllComponentsWithRetry(ctx context.Context, maxRetries int) {
	hm.CheckAllComponents(ctx)

	// Retry failed components
	for retry := 0; retry < maxRetries; retry++ {
		// Get failed components from last check
		failedComponents := hm.getFailedComponents()
		if len(failedComponents) == 0 {
			break // All components are now healthy
		}

		// Wait before retry (exponential backoff)
		if retry > 0 {
			backoff := time.Duration(1<<uint(retry)) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
		}

		// Retry only failed components
		hm.retryFailedComponents(ctx, failedComponents)
	}
}

// getFailedComponents returns the list of components that failed the last health check.
func (hm *HealthManager) getFailedComponents() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var failed []string
	for name, health := range hm.health {
		if health.Status == StatusUnhealthy {
			failed = append(failed, name)
		}
	}
	return failed
}

// retryFailedComponents retries health checks for the specified failed components.
func (hm *HealthManager) retryFailedComponents(ctx context.Context, failedComponents []string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3) // Limit concurrent retries

	for _, name := range failedComponents {
		wg.Add(1)
		go func(componentName string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			_ = hm.CheckComponent(checkCtx, componentName)
		}(name)
	}

	wg.Wait()
}

// GetHealthHistory returns the health history.
func (hm *HealthManager) GetHealthHistory() []HealthHistoryEntry {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]HealthHistoryEntry, len(hm.history))
	copy(result, hm.history)

	return result
}

// QueryHealthHistory queries health history with filtering and pagination.
func (hm *HealthManager) QueryHealthHistory(query HealthHistoryQuery) HealthHistoryQueryResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Filter entries based on query criteria
	var filtered []HealthHistoryEntry
	for _, entry := range hm.history {
		// Time range filter
		if query.StartTime != nil && entry.Timestamp.Before(*query.StartTime) {
			continue
		}
		if query.EndTime != nil && entry.Timestamp.After(*query.EndTime) {
			continue
		}

		// State filter
		if query.State != nil && entry.State != *query.State {
			continue
		}

		// Component filter
		if query.Component != "" {
			found := false
			for _, comp := range entry.Components {
				if comp == query.Component {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	// Calculate total before pagination
	total := len(filtered)

	// Apply pagination
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	// Ensure we don't exceed bounds
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	entries := make([]HealthHistoryEntry, 0)
	if offset < len(filtered) {
		entries = filtered[offset:end]
	}

	hasMore := end < len(filtered)

	return HealthHistoryQueryResult{
		Entries: entries,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
}

// GetOverallHealth returns the overall health status of the application.
func (hm *HealthManager) GetOverallHealth() HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	allHealthy := true
	var messages []string

	for _, health := range hm.health {
		if health.Status != StatusHealthy {
			allHealthy = false
			if health.Message != "" {
				messages = append(messages, fmt.Sprintf("%s: %s", health.HealthStatus.Status, health.Message))
			}
		}
	}

	status := StatusHealthy
	if !allHealthy {
		status = StatusDegraded
		// Check if any component is unhealthy (not just degraded)
		for _, health := range hm.health {
			if health.Status == StatusUnhealthy {
				status = StatusUnhealthy
				break
			}
		}
	}

	var message string
	if len(messages) > 0 {
		message = fmt.Sprintf("Issues detected: %v", messages)
	}

	return HealthStatus{
		Status:       status,
		Message:      message,
		Timestamp:    time.Now(),
		Details:      map[string]any{},
		Duration:     hm.getLastCheckDuration(),
		Dependencies: hm.getComponentNames(),
	}
}

// getOverallState determines the overall health state from component healths.
func getOverallState(healths map[string]*ComponentHealth) HealthState {
	hasUnhealthy := false
	hasDegraded := false

	for _, health := range healths {
		switch health.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

// getComponentNames returns the list of all component names.
func (hm *HealthManager) getComponentNames() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	names := make([]string, 0, len(hm.components))
	for name := range hm.components {
		names = append(names, name)
	}
	return names
}

// getLastCheckDuration returns the duration since last check.
func (hm *HealthManager) getLastCheckDuration() time.Duration {
	if hm.lastCheck.IsZero() {
		return 0
	}
	return time.Since(hm.lastCheck)
}

// SetConfig updates the health check configuration and restarts cleanup timer if needed.
func (hm *HealthManager) SetConfig(config HealthCheckConfig) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.config = config

	// Restart cleanup timer if auto cleanup is enabled
	if hm.cleanupTimer != nil {
		hm.cleanupTimer.Stop()
		hm.cleanupTimer = nil
	}

	if config.AutoCleanupEnabled {
		hm.startCleanupTimer()
	}

	// Immediately run cleanup based on new settings
	hm.cleanupHistory()
}

// GetConfig returns the current health check configuration.
func (hm *HealthManager) GetConfig() HealthCheckConfig {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.config
}

// startCleanupTimer starts the automatic cleanup timer.
func (hm *HealthManager) startCleanupTimer() {
	if !hm.config.AutoCleanupEnabled || hm.cleanupTimer != nil {
		return
	}

	hm.cleanupTimer = time.AfterFunc(hm.config.CleanupInterval, func() {
		hm.mu.Lock()
		hm.cleanupHistory()
		hm.mu.Unlock()

		// Schedule next cleanup
		hm.startCleanupTimer()
	})
}

// cleanupHistory performs cleanup of old health history entries based on retention policy.
func (hm *HealthManager) cleanupHistory() {
	if len(hm.history) == 0 {
		return
	}

	cutoffTime := time.Now().Add(-hm.config.HistoryRetention)

	// Find the index where entries start to be within retention period
	keepIndex := 0
	for i, entry := range hm.history {
		if entry.Timestamp.After(cutoffTime) {
			keepIndex = i
			break
		}
		keepIndex = i + 1
	}

	// Apply retention policy: keep only entries within retention period AND within max entries limit
	if keepIndex < len(hm.history) {
		hm.history = hm.history[keepIndex:]
	}

	// Also enforce maximum entries limit
	maxEntries := hm.config.MaxHistoryEntries
	if len(hm.history) > maxEntries {
		hm.history = hm.history[len(hm.history)-maxEntries:]
	}
}

// GetHealthHistoryStats returns statistics about the health history.
func (hm *HealthManager) GetHealthHistoryStats() map[string]interface{} {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.history) == 0 {
		return map[string]interface{}{
			"total_entries":    0,
			"oldest_entry":     nil,
			"newest_entry":     nil,
			"entries_by_state": map[string]int{},
			"retention_config": hm.config,
		}
	}

	// Count entries by state
	stateCounts := make(map[string]int)
	for _, entry := range hm.history {
		stateCounts[string(entry.State)]++
	}

	oldestEntry := hm.history[0]
	newestEntry := hm.history[len(hm.history)-1]

	return map[string]interface{}{
		"total_entries":    len(hm.history),
		"oldest_entry":     oldestEntry,
		"newest_entry":     newestEntry,
		"entries_by_state": stateCounts,
		"retention_config": hm.config,
		"time_span":        newestEntry.Timestamp.Sub(oldestEntry.Timestamp),
	}
}

// ForceCleanup forces an immediate cleanup of health history.
func (hm *HealthManager) ForceCleanup() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.cleanupHistory()
}
