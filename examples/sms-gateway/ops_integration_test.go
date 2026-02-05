package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/core/components/ops"
	"github.com/spcent/plumego/examples/sms-gateway/internal/message"
	"github.com/spcent/plumego/net/mq"
	mqstore "github.com/spcent/plumego/net/mq/store"
	plumrouter "github.com/spcent/plumego/router"
)

type queueStatsResp struct {
	Data struct {
		Queues []ops.QueueStats `json:"queues"`
	} `json:"data"`
}

type replayResp struct {
	Data struct {
		Replay ops.QueueReplayResult `json:"replay"`
	} `json:"data"`
}

type receiptResp struct {
	Data struct {
		Receipt ops.ReceiptRecord `json:"receipt"`
	} `json:"data"`
}

func TestOpsQueueReplayAndReceiptLookup(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC)

	repo := message.NewMemoryRepository()
	if err := repo.Insert(ctx, message.Message{
		ID:        "msg-1",
		TenantID:  "tenant-1",
		Provider:  "provider-a",
		Status:    message.StatusSent,
		CreatedAt: now.Add(-2 * time.Minute),
		UpdatedAt: now.Add(-1 * time.Minute),
	}); err != nil {
		t.Fatalf("insert message: %v", err)
	}

	queueStore := mqstore.NewMemory(mqstore.MemConfig{Now: func() time.Time { return now }})
	queue := mq.NewTaskQueue(queueStore, mq.WithQueueNowFunc(func() time.Time { return now }))
	task := mq.Task{
		ID:        "task-1",
		Topic:     "send",
		TenantID:  "tenant-1",
		Payload:   []byte(`{"message_id":"msg-1"}`),
		CreatedAt: now.Add(-1 * time.Minute),
	}
	if err := queue.Enqueue(ctx, task, mq.EnqueueOptions{}); err != nil {
		t.Fatalf("enqueue task: %v", err)
	}
	reserved, err := queue.Reserve(ctx, mq.ReserveOptions{
		Topics:     []string{"send"},
		Limit:      1,
		Lease:      30 * time.Second,
		ConsumerID: "worker-1",
		Now:        now,
	})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if len(reserved) != 1 {
		t.Fatalf("expected 1 reserved task, got %d", len(reserved))
	}
	if err := queue.MoveToDLQ(ctx, reserved[0].ID, "worker-1", "failed", now); err != nil {
		t.Fatalf("move to DLQ: %v", err)
	}

	queueName := "send"
	hooks := ops.Hooks{
		QueueStats: func(ctx context.Context, queueID string) (ops.QueueStats, error) {
			stats, err := queue.Stats(ctx)
			if err != nil {
				return ops.QueueStats{}, err
			}
			return ops.QueueStats{
				Queue:     queueName,
				Queued:    stats.Queued,
				Leased:    stats.Leased,
				Dead:      stats.Dead,
				Expired:   stats.Expired,
				UpdatedAt: now,
			}, nil
		},
		QueueList: func(ctx context.Context) ([]string, error) {
			return []string{queueName}, nil
		},
		QueueReplay: func(ctx context.Context, req ops.QueueReplayRequest) (ops.QueueReplayResult, error) {
			replayer, ok := any(queueStore).(mqstore.DLQReplayer)
			if !ok {
				return ops.QueueReplayResult{}, nil
			}
			result, err := replayer.ReplayDLQ(ctx, mqstore.ReplayOptions{
				Max:           req.Max,
				ResetAttempts: true,
				Now:           now,
				AvailableAt:   now,
			})
			if err != nil {
				return ops.QueueReplayResult{}, err
			}
			return ops.QueueReplayResult{
				Queue:     queueName,
				Requested: req.Max,
				Replayed:  result.Replayed,
				Remaining: result.Remaining,
			}, nil
		},
		ReceiptLookup: func(ctx context.Context, messageID string) (ops.ReceiptRecord, error) {
			msg, found, err := repo.Get(ctx, messageID)
			if err != nil {
				return ops.ReceiptRecord{}, err
			}
			if !found {
				return ops.ReceiptRecord{MessageID: messageID, Status: "not_found"}, nil
			}
			return ops.ReceiptRecord{
				MessageID: msg.ID,
				Status:    string(msg.Status),
				Provider:  msg.Provider,
				UpdatedAt: msg.UpdatedAt,
			}, nil
		},
	}

	opsComponent := ops.NewComponent(ops.Options{
		Enabled: true,
		Auth:    ops.AuthConfig{AllowInsecure: true},
		Hooks:   hooks,
	})
	r := plumrouter.NewRouter()
	opsComponent.RegisterRoutes(r)

	statsBefore := queueStatsResp{}
	req := httptest.NewRequest(http.MethodGet, "/ops/queue?queue=send", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("queue stats status: %d", rec.Code)
	}
	if err := json.NewDecoder(rec.Body).Decode(&statsBefore); err != nil {
		t.Fatalf("decode queue stats: %v", err)
	}
	if len(statsBefore.Data.Queues) != 1 || statsBefore.Data.Queues[0].Dead != 1 {
		t.Fatalf("expected dead=1, got %+v", statsBefore.Data.Queues)
	}

	replayPayload, _ := json.Marshal(map[string]any{"queue": "send", "max": 1})
	replayReq := httptest.NewRequest(http.MethodPost, "/ops/queue/replay", bytes.NewBuffer(replayPayload))
	replayReq.Header.Set("Content-Type", "application/json")
	replayRec := httptest.NewRecorder()
	r.ServeHTTP(replayRec, replayReq)
	if replayRec.Code != http.StatusOK {
		t.Fatalf("replay status: %d", replayRec.Code)
	}
	var replay replayResp
	if err := json.NewDecoder(replayRec.Body).Decode(&replay); err != nil {
		t.Fatalf("decode replay: %v", err)
	}
	if replay.Data.Replay.Replayed != 1 {
		t.Fatalf("expected replayed=1, got %+v", replay.Data.Replay)
	}

	statsAfter := queueStatsResp{}
	reqAfter := httptest.NewRequest(http.MethodGet, "/ops/queue?queue=send", nil)
	recAfter := httptest.NewRecorder()
	r.ServeHTTP(recAfter, reqAfter)
	if recAfter.Code != http.StatusOK {
		t.Fatalf("queue stats status after: %d", recAfter.Code)
	}
	if err := json.NewDecoder(recAfter.Body).Decode(&statsAfter); err != nil {
		t.Fatalf("decode queue stats after: %v", err)
	}
	if len(statsAfter.Data.Queues) != 1 || statsAfter.Data.Queues[0].Dead != 0 {
		t.Fatalf("expected dead=0, got %+v", statsAfter.Data.Queues)
	}

	receiptReq := httptest.NewRequest(http.MethodGet, "/ops/receipts?message_id=msg-1", nil)
	receiptRec := httptest.NewRecorder()
	r.ServeHTTP(receiptRec, receiptReq)
	if receiptRec.Code != http.StatusOK {
		t.Fatalf("receipt status: %d", receiptRec.Code)
	}
	var receipt receiptResp
	if err := json.NewDecoder(receiptRec.Body).Decode(&receipt); err != nil {
		t.Fatalf("decode receipt: %v", err)
	}
	if receipt.Data.Receipt.Status != string(message.StatusSent) {
		t.Fatalf("expected status %s, got %s", message.StatusSent, receipt.Data.Receipt.Status)
	}
	if receipt.Data.Receipt.Provider != "provider-a" {
		t.Fatalf("expected provider provider-a, got %s", receipt.Data.Receipt.Provider)
	}
}
