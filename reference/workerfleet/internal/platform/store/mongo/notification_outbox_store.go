package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	platformstore "workerfleet/internal/platform/store"
)

const notificationClaimTTL = 2 * time.Minute

type NotificationJobDoc struct {
	ID             string                              `bson:"_id"`
	AlertID        string                              `bson:"alert_id"`
	SinkType       string                              `bson:"sink_type"`
	Alert          platformstore.AlertRecord           `bson:"alert"`
	Status         platformstore.NotificationJobStatus `bson:"status"`
	Attempts       int                                 `bson:"attempts"`
	NextAttemptAt  time.Time                           `bson:"next_attempt_at"`
	LockedUntil    time.Time                           `bson:"locked_until,omitempty"`
	LastErrorClass string                              `bson:"last_error_class,omitempty"`
	LastError      string                              `bson:"last_error,omitempty"`
	CreatedAt      time.Time                           `bson:"created_at"`
	UpdatedAt      time.Time                           `bson:"updated_at"`
	DeliveredAt    time.Time                           `bson:"delivered_at,omitempty"`
}

func (s *Store) EnqueueNotificationJobs(ctx context.Context, jobs []platformstore.NotificationJob) error {
	ctx, cancel := s.operationContext(ctx)
	defer cancel()

	now := s.now()
	for _, job := range jobs {
		if job.JobID == "" {
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
		doc := NotificationJobDocFromRecord(job)
		_, err := s.collections.NotificationJobs.UpdateOne(
			ctx,
			bson.D{{Key: "_id", Value: doc.ID}},
			bson.D{{Key: "$setOnInsert", Value: doc}},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			return err
		}
	}
	return nil
}

func (s *Store) ClaimNotificationJobs(ctx context.Context, now time.Time, limit int) ([]platformstore.NotificationJob, error) {
	ctx, cancel := s.operationContext(ctx)
	defer cancel()

	if limit <= 0 {
		limit = 25
	}
	now = now.UTC()
	filter := bson.D{{Key: "$or", Value: bson.A{
		bson.D{
			{Key: "status", Value: platformstore.NotificationJobPending},
			{Key: "next_attempt_at", Value: bson.D{{Key: "$lte", Value: now}}},
		},
		bson.D{
			{Key: "status", Value: platformstore.NotificationJobProcessing},
			{Key: "locked_until", Value: bson.D{{Key: "$lte", Value: now}}},
		},
	}}}
	cursor, err := s.collections.NotificationJobs.Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "next_attempt_at", Value: 1}, {Key: "_id", Value: 1}}).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []NotificationJobDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	claimed := make([]platformstore.NotificationJob, 0, len(docs))
	for _, doc := range docs {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "status", Value: platformstore.NotificationJobProcessing},
				{Key: "locked_until", Value: now.Add(notificationClaimTTL)},
				{Key: "updated_at", Value: now},
			}},
			{Key: "$inc", Value: bson.D{{Key: "attempts", Value: 1}}},
		}
		result, err := s.collections.NotificationJobs.UpdateOne(ctx, bson.D{
			{Key: "_id", Value: doc.ID},
			{Key: "status", Value: doc.Status},
		}, update)
		if err != nil {
			return nil, err
		}
		if result.MatchedCount == 0 {
			continue
		}
		doc.Status = platformstore.NotificationJobProcessing
		doc.Attempts++
		doc.LockedUntil = now.Add(notificationClaimTTL)
		doc.UpdatedAt = now
		claimed = append(claimed, doc.Record())
	}
	return claimed, nil
}

func (s *Store) MarkNotificationDelivered(ctx context.Context, jobID string, deliveredAt time.Time) error {
	ctx, cancel := s.operationContext(ctx)
	defer cancel()

	_, err := s.collections.NotificationJobs.UpdateOne(ctx, bson.D{{Key: "_id", Value: jobID}}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "status", Value: platformstore.NotificationJobDelivered},
		{Key: "delivered_at", Value: deliveredAt.UTC()},
		{Key: "updated_at", Value: deliveredAt.UTC()},
		{Key: "locked_until", Value: time.Time{}},
	}}})
	return err
}

func (s *Store) MarkNotificationFailed(ctx context.Context, jobID string, failure platformstore.NotificationFailure) error {
	ctx, cancel := s.operationContext(ctx)
	defer cancel()

	status := platformstore.NotificationJobPending
	nextAttempt := failure.NextAttemptAt.UTC()
	if failure.Permanent {
		status = platformstore.NotificationJobFailed
		nextAttempt = time.Time{}
	}
	_, err := s.collections.NotificationJobs.UpdateOne(ctx, bson.D{{Key: "_id", Value: jobID}}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "status", Value: status},
		{Key: "next_attempt_at", Value: nextAttempt},
		{Key: "last_error_class", Value: failure.ErrorClass},
		{Key: "last_error", Value: failure.ErrorMessage},
		{Key: "locked_until", Value: time.Time{}},
		{Key: "updated_at", Value: s.now()},
	}}})
	return err
}

func NotificationJobDocFromRecord(job platformstore.NotificationJob) NotificationJobDoc {
	return NotificationJobDoc{
		ID:             job.JobID,
		AlertID:        job.AlertID,
		SinkType:       string(job.SinkType),
		Alert:          job.Alert,
		Status:         job.Status,
		Attempts:       job.Attempts,
		NextAttemptAt:  job.NextAttemptAt,
		LockedUntil:    job.LockedUntil,
		LastErrorClass: job.LastErrorClass,
		LastError:      job.LastError,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
		DeliveredAt:    job.DeliveredAt,
	}
}

func (doc NotificationJobDoc) Record() platformstore.NotificationJob {
	return platformstore.NotificationJob{
		JobID:          doc.ID,
		AlertID:        doc.AlertID,
		SinkType:       platformstore.NotificationSinkType(doc.SinkType),
		Alert:          doc.Alert,
		Status:         doc.Status,
		Attempts:       doc.Attempts,
		NextAttemptAt:  doc.NextAttemptAt,
		LockedUntil:    doc.LockedUntil,
		LastErrorClass: doc.LastErrorClass,
		LastError:      doc.LastError,
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
		DeliveredAt:    doc.DeliveredAt,
	}
}
