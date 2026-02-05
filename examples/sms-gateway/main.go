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
	"github.com/spcent/plumego/core/components/ops"
	"github.com/spcent/plumego/examples/sms-gateway/internal/message"
	"github.com/spcent/plumego/examples/sms-gateway/internal/pipeline"
	"github.com/spcent/plumego/examples/sms-gateway/internal/routing"
	"github.com/spcent/plumego/examples/sms-gateway/internal/tasks"
	logpkg "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/metrics/smsgateway"
	"github.com/spcent/plumego/middleware/observability"
	tenantmw "github.com/spcent/plumego/middleware/tenant"
	"github.com/spcent/plumego/net/mq"
	mqstore "github.com/spcent/plumego/net/mq/store"
	plumrouter "github.com/spcent/plumego/router"
	storedb "github.com/spcent/plumego/store/db"
	"github.com/spcent/plumego/store/idempotency"
	kvstore "github.com/spcent/plumego/store/kv"
	"github.com/spcent/plumego/tenant"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	tenantConfig := tenant.NewInMemoryConfigManager()
	tenantConfig.SetTenantConfig(tenant.Config{
		TenantID: "tenant-1",
		Quota: tenant.QuotaConfig{
			Limits: []tenant.QuotaLimit{
				{Window: tenant.QuotaWindowDay, Requests: 200000},
			},
		},
	})
	tenantConfig.SetTenantConfig(tenant.Config{
		TenantID: "tenant-2",
		Quota: tenant.QuotaConfig{
			Limits: []tenant.QuotaLimit{
				{Window: tenant.QuotaWindowDay, Requests: 200000},
			},
		},
	})
	quotaStore := tenant.NewInMemoryQuotaStore()
	quotaManager := tenant.NewWindowQuotaManager(tenantConfig, quotaStore)

	queueStore := mqstore.NewMemory(mqstore.MemConfig{})
	queue := mq.NewTaskQueue(queueStore, mq.WithQueueMetricsCollector(collector))
	var queueReplayer mqstore.DLQReplayer
	if replayer, ok := any(queueStore).(mqstore.DLQReplayer); ok {
		queueReplayer = replayer
	}

	deduper, dedupeTTL, dedupeCleanup := newTaskDeduper(ctx)
	defer dedupeCleanup()

	worker := mq.NewWorker(queue, mq.WorkerConfig{
		ConsumerID:       "sms-worker",
		Concurrency:      2,
		PollInterval:     100 * time.Millisecond,
		RetryPolicy:      mq.ExponentialBackoff{Base: 200 * time.Millisecond, Max: 2 * time.Second, Factor: 2},
		ShutdownTimeout:  3 * time.Second,
		MetricsCollector: collector,
		Deduper:          deduper,
		DedupeTTL:        dedupeTTL,
	})

	smsMetrics := smsgateway.NewReporter(collector)

	processor := &pipeline.Processor{
		Repo:    repo,
		Router:  router,
		Metrics: smsMetrics,
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

	go reportQueueMetrics(ctx, queue, smsMetrics, 2*time.Second)

	idemStore, cleanup := newIdempotencyStore()
	defer cleanup()

	logger := logpkg.NewGLogger()
	handler := contract.AdaptCtxHandler(
		message.ExampleSendHandler(idemStore, repo, router, queue),
		logger,
	)
	handler = tenantmw.TenantQuota(tenantmw.TenantQuotaOptions{
		Manager: quotaManager,
		Hooks:   tenant.Hooks{},
	})(handler)
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
	mux.Handle("/v1/receipts", contract.AdaptCtxHandler(
		message.ExampleReceiptHandler(repo, smsMetrics),
		logger,
	))

	queueName := "send"
	opsHooks := ops.Hooks{
		QueueStats: func(ctx context.Context, queueID string) (ops.QueueStats, error) {
			if queueID != "" && queueID != queueName {
				return ops.QueueStats{}, fmt.Errorf("unknown queue %q", queueID)
			}
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
				UpdatedAt: time.Now().UTC(),
			}, nil
		},
		QueueList: func(ctx context.Context) ([]string, error) {
			return []string{queueName}, nil
		},
		QueueReplay: func(ctx context.Context, req ops.QueueReplayRequest) (ops.QueueReplayResult, error) {
			if req.Queue != "" && req.Queue != queueName {
				return ops.QueueReplayResult{}, fmt.Errorf("unknown queue %q", req.Queue)
			}
			if queueReplayer == nil {
				return ops.QueueReplayResult{}, fmt.Errorf("queue replay not supported")
			}
			result, err := queueReplayer.ReplayDLQ(ctx, mqstore.ReplayOptions{
				Max:           req.Max,
				ResetAttempts: true,
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
				return ops.ReceiptRecord{
					MessageID: messageID,
					Status:    "not_found",
					Details:   map[string]any{"found": false},
				}, nil
			}
			record := ops.ReceiptRecord{
				MessageID: msg.ID,
				Status:    string(msg.Status),
				Provider:  msg.Provider,
				UpdatedAt: msg.UpdatedAt,
				Details: map[string]any{
					"attempts":     msg.Attempts,
					"max_attempts": msg.MaxAttempts,
				},
			}
			if msg.ProviderMsgID != "" {
				record.Details["provider_message_id"] = msg.ProviderMsgID
			}
			if msg.ReasonCode != "" {
				record.Details["reason_code"] = msg.ReasonCode
			}
			if msg.ReasonDetail != "" {
				record.Details["reason_detail"] = msg.ReasonDetail
			}
			if !msg.NextAttemptAt.IsZero() {
				record.Details["next_attempt_at"] = msg.NextAttemptAt
			}
			if msg.Status == message.StatusDelivered {
				record.DeliveredAt = msg.UpdatedAt
			}
			return record, nil
		},
		TenantQuota: func(ctx context.Context, tenantID string) (ops.TenantQuotaSnapshot, error) {
			return buildTenantQuotaSnapshot(ctx, tenantID, tenantConfig, quotaStore)
		},
	}
	opsComponent := ops.NewComponent(ops.Options{
		Enabled:  true,
		BasePath: "/ops",
		Logger:   logger,
		Auth: ops.AuthConfig{
			Token:         strings.TrimSpace(os.Getenv("SMS_GATEWAY_OPS_TOKEN")),
			AllowInsecure: getenvBool("SMS_GATEWAY_OPS_INSECURE", false),
		},
		Hooks: opsHooks,
	})
	opsRouter := plumrouter.NewRouter()
	opsComponent.RegisterRoutes(opsRouter)
	mux.Handle("/ops", opsRouter)
	mux.Handle("/ops/", opsRouter)

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

func getenvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid duration %s=%q, using default %s", key, value, fallback)
		return fallback
	}
	return parsed
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

func reportQueueMetrics(ctx context.Context, queue *mq.TaskQueue, reporter *smsgateway.Reporter, interval time.Duration) {
	if reporter == nil || queue == nil {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, err := queue.Stats(ctx)
			if err != nil {
				continue
			}
			reporter.RecordQueueDepth(ctx, "send", smsgateway.QueueStats{
				Queued:  stats.Queued,
				Leased:  stats.Leased,
				Dead:    stats.Dead,
				Expired: stats.Expired,
			})
		}
	}
}

func newTaskDeduper(ctx context.Context) (mq.TaskDeduper, time.Duration, func()) {
	dsn := strings.TrimSpace(os.Getenv("SMS_GATEWAY_MQ_DEDUPE_DSN"))
	if dsn == "" {
		return nil, 0, func() {}
	}

	driver := strings.TrimSpace(getenv("SMS_GATEWAY_MQ_DEDUPE_DRIVER", "postgres"))
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("open dedupe db: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		log.Fatalf("ping dedupe db: %v", err)
	}

	ttl := getenvDuration("SMS_GATEWAY_MQ_DEDUPE_TTL", 24*time.Hour)
	cfg := mq.SQLDeduperConfig{
		Dialect:    dedupeDialectFromDriver(driver),
		Table:      strings.TrimSpace(os.Getenv("SMS_GATEWAY_MQ_DEDUPE_TABLE")),
		Prefix:     strings.TrimSpace(getenv("SMS_GATEWAY_MQ_DEDUPE_PREFIX", "mq-dedupe")),
		DefaultTTL: ttl,
	}

	log.Printf("sms-gateway mq deduper: sql driver=%s table=%s", driver, cfg.Table)

	return mq.NewSQLDeduper(db, cfg), ttl, func() {
		_ = db.Close()
	}
}

func dedupeDialectFromDriver(driver string) idempotency.Dialect {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "pgx", "pgxpool", "pq":
		return idempotency.DialectPostgres
	case "mysql", "mariadb":
		return idempotency.DialectMySQL
	default:
		return idempotency.DialectPostgres
	}
}

func buildTenantQuotaSnapshot(ctx context.Context, tenantID string, manager tenant.ConfigManager, store any) (ops.TenantQuotaSnapshot, error) {
	snapshot := ops.TenantQuotaSnapshot{
		TenantID:  tenantID,
		UpdatedAt: time.Now().UTC(),
	}

	if manager == nil {
		snapshot.Details = map[string]any{"configured": false}
		return snapshot, nil
	}

	cfg, err := manager.GetTenantConfig(ctx, tenantID)
	if err != nil {
		if err == tenant.ErrTenantNotFound {
			snapshot.Details = map[string]any{"found": false}
			return snapshot, nil
		}
		return snapshot, err
	}

	limits := normalizeQuotaLimits(cfg.Quota)
	now := time.Now().UTC()

	type usageReader interface {
		Usage(tenantID string, window tenant.QuotaWindow, windowStart time.Time) (tenant.QuotaUsage, bool)
	}
	var reader usageReader
	if store != nil {
		if typed, ok := store.(usageReader); ok {
			reader = typed
		}
	}

	snapshot.Limits = make([]ops.QuotaLimit, 0, len(limits))
	snapshot.Usage = make([]ops.QuotaUsage, 0, len(limits))

	for _, limit := range limits {
		snapshot.Limits = append(snapshot.Limits, ops.QuotaLimit{
			Window:   string(limit.Window),
			Requests: limit.Requests,
			Tokens:   limit.Tokens,
		})

		start, end := quotaWindowBounds(now, limit.Window)
		usage := ops.QuotaUsage{
			Window:      string(limit.Window),
			WindowStart: start,
			WindowEnd:   end,
		}
		if reader != nil {
			if current, ok := reader.Usage(tenantID, limit.Window, start); ok {
				usage.Requests = current.Requests
				usage.Tokens = current.Tokens
			}
		}
		snapshot.Usage = append(snapshot.Usage, usage)
	}

	return snapshot, nil
}

func normalizeQuotaLimits(cfg tenant.QuotaConfig) []tenant.QuotaLimit {
	if len(cfg.Limits) > 0 {
		limits := make([]tenant.QuotaLimit, 0, len(cfg.Limits))
		for _, limit := range cfg.Limits {
			if !isValidQuotaWindow(limit.Window) {
				continue
			}
			limits = append(limits, limit)
		}
		return limits
	}

	if cfg.RequestsPerMinute > 0 || cfg.TokensPerMinute > 0 {
		return []tenant.QuotaLimit{{
			Window:   tenant.QuotaWindowMinute,
			Requests: cfg.RequestsPerMinute,
			Tokens:   cfg.TokensPerMinute,
		}}
	}

	return nil
}

func isValidQuotaWindow(window tenant.QuotaWindow) bool {
	switch window {
	case tenant.QuotaWindowMinute, tenant.QuotaWindowHour, tenant.QuotaWindowDay, tenant.QuotaWindowMonth:
		return true
	default:
		return false
	}
}

func quotaWindowBounds(now time.Time, window tenant.QuotaWindow) (time.Time, time.Time) {
	now = now.UTC()
	switch window {
	case tenant.QuotaWindowHour:
		start := now.Truncate(time.Hour)
		return start, start.Add(time.Hour)
	case tenant.QuotaWindowDay:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return start, start.Add(24 * time.Hour)
	case tenant.QuotaWindowMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return start, start.AddDate(0, 1, 0)
	case tenant.QuotaWindowMinute:
		fallthrough
	default:
		start := now.Truncate(time.Minute)
		return start, start.Add(time.Minute)
	}
}
