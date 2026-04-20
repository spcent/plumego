package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

const defaultOperationTimeout = 10 * time.Second

type ClientConfig struct {
	URI              string
	Database         string
	ConnectTimeout   time.Duration
	OperationTimeout time.Duration
	MaxPoolSize      uint64
	Retention        time.Duration
}

type Client struct {
	raw   *mongo.Client
	store *Store
}

func Connect(ctx context.Context, cfg ClientConfig) (*Client, error) {
	if cfg.URI == "" {
		return nil, errors.New("mongo URI is required")
	}
	if cfg.Database == "" {
		return nil, errors.New("mongo database is required")
	}

	clientOptions := options.Client().ApplyURI(cfg.URI)
	if cfg.ConnectTimeout > 0 {
		clientOptions.SetConnectTimeout(cfg.ConnectTimeout)
	}
	if cfg.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(cfg.MaxPoolSize)
	}

	raw, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}
	if err := raw.Ping(ctx, readpref.Primary()); err != nil {
		_ = raw.Disconnect(ctx)
		return nil, err
	}

	db := raw.Database(cfg.Database)
	if err := EnsureIndexes(ctx, db); err != nil {
		_ = raw.Disconnect(ctx)
		return nil, err
	}

	store, err := NewStore(db, WithOperationTimeout(cfg.OperationTimeout), WithRetention(cfg.Retention))
	if err != nil {
		_ = raw.Disconnect(ctx)
		return nil, err
	}

	return &Client{
		raw:   raw,
		store: store,
	}, nil
}

func (c *Client) Store() *Store {
	if c == nil {
		return nil
	}
	return c.store
}

func (c *Client) Disconnect(ctx context.Context) error {
	if c == nil || c.raw == nil {
		return nil
	}
	return c.raw.Disconnect(ctx)
}

type Store struct {
	collections      Collections
	operationTimeout time.Duration
	retention        time.Duration
	now              func() time.Time
}

type StoreOption func(*Store)

func WithOperationTimeout(timeout time.Duration) StoreOption {
	return func(s *Store) {
		if timeout > 0 {
			s.operationTimeout = timeout
		}
	}
}

func WithClock(now func() time.Time) StoreOption {
	return func(s *Store) {
		if now != nil {
			s.now = now
		}
	}
}

func WithRetention(retention time.Duration) StoreOption {
	return func(s *Store) {
		if retention > 0 {
			s.retention = retention
		}
	}
}

func NewStore(db *mongo.Database, opts ...StoreOption) (*Store, error) {
	collections, err := NewCollections(db)
	if err != nil {
		return nil, err
	}
	store := &Store{
		collections:      collections,
		operationTimeout: defaultOperationTimeout,
		retention:        platformstore.DefaultRetention,
		now:              func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(store)
	}
	return store, nil
}

func (s *Store) operationContext() (context.Context, context.CancelFunc) {
	if s.operationTimeout <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), s.operationTimeout)
}

var _ platformstore.WorkerSnapshotStore = (*Store)(nil)
var _ platformstore.ActiveTaskStore = (*Store)(nil)
var _ platformstore.TaskHistoryStore = (*Store)(nil)
var _ platformstore.WorkerEventStore = (*Store)(nil)
var _ platformstore.AlertStore = (*Store)(nil)
var _ platformstore.QueryStore = (*Store)(nil)
var _ domain.SnapshotStore = (*Store)(nil)
var _ domain.TaskHistoryStore = (*Store)(nil)
var _ domain.WorkerEventStore = (*Store)(nil)
