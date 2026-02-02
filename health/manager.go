package health

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

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

// getComponentNames returns the list of all component names.
func (hm *healthManager) getComponentNames() []string {
	names := make([]string, 0, len(hm.components))
	for name := range hm.components {
		names = append(names, name)
	}
	return names
}

// getLastCheckDuration returns the duration of the last health check.
func (hm *healthManager) getLastCheckDuration() time.Duration {
	return hm.lastCheckDuration
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
