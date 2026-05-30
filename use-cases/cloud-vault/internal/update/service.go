package update

import (
	"context"
	"sync"
	"time"

	plumelog "github.com/spcent/plumego/log"
)

// Service manages update checking and caching.
type Service struct {
	checker       *Checker
	config        *Config
	logger        plumelog.StructuredLogger
	mu            sync.RWMutex
	cachedStatus  *UpdateStatus
	lastCheckTime *time.Time
}

// Config holds update service configuration.
type Config struct {
	Enabled          bool
	CheckOnStartup   bool
	UpdateURL        string
	CheckIntervalMin int
	Channel          string
}

// NewService creates a new update service.
func NewService(checker *Checker, config *Config, logger plumelog.StructuredLogger) *Service {
	return &Service{
		checker: checker,
		config:  config,
		logger:  logger,
	}
}

// GetStatus returns the current update status.
func (s *Service) GetStatus(ctx context.Context) *UpdateStatus {
	s.mu.RLock()
	cached := s.cachedStatus
	s.mu.RUnlock()

	if cached != nil {
		return cached
	}

	// If no cached status, return current version only
	return &UpdateStatus{
		CurrentVersion: s.checker.currentVersion,
		CheckEnabled:   s.config.Enabled,
	}
}

// CheckNow performs an immediate update check.
func (s *Service) CheckNow(ctx context.Context) (*UpdateStatus, error) {
	if !s.config.Enabled {
		return &UpdateStatus{
			CurrentVersion: s.checker.currentVersion,
			CheckEnabled:   false,
		}, nil
	}

	release, err := s.checker.CheckForUpdate(ctx)
	if err != nil {
		s.logger.Error("update check failed", plumelog.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	now := time.Now()
	nextCheck := now.Add(time.Duration(s.config.CheckIntervalMin) * time.Minute)

	status := &UpdateStatus{
		CurrentVersion:  s.checker.currentVersion,
		LatestRelease:   release,
		UpdateAvailable: CompareVersions(s.checker.currentVersion.Version, release.Version),
		LastCheck:       &now,
		NextCheck:       &nextCheck,
		CheckEnabled:    true,
	}

	s.mu.Lock()
	s.cachedStatus = status
	s.lastCheckTime = &now
	s.mu.Unlock()

	s.logger.Info("update check completed", plumelog.Fields{
		"current_version":  s.checker.currentVersion.Version,
		"latest_version":   release.Version,
		"update_available": status.UpdateAvailable,
	})

	return status, nil
}

// StartBackgroundChecker starts a background goroutine that periodically checks for updates.
func (s *Service) StartBackgroundChecker(ctx context.Context) {
	if !s.config.Enabled || s.config.CheckIntervalMin <= 0 {
		return
	}

	// Check on startup if configured
	if s.config.CheckOnStartup {
		go func() {
			// Delay startup check by 30 seconds
			time.Sleep(30 * time.Second)
			if _, err := s.CheckNow(ctx); err != nil {
				s.logger.Warn("startup update check failed", plumelog.Fields{
					"error": err.Error(),
				})
			}
		}()
	}

	// Periodic checks
	go func() {
		ticker := time.NewTicker(time.Duration(s.config.CheckIntervalMin) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := s.CheckNow(ctx); err != nil {
					s.logger.Warn("periodic update check failed", plumelog.Fields{
						"error": err.Error(),
					})
				}
			}
		}
	}()
}
