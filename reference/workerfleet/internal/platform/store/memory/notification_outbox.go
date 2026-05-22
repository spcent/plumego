package memory

import (
	"context"
	"sort"
	"time"

	platformstore "workerfleet/internal/platform/store"
)

const notificationClaimTTL = 2 * time.Minute

func (s *Store) EnqueueNotificationJobs(ctx context.Context, jobs []platformstore.NotificationJob) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range jobs {
		if job.JobID == "" {
			continue
		}
		if _, exists := s.notificationJobs[job.JobID]; exists {
			continue
		}
		if job.Status == "" {
			job.Status = platformstore.NotificationJobPending
		}
		if job.NextAttemptAt.IsZero() {
			job.NextAttemptAt = now
		}
		if job.CreatedAt.IsZero() {
			job.CreatedAt = now
		}
		job.UpdatedAt = now
		s.notificationJobs[job.JobID] = cloneNotificationJob(job)
	}
	return nil
}

func (s *Store) ClaimNotificationJobs(ctx context.Context, now time.Time, limit int) ([]platformstore.NotificationJob, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 25
	}
	now = now.UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]string, 0, len(s.notificationJobs))
	for id, job := range s.notificationJobs {
		duePending := job.Status == platformstore.NotificationJobPending && !job.NextAttemptAt.After(now)
		expiredProcessing := job.Status == platformstore.NotificationJobProcessing && !job.LockedUntil.After(now)
		if duePending || expiredProcessing {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	if len(ids) > limit {
		ids = ids[:limit]
	}

	claimed := make([]platformstore.NotificationJob, 0, len(ids))
	for _, id := range ids {
		job := s.notificationJobs[id]
		job.Status = platformstore.NotificationJobProcessing
		job.Attempts++
		job.LockedUntil = now.Add(notificationClaimTTL)
		job.UpdatedAt = now
		s.notificationJobs[id] = job
		claimed = append(claimed, cloneNotificationJob(job))
	}
	return claimed, nil
}

func (s *Store) MarkNotificationDelivered(ctx context.Context, jobID string, deliveredAt time.Time) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.notificationJobs[jobID]
	if !ok {
		return nil
	}
	job.Status = platformstore.NotificationJobDelivered
	job.DeliveredAt = deliveredAt.UTC()
	job.UpdatedAt = deliveredAt.UTC()
	job.LockedUntil = time.Time{}
	s.notificationJobs[jobID] = job
	return nil
}

func (s *Store) MarkNotificationFailed(ctx context.Context, jobID string, failure platformstore.NotificationFailure) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.notificationJobs[jobID]
	if !ok {
		return nil
	}
	if failure.Permanent {
		job.Status = platformstore.NotificationJobFailed
	} else {
		job.Status = platformstore.NotificationJobPending
		job.NextAttemptAt = failure.NextAttemptAt.UTC()
	}
	job.LastErrorClass = failure.ErrorClass
	job.LastError = failure.ErrorMessage
	job.LockedUntil = time.Time{}
	job.UpdatedAt = time.Now().UTC()
	s.notificationJobs[jobID] = job
	return nil
}

func cloneNotificationJob(job platformstore.NotificationJob) platformstore.NotificationJob {
	job.Alert.Details = cloneStringMap(job.Alert.Details)
	return job
}
