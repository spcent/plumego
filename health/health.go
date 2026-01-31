package health

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// HealthState represents an allowed component health status.
type HealthState string

const (
	StatusHealthy   HealthState = "healthy"
	StatusDegraded  HealthState = "degraded"
	StatusUnhealthy HealthState = "unhealthy"
)

// ComponentChecker defines the interface for health check components.
type ComponentChecker interface {
	Name() string                    // Return component name
	Check(ctx context.Context) error // Perform health check
}

// HealthStatus describes the health of a component in a structured format.
type HealthStatus struct {
	Status       HealthState    `json:"status"`
	Message      string         `json:"message,omitempty"`
	Details      map[string]any `json:"details,omitempty"`
	Timestamp    time.Time      `json:"timestamp"`
	Duration     time.Duration  `json:"duration,omitempty"`
	Dependencies []string       `json:"dependencies,omitempty"`
}

// ComponentHealth represents the health status of a specific component.
type ComponentHealth struct {
	HealthStatus
	Enabled bool `json:"enabled"`
}

// HealthHistoryEntry represents a single entry in health state history.
type HealthHistoryEntry struct {
	Timestamp  time.Time     `json:"timestamp"`
	State      HealthState   `json:"state"`
	Message    string        `json:"message"`
	Components []string      `json:"components,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
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
	HistoryRetention    time.Duration `json:"history_retention"`
	AutoCleanupEnabled  bool          `json:"auto_cleanup_enabled"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
}

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

var errManagerClosed = errors.New("manager is closed")

// These variables can be overridden at build time using -ldflags.
var (
	// Version reports the application version.
	Version = version
	// Commit reports the git commit hash used for the build.
	Commit = commit
	// BuildTime reports when the binary was built.
	BuildTime = buildTime
)

// GetBuildInfo returns the current build metadata.
func GetBuildInfo() BuildInfo {
	return BuildInfo{Version: Version, Commit: Commit, BuildTime: BuildTime}
}

// HealthManager defines the interface for health check management.
type HealthManager interface {
	RegisterComponent(checker ComponentChecker) error
	UnregisterComponent(name string) error
	GetComponentHealth(name string) (*ComponentHealth, bool)
	GetAllHealth() map[string]*ComponentHealth
	CheckComponent(ctx context.Context, name string) error
	CheckAllComponents(ctx context.Context) HealthStatus
	GetOverallHealth() HealthStatus
	GetHealthHistory() []HealthHistoryEntry
	QueryHealthHistory(query HealthHistoryQuery) HealthHistoryQueryResult
	GetHealthHistoryStats() map[string]any
	SetConfig(config HealthCheckConfig) error
	GetConfig() HealthCheckConfig
	ForceCleanup()
	Close() error
}

// healthManager is the concrete implementation of HealthManager.
type healthManager struct {
	mu                sync.RWMutex
	components        map[string]ComponentChecker
	health            map[string]*ComponentHealth
	history           []HealthHistoryEntry
	config            HealthCheckConfig
	cleanupTimer      *time.Timer
	closed            bool
	metrics           *MetricsCollector
	lastCheckTime     time.Time
	lastCheckDuration time.Duration
}

// NewHealthManager creates a new HealthManager instance.
func NewHealthManager(config HealthCheckConfig) (HealthManager, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	hm := &healthManager{
		components: make(map[string]ComponentChecker),
		health:     make(map[string]*ComponentHealth),
		config:     config,
	}

	if config.AutoCleanupEnabled {
		hm.startCleanupTimer()
	}

	return hm, nil
}

// validateConfig validates the health check configuration.
func validateConfig(config HealthCheckConfig) error {
	if config.MaxHistoryEntries < 0 {
		return errors.New("max history entries cannot be negative")
	}
	if config.HistoryRetention < 0 {
		return errors.New("history retention cannot be negative")
	}
	if config.CleanupInterval < 0 {
		return errors.New("cleanup interval cannot be negative")
	}
	if config.Timeout < 0 {
		return errors.New("timeout cannot be negative")
	}
	if config.RetryCount < 0 {
		return errors.New("retry count cannot be negative")
	}
	if config.RetryDelay < 0 {
		return errors.New("retry delay cannot be negative")
	}
	return nil
}

// RegisterComponent registers a health check component.
func (hm *healthManager) RegisterComponent(checker ComponentChecker) error {
	if checker == nil {
		return errors.New("checker cannot be nil")
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.closed {
		return errManagerClosed
	}

	name := checker.Name()
	if name == "" {
		return errors.New("component name cannot be empty")
	}

	hm.components[name] = checker
	hm.health[name] = &ComponentHealth{
		HealthStatus: HealthStatus{
			Status:    StatusUnhealthy,
			Timestamp: time.Now(),
			Message:   "component registered",
		},
		Enabled: true,
	}

	return nil
}

// UnregisterComponent removes a health check component.
func (hm *healthManager) UnregisterComponent(name string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.closed {
		return errManagerClosed
	}

	if _, exists := hm.components[name]; !exists {
		return fmt.Errorf("component %s not found", name)
	}

	delete(hm.components, name)
	delete(hm.health, name)

	return nil
}

// GetComponentHealth returns the health status of a specific component.
func (hm *healthManager) GetComponentHealth(name string) (*ComponentHealth, bool) {
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
func (hm *healthManager) GetAllHealth() map[string]*ComponentHealth {
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
func (hm *healthManager) CheckComponent(ctx context.Context, name string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	hm.mu.RLock()
	if hm.closed {
		hm.mu.RUnlock()
		return errManagerClosed
	}

	checker, exists := hm.components[name]
	config := hm.config
	metrics := hm.metrics
	hm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("component %s not found", name)
	}

	checkCtx, cancel := withCheckTimeout(ctx, config.Timeout)
	defer cancel()

	start := time.Now()
	attempts, err := runCheckWithRetry(checkCtx, checker, config)
	duration := time.Since(start)

	timestamp := time.Now()
	status := StatusHealthy
	message := "component is healthy"
	if err != nil {
		status = StatusUnhealthy
		message = err.Error()
	}

	hm.mu.Lock()
	health, healthExists := hm.health[name]
	if !healthExists {
		hm.mu.Unlock()
		if metrics != nil {
			metrics.RecordCheckWithError(name, duration, err == nil, status, err)
		}
		return fmt.Errorf("health status for component %s not found", name)
	}

	// Update health status
	health.Status = status
	health.Message = message
	health.Timestamp = timestamp
	health.Duration = duration
	if health.Details == nil {
		health.Details = make(map[string]any)
	}
	health.Details["last_check"] = timestamp
	health.Details["check_duration"] = duration.String()
	health.Details["check_attempts"] = attempts
	hm.lastCheckTime = timestamp
	hm.lastCheckDuration = duration
	hm.mu.Unlock()

	if metrics != nil {
		metrics.RecordCheckWithError(name, duration, err == nil, status, err)
	}

	return err
}

// CheckAllComponents performs health checks for all registered components.
func (hm *healthManager) CheckAllComponents(ctx context.Context) HealthStatus {
	if ctx == nil {
		ctx = context.Background()
	}

	hm.mu.RLock()
	if hm.closed {
		hm.mu.RUnlock()
		return HealthStatus{
			Status:    StatusUnhealthy,
			Message:   errManagerClosed.Error(),
			Timestamp: time.Now(),
		}
	}
	components := make([]string, 0, len(hm.components))
	for name := range hm.components {
		components = append(components, name)
	}
	hm.mu.RUnlock()

	if len(components) == 0 {
		return HealthStatus{
			Status:    StatusHealthy,
			Message:   "no components registered",
			Timestamp: time.Now(),
		}
	}

	start := time.Now()

	// Use concurrency for better performance
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

			err := hm.CheckComponent(ctx, componentName)

			results <- componentCheckResult{
				Name:  componentName,
				Error: err,
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

	// Update overall readiness based on component health
	timestamp := time.Now()
	readiness := ReadinessStatus{
		Timestamp:  timestamp,
		Components: make(map[string]bool, len(hm.health)),
	}

	if allHealthy {
		readiness.Ready = true
		readiness.Reason = ""
	} else {
		readiness.Ready = false
		readiness.Reason = fmt.Sprintf("components failed: %v", failedComponents)
	}

	for name, compHealth := range hm.health {
		readiness.Components[name] = compHealth.Status.isReady()
	}

	// Update readiness
	updateReadiness(readiness)

	// Add to history if enabled
	if hm.config.EnableHistory {
		entry := HealthHistoryEntry{
			Timestamp:  timestamp,
			State:      getOverallState(hm.health),
			Message:    readiness.Reason,
			Components: components,
			Duration:   duration,
		}

		hm.history = append(hm.history, entry)
		hm.applyRetentionPolicy()
	}

	// Calculate overall health without calling GetOverallHealth to avoid deadlock
	// We already have the lock, so we can compute it directly
	overallHealthy := true
	var overallMessages []string

	for _, health := range hm.health {
		if health.Status != StatusHealthy {
			overallHealthy = false
			if health.Message != "" {
				overallMessages = append(overallMessages, fmt.Sprintf("%s: %s", health.Status, health.Message))
			}
		}
	}

	overallStatus := StatusHealthy
	if !overallHealthy {
		overallStatus = StatusDegraded
		for _, health := range hm.health {
			if health.Status == StatusUnhealthy {
				overallStatus = StatusUnhealthy
				break
			}
		}
	}

	var overallMessage string
	if len(overallMessages) > 0 {
		overallMessage = fmt.Sprintf("Issues detected: %v", overallMessages)
	}

	hm.lastCheckTime = timestamp
	hm.lastCheckDuration = duration

	return HealthStatus{
		Status:       overallStatus,
		Message:      overallMessage,
		Timestamp:    timestamp,
		Details:      map[string]any{},
		Duration:     duration,
		Dependencies: components,
	}
}

// getOptimalConcurrency determines the optimal number of concurrent checks.
func (hm *healthManager) getOptimalConcurrency(componentCount int) int {
	switch {
	case componentCount <= 5:
		return componentCount
	case componentCount <= 20:
		return 5
	default:
		return 8
	}
}

func withCheckTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		return ctx, func() {}
	}

	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining <= timeout {
			return ctx, func() {}
		}
	}

	return context.WithTimeout(ctx, timeout)
}

func runCheckWithRetry(ctx context.Context, checker ComponentChecker, config HealthCheckConfig) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	maxAttempts := config.RetryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = checker.Check(ctx)
		if err == nil {
			return attempt, nil
		}
		if attempt >= maxAttempts {
			return attempt, err
		}
		if ctx != nil && ctx.Err() != nil {
			return attempt, err
		}
		if !sleepWithContext(ctx, config.RetryDelay) {
			return attempt, err
		}
	}

	return maxAttempts, err
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	if ctx == nil {
		time.Sleep(delay)
		return true
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// componentCheckResult represents the result of a single component health check.
type componentCheckResult struct {
	Name  string
	Error error
}

// GetOverallHealth returns the overall health status of the application.
func (hm *healthManager) GetOverallHealth() HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	allHealthy := true
	var messages []string

	for _, health := range hm.health {
		if health.Status != StatusHealthy {
			allHealthy = false
			if health.Message != "" {
				messages = append(messages, fmt.Sprintf("%s: %s", health.Status, health.Message))
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
func (hm *healthManager) getComponentNames() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	names := make([]string, 0, len(hm.components))
	for name := range hm.components {
		names = append(names, name)
	}
	return names
}

// getLastCheckDuration returns the duration of the last health check.
func (hm *healthManager) getLastCheckDuration() time.Duration {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.lastCheckDuration
}

// GetHealthHistory returns the health history.
func (hm *healthManager) GetHealthHistory() []HealthHistoryEntry {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]HealthHistoryEntry, len(hm.history))
	copy(result, hm.history)

	return result
}

// QueryHealthHistory queries health history with filtering and pagination.
func (hm *healthManager) QueryHealthHistory(query HealthHistoryQuery) HealthHistoryQueryResult {
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

// GetHealthHistoryStats returns statistics about the health history.
func (hm *healthManager) GetHealthHistoryStats() map[string]any {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.history) == 0 {
		return map[string]any{
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

	return map[string]any{
		"total_entries":    len(hm.history),
		"oldest_entry":     oldestEntry,
		"newest_entry":     newestEntry,
		"entries_by_state": stateCounts,
		"retention_config": hm.config,
		"time_span":        newestEntry.Timestamp.Sub(oldestEntry.Timestamp),
	}
}

// SetConfig updates the health check configuration.
func (hm *healthManager) SetConfig(config HealthCheckConfig) error {
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.closed {
		return errManagerClosed
	}

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

	return nil
}

// GetConfig returns the current health check configuration.
func (hm *healthManager) GetConfig() HealthCheckConfig {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.config
}

// startCleanupTimer starts the automatic cleanup timer.
func (hm *healthManager) startCleanupTimer() {
	if !hm.config.AutoCleanupEnabled || hm.cleanupTimer != nil || hm.closed {
		return
	}

	hm.cleanupTimer = time.AfterFunc(hm.config.CleanupInterval, func() {
		hm.mu.Lock()
		hm.cleanupHistory()
		hm.mu.Unlock()

		// Schedule next cleanup if still enabled
		hm.mu.RLock()
		shouldContinue := hm.config.AutoCleanupEnabled && !hm.closed
		hm.mu.RUnlock()

		if shouldContinue {
			hm.startCleanupTimer()
		}
	})
}

// applyRetentionPolicy applies both time-based and count-based retention.
func (hm *healthManager) applyRetentionPolicy() {
	if len(hm.history) == 0 {
		return
	}

	// Apply time-based retention
	if hm.config.HistoryRetention > 0 {
		cutoffTime := time.Now().Add(-hm.config.HistoryRetention)
		keepIndex := 0
		for i, entry := range hm.history {
			if entry.Timestamp.After(cutoffTime) {
				keepIndex = i
				break
			}
			keepIndex = i + 1
		}
		if keepIndex < len(hm.history) {
			hm.history = hm.history[keepIndex:]
		}
	}

	// Apply count-based retention
	if hm.config.MaxHistoryEntries > 0 && len(hm.history) > hm.config.MaxHistoryEntries {
		hm.history = hm.history[len(hm.history)-hm.config.MaxHistoryEntries:]
	}
}

// cleanupHistory performs cleanup of old health history entries.
func (hm *healthManager) cleanupHistory() {
	hm.applyRetentionPolicy()
}

// ForceCleanup forces an immediate cleanup of health history.
func (hm *healthManager) ForceCleanup() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.cleanupHistory()
}

// Close stops the manager and cleans up resources.
func (hm *healthManager) Close() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.closed {
		return nil
	}

	hm.closed = true

	if hm.cleanupTimer != nil {
		hm.cleanupTimer.Stop()
		hm.cleanupTimer = nil
	}

	return nil
}

// HealthHistoryQuery represents a query for filtering health history.
type HealthHistoryQuery struct {
	StartTime *time.Time   `json:"start_time,omitempty"`
	EndTime   *time.Time   `json:"end_time,omitempty"`
	State     *HealthState `json:"state,omitempty"`
	Component string       `json:"component,omitempty"`
	Limit     int          `json:"limit,omitempty"`
	Offset    int          `json:"offset,omitempty"`
}

// HealthHistoryQueryResult represents the result of a health history query.
type HealthHistoryQueryResult struct {
	Entries []HealthHistoryEntry `json:"entries"`
	Total   int                  `json:"total"`
	Limit   int                  `json:"limit"`
	Offset  int                  `json:"offset"`
	HasMore bool                 `json:"has_more"`
}

// ReadinessStatus describes whether the application is ready to serve traffic.
type ReadinessStatus struct {
	Ready      bool            `json:"ready"`
	Reason     string          `json:"reason,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	Components map[string]bool `json:"components,omitempty"`
}

// Global readiness management (for backward compatibility with external packages)
var (
	readinessMu     sync.RWMutex
	globalReadiness ReadinessStatus
)

func init() {
	globalReadiness = ReadinessStatus{
		Ready:     false,
		Reason:    "starting",
		Timestamp: time.Now(),
	}
}

// updateReadiness updates the global readiness status.
func updateReadiness(status ReadinessStatus) {
	readinessMu.Lock()
	defer readinessMu.Unlock()
	globalReadiness = status
}

// GetReadiness returns the current readiness status.
func GetReadiness() ReadinessStatus {
	readinessMu.RLock()
	defer readinessMu.RUnlock()
	return globalReadiness
}

// SetReady marks the application as ready to serve traffic.
func SetReady() {
	readinessMu.Lock()
	defer readinessMu.Unlock()
	globalReadiness = ReadinessStatus{
		Ready:     true,
		Timestamp: time.Now(),
	}
}

// SetNotReady marks the application as not ready and records the reason.
func SetNotReady(reason string) {
	readinessMu.Lock()
	defer readinessMu.Unlock()
	globalReadiness = ReadinessStatus{
		Ready:     false,
		Reason:    reason,
		Timestamp: time.Now(),
	}
}
