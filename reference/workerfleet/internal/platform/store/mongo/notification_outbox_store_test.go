package mongo

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func TestNotificationOutboxEnqueueClaimRetryDeliver(t *testing.T) {
	uri := os.Getenv("WORKERFLEET_MONGO_TEST_URI")
	if strings.TrimSpace(uri) == "" {
		t.Skip("set WORKERFLEET_MONGO_TEST_URI to run MongoDB notification outbox integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Fatalf("disconnect mongo: %v", err)
		}
	}()

	db := client.Database(fmt.Sprintf("workerfleet_outbox_test_%d", time.Now().UnixNano()))
	defer func() {
		if err := db.Drop(context.Background()); err != nil {
			t.Fatalf("drop test database: %v", err)
		}
	}()
	if err := EnsureIndexes(ctx, db); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	store, err := NewStore(db, WithOperationTimeout(5*time.Second))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	job := platformstore.NotificationJob{
		JobID:         "alert-1:webhook",
		AlertID:       "alert-1",
		SinkType:      platformstore.NotificationSinkWebhook,
		Status:        platformstore.NotificationJobPending,
		NextAttemptAt: now,
		Alert: platformstore.AlertRecord{
			AlertID:   "alert-1",
			AlertType: domain.AlertWorkerOffline,
			Status:    domain.AlertStatusFiring,
			WorkerID:  "worker-1",
		},
	}
	if err := store.EnqueueNotificationJobs(ctx, []platformstore.NotificationJob{job, job}); err != nil {
		t.Fatalf("enqueue jobs: %v", err)
	}
	claimed, err := store.ClaimNotificationJobs(ctx, now, 10)
	if err != nil {
		t.Fatalf("claim jobs: %v", err)
	}
	if len(claimed) != 1 || claimed[0].Attempts != 1 {
		t.Fatalf("claimed = %#v, want one first attempt", claimed)
	}

	if err := store.MarkNotificationFailed(ctx, claimed[0].JobID, platformstore.NotificationFailure{
		ErrorClass:    "http_5xx",
		ErrorMessage:  "server error",
		NextAttemptAt: now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("mark failed: %v", err)
	}
	claimed, err = store.ClaimNotificationJobs(ctx, now.Add(30*time.Second), 10)
	if err != nil {
		t.Fatalf("claim before retry: %v", err)
	}
	if len(claimed) != 0 {
		t.Fatalf("claimed before retry = %#v, want none", claimed)
	}
	claimed, err = store.ClaimNotificationJobs(ctx, now.Add(time.Minute), 10)
	if err != nil {
		t.Fatalf("claim retry: %v", err)
	}
	if len(claimed) != 1 || claimed[0].Attempts != 2 {
		t.Fatalf("retry claimed = %#v, want second attempt", claimed)
	}
	if err := store.MarkNotificationDelivered(ctx, claimed[0].JobID, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("mark delivered: %v", err)
	}
	claimed, err = store.ClaimNotificationJobs(ctx, now.Add(3*time.Minute), 10)
	if err != nil {
		t.Fatalf("claim after delivered: %v", err)
	}
	if len(claimed) != 0 {
		t.Fatalf("claimed after delivered = %#v, want none", claimed)
	}
}
