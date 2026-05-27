package app

import (
	"context"
	"fmt"

	"workerfleet/internal/domain"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
	"workerfleet/internal/platform/store/memory"
	mongostore "workerfleet/internal/platform/store/mongo"
)

func Bootstrap(ctx context.Context, cfg Config) (*Runtime, error) {
	if cfg.StoreBackend == "" {
		cfg.StoreBackend = StoreBackendMemory
	}
	if cfg.Profile == "" {
		cfg.Profile = DefaultConfig().Profile
	}
	if cfg.Retention <= 0 {
		cfg.Retention = DefaultConfig().Retention
	}
	if cfg.Policy.Status == (domain.StatusPolicy{}) {
		cfg.Policy.Status = defaultStatusPolicyForProfile(cfg.Profile)
	}
	if cfg.Policy.Alert == (domain.AlertPolicy{}) {
		cfg.Policy.Alert = cfg.Policy.Status.AlertPolicy()
	}
	if cfg.Runtime.LoopLeaseTTL <= 0 {
		cfg.Runtime.LoopLeaseTTL = defaultLoopLeaseTTL
	}
	if cfg.Runtime.LoopLeaseOwner == "" {
		cfg.Runtime.LoopLeaseOwner = defaultLoopLeaseOwner()
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	statusPolicy, alertPolicy := configuredPolicies(cfg)
	metrics := workerfleetmetrics.NewCollector()
	metricsObserver := workerfleetmetrics.NewObserver(
		metrics,
		workerfleetmetrics.WithExperimentalMetrics(cfg.Metrics.ExperimentalSeriesEnabled),
	)

	switch cfg.StoreBackend {
	case StoreBackendMemory:
		store := memory.NewStore()
		return buildRuntime(store, func(context.Context) error { return nil }, metrics, metricsObserver, statusPolicy, alertPolicy, nopLoopLease{}), nil
	case StoreBackendMongo:
		client, err := mongostore.Connect(ctx, mongostore.ClientConfig{
			URI:              cfg.Mongo.URI,
			Database:         cfg.Mongo.Database,
			ConnectTimeout:   cfg.Mongo.ConnectTimeout,
			OperationTimeout: cfg.Mongo.OperationTimeout,
			MaxPoolSize:      cfg.Mongo.MaxPoolSize,
			Retention:        cfg.Retention,
		})
		if err != nil {
			return nil, fmt.Errorf("bootstrap mongo store: %w", err)
		}
		store := client.Store()
		lease, err := store.LoopLeaseCoordinator(cfg.Runtime.LoopLeaseOwner, cfg.Runtime.LoopLeaseTTL)
		if err != nil {
			_ = client.Disconnect(ctx)
			return nil, fmt.Errorf("bootstrap mongo loop lease: %w", err)
		}
		return buildRuntime(store, client.Disconnect, metrics, metricsObserver, statusPolicy, alertPolicy, lease), nil
	default:
		return nil, fmt.Errorf("unsupported store backend %q", cfg.StoreBackend)
	}
}

func newRuntime(store runtimeStore, close func(context.Context) error) *Runtime {
	metrics := workerfleetmetrics.NewCollector()
	metricsObserver := workerfleetmetrics.NewObserver(metrics)
	policy := domain.DefaultStatusPolicy()
	alertPolicy := policy.AlertPolicy()
	return buildRuntime(store, close, metrics, metricsObserver, policy, alertPolicy, nopLoopLease{})
}

func buildRuntime(
	store runtimeStore,
	close func(context.Context) error,
	metrics *workerfleetmetrics.Collector,
	metricsObserver *workerfleetmetrics.Observer,
	policy domain.StatusPolicy,
	alertPolicy domain.AlertPolicy,
	lease LoopLeaseCoordinator,
) *Runtime {
	ingest := domain.NewIngestService(
		store,
		store,
		store,
		policy,
		nil,
		domain.WithIngestMetrics(metricsObserver),
	)
	service := NewService(ingest, store)
	loops := NewLoopRunner(store, policy, metricsObserver, metricsObserver)
	alerts := NewAlertRunner(store, policy, alertPolicy, metricsObserver, metricsObserver)
	if lease == nil {
		lease = nopLoopLease{}
	}
	loops.lease = lease
	alerts.lease = lease
	return &Runtime{
		Service: service,
		Metrics: metrics,
		Close:   close,
		shell: runtimeShell{
			ingest: ingest,
			query:  service,
			loops:  loops,
			alerts: alerts,
		},
		Ready: func(ctx context.Context) error {
			if store == nil {
				return fmt.Errorf("workerfleet store is not configured")
			}
			_, err := store.ListCurrentWorkerSnapshots(ctx)
			if err != nil {
				return fmt.Errorf("workerfleet store is not ready: %w", err)
			}
			return nil
		},
	}
}

func (r *Runtime) reportRuntimeError(operation string, err error) {
	if r == nil || r.shell.loops == nil {
		return
	}
	r.shell.loops.reportRuntimeError(operation, err)
}
