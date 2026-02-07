package messaging

import (
	"context"
	"time"

	log "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/net/mq/store"
	"github.com/spcent/plumego/scheduler"
)

// registerScheduledJobs sets up periodic maintenance tasks on the scheduler.
func (s *Service) registerScheduledJobs() {
	if s.scheduler == nil {
		return
	}

	// Retry dead-letter tasks every hour.
	s.scheduler.AddCron("messaging.dlq-retry", "0 * * * *", s.retryDeadLetters,
		scheduler.WithTimeout(5*time.Minute),
		scheduler.WithOverlapPolicy(scheduler.SkipIfRunning),
		scheduler.WithGroup("messaging"),
		scheduler.WithTags("dlq", "retry"),
	)

	// Log queue stats every 5 minutes.
	s.scheduler.AddCron("messaging.stats-log", "*/5 * * * *", s.logStats,
		scheduler.WithTimeout(30*time.Second),
		scheduler.WithOverlapPolicy(scheduler.SkipIfRunning),
		scheduler.WithGroup("messaging"),
		scheduler.WithTags("stats"),
	)
}

// retryDeadLetters replays dead-letter tasks back into the queue.
func (s *Service) retryDeadLetters(ctx context.Context) error {
	replayer, ok := s.store.(store.DLQReplayer)
	if !ok {
		return nil
	}
	result, err := replayer.ReplayDLQ(ctx, store.ReplayOptions{
		Max:           100,
		Now:           time.Now(),
		AvailableAt:   time.Now(),
		ResetAttempts: true,
	})
	if err != nil {
		return err
	}
	if s.logger != nil && result.Replayed > 0 {
		s.logger.Info("dlq replay completed", log.Fields{
			"replayed":  result.Replayed,
			"remaining": result.Remaining,
		})
	}
	return nil
}

// logStats writes queue depth to the logger.
func (s *Service) logStats(ctx context.Context) error {
	stats, err := s.queue.Stats(ctx)
	if err != nil {
		return err
	}
	if s.logger != nil {
		s.logger.Info("messaging queue stats", log.Fields{
			"queued":       stats.Queued,
			"leased":       stats.Leased,
			"dead":         stats.Dead,
			"expired":      stats.Expired,
			"total_sent":   s.totalSent.Load(),
			"total_failed": s.totalFailed.Load(),
		})
	}
	return nil
}
