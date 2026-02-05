package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/examples/sms-gateway/internal/message"
	"github.com/spcent/plumego/examples/sms-gateway/internal/pipeline"
	"github.com/spcent/plumego/examples/sms-gateway/internal/routing"
	"github.com/spcent/plumego/examples/sms-gateway/internal/tasks"
	logpkg "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware/observability"
	tenantmw "github.com/spcent/plumego/middleware/tenant"
	"github.com/spcent/plumego/net/mq"
	mqstore "github.com/spcent/plumego/net/mq/store"
	storedb "github.com/spcent/plumego/store/db"
	"github.com/spcent/plumego/store/idempotency"
	kvstore "github.com/spcent/plumego/store/kv"
	"github.com/spcent/plumego/tenant"
)

func main() {
	ctx := context.Background()

	tracer := metrics.NewOpenTelemetryTracer("sms-gateway")
	collector := tracer
	repo, repoCleanup := newMessageRepository(ctx, collector)
	defer repoCleanup()

	routePolicyStore := tenant.NewInMemoryRoutePolicyStore()
	_ = routePolicyStore.SetRoutePolicy(ctx, tenant.RoutePolicy{
		TenantID: "tenant-1",
		Strategy: "direct",
		Payload:  []byte(`{"default_provider":"provider-a"}`),
	})
	_ = routePolicyStore.SetRoutePolicy(ctx, tenant.RoutePolicy{
		TenantID: "tenant-2",
		Strategy: "direct",
		Payload:  []byte(`{"default_provider":"provider-fail"}`),
	})
	routePolicyCache := tenant.NewInMemoryRoutePolicyCache(100, 2*time.Minute)
	routePolicyProvider := tenant.NewCachedRoutePolicyProvider(routePolicyStore, routePolicyCache)
	router := &routing.PolicyRouter{Provider: routePolicyProvider}

	queue := mq.NewTaskQueue(mqstore.NewMemory(mqstore.MemConfig{}), mq.WithQueueMetricsCollector(collector))
	worker := mq.NewWorker(queue, mq.WorkerConfig{
		ConsumerID:      "sms-worker",
		Concurrency:     2,
		PollInterval:    100 * time.Millisecond,
		RetryPolicy:     mq.ExponentialBackoff{Base: 200 * time.Millisecond, Max: 2 * time.Second, Factor: 2},
		ShutdownTimeout: 3 * time.Second,
		MetricsCollector: collector,
	})

	processor := &pipeline.Processor{
		Repo:   repo,
		Router: router,
		Providers: map[string]pipeline.ProviderSender{
			"provider-a":    &pipeline.MockProvider{Name: "provider-a", FailFirstN: 1},
			"provider-b":    &pipeline.MockProvider{Name: "provider-b"},
			"provider-fail": &pipeline.MockProvider{Name: "provider-fail", FailFirstN: 10},
		},
		OnFailure: func(ctx context.Context, msg message.Message, task mq.Task, err error) {
			traceID := contract.TraceIDFromContext(ctx)
			log.Printf("send failed trace=%s msg=%s tenant=%s attempts=%d/%d err=%v", traceID, msg.ID, msg.TenantID, task.Attempts, task.MaxAttempts, err)
		},
		OnDLQ: func(ctx context.Context, msg message.Message, task mq.Task, err error) {
			traceID := contract.TraceIDFromContext(ctx)
			log.Printf("DLQ trace=%s msg=%s tenant=%s attempts=%d reason=%v", traceID, msg.ID, msg.TenantID, task.Attempts, err)
		},
	}

	worker.Register(tasks.SendTopic, processor.Handle)
	worker.Start(ctx)
	defer func() {
		_ = worker.Stop(context.Background())
	}()

	idemStore, cleanup := newIdempotencyStore()
	defer cleanup()

	logger := logpkg.NewGLogger()
	handler := contract.AdaptCtxHandler(
		message.ExampleSendHandler(idemStore, repo, router, queue),
		logger,
	)
	handler = tenantmw.TenantResolver(tenantmw.TenantResolverOptions{
		HeaderName: "X-Tenant-ID",
	})(handler)
	handler = observability.Logging(logger, httpMetricsAdapter{collector: collector}, tracer)(handler)

	addr := getenv("SMS_GATEWAY_ADDR", "127.0.0.1:8089")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/v1/messages", handler)
	server := &http.Server{
		Handler: mux,
	}
	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	baseURL := "http://" + ln.Addr().String()
	log.Printf("sms-gateway listening on %s", baseURL)
	log.Printf("send request: curl -X POST %s/v1/messages -H 'X-Tenant-ID: tenant-1' -H 'Idempotency-Key: demo-1' -d '{\"to\":\"+10000000001\",\"body\":\"hello\"}'", baseURL)
	log.Printf("send failure: curl -X POST %s/v1/messages -H 'X-Tenant-ID: tenant-2' -H 'Idempotency-Key: demo-2' -d '{\"to\":\"+10000000002\",\"body\":\"fail\",\"max_attempts\":2}'", baseURL)

	waitForSignal(ctx, server)
}

func newIdempotencyStore() (idempotency.Store, func()) {
	dir, err := os.MkdirTemp("", "plumego-sms-idem")
	if err != nil {
		log.Fatalf("temp dir: %v", err)
	}
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: dir})
	if err != nil {
		log.Fatalf("kv store: %v", err)
	}
	store := idempotency.NewKVStore(kv, idempotency.DefaultKVConfig())
	cleanup := func() {
		_ = kv.Close()
		_ = os.RemoveAll(dir)
	}
	return store, cleanup
}

func newMessageRepository(ctx context.Context, collector metrics.MetricsCollector) (message.Store, func()) {
	dsn := strings.TrimSpace(os.Getenv("SMS_GATEWAY_MESSAGE_DSN"))
	if dsn == "" {
		return message.NewMemoryRepository(), func() {}
	}

	driver := strings.TrimSpace(getenv("SMS_GATEWAY_MESSAGE_DRIVER", "postgres"))
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("open message db: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		log.Fatalf("ping message db: %v", err)
	}

	var exec message.SQLExecutor
	cleanup := func() {
		_ = db.Close()
	}
	if collector != nil {
		instrumented := storedb.NewInstrumentedDB(db, collector, driver)
		exec = instrumented
		cleanup = func() {
			_ = instrumented.Close()
		}
	} else {
		exec = db
	}

	cfg := message.DefaultSQLConfig()
	cfg.Dialect = sqlDialectFromDriver(driver)
	cfg.Table = getenv("SMS_GATEWAY_MESSAGE_TABLE", cfg.Table)

	log.Printf("sms-gateway message store: sql driver=%s table=%s", driver, cfg.Table)
	return message.NewSQLRepository(exec, cfg), cleanup
}

func sqlDialectFromDriver(driver string) message.SQLDialect {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "pgx", "pgxpool", "pq":
		return message.DialectPostgres
	case "mysql", "mariadb":
		return message.DialectMySQL
	default:
		return message.DialectPostgres
	}
}

type httpMetricsAdapter struct {
	collector metrics.MetricsCollector
}

func (h httpMetricsAdapter) Observe(ctx context.Context, m observability.RequestMetrics) {
	if h.collector == nil {
		return
	}
	h.collector.ObserveHTTP(ctx, m.Method, m.Path, m.Status, m.Bytes, m.Duration)
}

func extractMessageID(data []byte) string {
	var resp struct {
		MessageID string `json:"message_id"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return ""
	}
	return resp.MessageID
}

func sendRequest(baseURL, tenantID, to, body string, maxAttempts int) (string, error) {
	payload := map[string]interface{}{
		"to":   to,
		"body": body,
	}
	if maxAttempts > 0 {
		payload["max_attempts"] = maxAttempts
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimSuffix(baseURL, "/") + "/v1/messages"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", fmt.Sprintf("idem-%d", time.Now().UnixNano()))
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	msgID := extractMessageID(respBody)
	if msgID == "" {
		return "", fmt.Errorf("missing message id in response")
	}
	return msgID, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func waitUntil(ctx context.Context, timeout time.Duration, fn func() bool) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Printf("timeout waiting for completion")
}

func waitForSignal(ctx context.Context, server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
