package main

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/examples/sms-gateway/internal/message"
	"github.com/spcent/plumego/examples/sms-gateway/internal/pipeline"
	"github.com/spcent/plumego/examples/sms-gateway/internal/routing"
	"github.com/spcent/plumego/examples/sms-gateway/internal/tasks"
	logpkg "github.com/spcent/plumego/log"
	tenantmw "github.com/spcent/plumego/middleware/tenant"
	"github.com/spcent/plumego/net/mq"
	mqstore "github.com/spcent/plumego/net/mq/store"
	"github.com/spcent/plumego/store/idempotency"
	kvstore "github.com/spcent/plumego/store/kv"
	"github.com/spcent/plumego/tenant"
)

func TestSMSGatewayPipeline(t *testing.T) {
	ctx := context.Background()

	repo := message.NewMemoryRepository()

	policyStore := tenant.NewInMemoryRoutePolicyStore()
	if err := policyStore.SetRoutePolicy(ctx, tenant.RoutePolicy{
		TenantID: "tenant-1",
		Strategy: "weighted",
		Payload:  []byte(`{"providers":[{"provider":"provider-a","weight":100}]}`),
	}); err != nil {
		t.Fatalf("set policy: %v", err)
	}
	if err := policyStore.SetRoutePolicy(ctx, tenant.RoutePolicy{
		TenantID: "tenant-2",
		Strategy: "direct",
		Payload:  []byte(`{"default_provider":"provider-fail"}`),
	}); err != nil {
		t.Fatalf("set policy: %v", err)
	}
	policyCache := tenant.NewInMemoryRoutePolicyCache(10, time.Minute)
	policyProvider := tenant.NewCachedRoutePolicyProvider(policyStore, policyCache)
	router := &routing.PolicyRouter{Provider: policyProvider, Random: fixedRand{value: 0}}

	queue := mq.NewTaskQueue(mqstore.NewMemory(mqstore.MemConfig{}))
	worker := mq.NewWorker(queue, mq.WorkerConfig{
		ConsumerID:      "sms-worker",
		Concurrency:     1,
		PollInterval:    50 * time.Millisecond,
		RetryPolicy:     mq.ExponentialBackoff{Base: 20 * time.Millisecond, Max: 200 * time.Millisecond, Factor: 2},
		ShutdownTimeout: time.Second,
	})

	processor := &pipeline.Processor{
		Repo:   repo,
		Router: router,
		Providers: map[string]pipeline.ProviderSender{
			"provider-a":    &pipeline.MockProvider{Name: "provider-a"},
			"provider-fail": &pipeline.MockProvider{Name: "provider-fail", FailFirstN: 10},
		},
	}

	worker.Register(tasks.SendTopic, processor.Handle)
	worker.Start(ctx)
	defer func() {
		_ = worker.Stop(context.Background())
	}()

	idemStore, cleanup := newTestIdemStore(t)
	defer cleanup()

	handler := contract.AdaptCtxHandler(
		message.ExampleSendHandler(idemStore, repo, router, queue),
		logpkg.NewGLogger(),
	)
	handler = tenantmw.TenantResolver(tenantmw.TenantResolverOptions{
		HeaderName: "X-Tenant-ID",
	})(handler)

	server := httptest.NewServer(handler)
	defer server.Close()

	successID, err := sendRequest(server.URL, "tenant-1", "+10000000001", "hello", 2)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	if !waitForStatus(ctx, 2*time.Second, repo, successID, message.StatusSent) {
		t.Fatalf("expected message to be sent")
	}
	latest, ok, _ := repo.Get(ctx, successID)
	if !ok {
		t.Fatalf("message missing")
	}
	if latest.Provider != "provider-a" {
		t.Fatalf("expected provider-a, got %s", latest.Provider)
	}

	failID, err := sendRequest(server.URL, "tenant-2", "+10000000002", "fail", 2)
	if err != nil {
		t.Fatalf("send failure request: %v", err)
	}
	if !waitForDLQ(ctx, 3*time.Second, queue, repo, failID) {
		t.Fatalf("expected DLQ")
	}
}

func waitForStatus(ctx context.Context, timeout time.Duration, repo *message.MemoryRepository, id string, status message.Status) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		msg, ok, _ := repo.Get(ctx, id)
		if ok && msg.Status == status {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func waitForDLQ(ctx context.Context, timeout time.Duration, queue *mq.TaskQueue, repo *message.MemoryRepository, id string) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		stats, _ := queue.Stats(ctx)
		msg, ok, _ := repo.Get(ctx, id)
		if ok && stats.Dead > 0 && msg.Status == message.StatusFailed {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

type fixedRand struct {
	value int
}

func (r fixedRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return r.value % n
}

func newTestIdemStore(t *testing.T) (idempotency.Store, func()) {
	t.Helper()
	dir := t.TempDir()
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: dir})
	if err != nil {
		t.Fatalf("kv store: %v", err)
	}
	store := idempotency.NewKVStore(kv, idempotency.DefaultKVConfig())
	cleanup := func() {
		_ = kv.Close()
	}
	return store, cleanup
}
