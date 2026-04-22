package app

import (
	"context"
	"fmt"

	"workerfleet/internal/domain"
	"workerfleet/internal/handler"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
	platformstore "workerfleet/internal/platform/store"
	"workerfleet/internal/platform/store/memory"
	mongostore "workerfleet/internal/platform/store/mongo"
)

type Runtime struct {
	Service *Service
	Handler *handler.Handler
	Metrics *workerfleetmetrics.Collector
	Close   func(context.Context) error
	Ready   func(context.Context) error
	store   runtimeStore
	policy  domain.StatusPolicy
	metrics *workerfleetmetrics.Observer
}

func Bootstrap(ctx context.Context, cfg Config) (*Runtime, error) {
	if cfg.StoreBackend == "" {
		cfg.StoreBackend = StoreBackendMemory
	}
	if cfg.Retention <= 0 {
		cfg.Retention = DefaultConfig().Retention
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	switch cfg.StoreBackend {
	case StoreBackendMemory:
		store := memory.NewStore()
		return newRuntime(store, func(context.Context) error { return nil }), nil
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
		return newRuntime(store, client.Disconnect), nil
	default:
		return nil, fmt.Errorf("unsupported store backend %q", cfg.StoreBackend)
	}
}

func newRuntime(store runtimeStore, close func(context.Context) error) *Runtime {
	metrics := workerfleetmetrics.NewCollector()
	metricsObserver := workerfleetmetrics.NewObserver(metrics)
	policy := domain.DefaultStatusPolicy()
	ingest := domain.NewIngestService(
		store,
		store,
		store,
		policy,
		nil,
		domain.WithIngestMetrics(metricsObserver),
	)
	service := NewService(ingest, store)
	return &Runtime{
		Service: service,
		Handler: handler.New(service),
		Metrics: metrics,
		Close:   close,
		store:   store,
		policy:  policy,
		metrics: metricsObserver,
		Ready: func(context.Context) error {
			if store == nil {
				return fmt.Errorf("workerfleet store is not configured")
			}
			_, err := store.ListCurrentWorkerSnapshots()
			if err != nil {
				return fmt.Errorf("workerfleet store is not ready: %w", err)
			}
			return nil
		},
	}
}

type runtimeStore interface {
	platformstore.QueryStore
	platformstore.WorkerEventStore
	domain.SnapshotStore
	domain.TaskHistoryStore
	domain.WorkerEventStore
}
