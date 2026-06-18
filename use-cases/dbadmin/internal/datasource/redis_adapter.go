package datasource

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/redismanager"

	"github.com/redis/go-redis/v9"
)

// RedisConfig adapts *connection.Connection to the ConnectionConfig interface
// for Redis connections. Mirrors the SQLConfig pattern in sql_adapter.go.
type RedisConfig struct {
	Conn *connection.Connection
}

// DSType implements ConnectionConfig.
func (r RedisConfig) DSType() DataSourceType { return TypeRedis }

// RedisAdapter implements DataSourceDriver for Redis.
// The ResourceExplorer uses it to list DBs (0-15) in the sidebar.
type RedisAdapter struct {
	manager        *redismanager.Manager
	timeoutSeconds int
}

// NewRedisAdapter creates a RedisAdapter backed by the given manager.
func NewRedisAdapter(mgr *redismanager.Manager, timeoutSeconds int) *RedisAdapter {
	return &RedisAdapter{manager: mgr, timeoutSeconds: timeoutSeconds}
}

// Type returns TypeRedis.
func (a *RedisAdapter) Type() DataSourceType { return TypeRedis }

// Test pings the Redis server without caching the client.
func (a *RedisAdapter) Test(ctx context.Context, cfg ConnectionConfig) error {
	rc, ok := cfg.(RedisConfig)
	if !ok {
		return fmt.Errorf("redis adapter: unexpected config type %T", cfg)
	}
	return a.manager.Test(ctx, rc.Conn)
}

// Open connects to the configured Redis DB and returns a Session.
func (a *RedisAdapter) Open(ctx context.Context, cfg ConnectionConfig) (*Session, error) {
	rc, ok := cfg.(RedisConfig)
	if !ok {
		return nil, fmt.Errorf("redis adapter: unexpected config type %T", cfg)
	}
	dbIndex := rc.Conn.RedisDBIndex
	cl, err := a.manager.Open(rc.Conn, dbIndex)
	if err != nil {
		return nil, err
	}
	return &Session{
		ProfileID: rc.Conn.ID,
		Type:      TypeRedis,
		Handle:    cl,
	}, nil
}

// Close is a no-op: the Manager owns the client lifecycle.
func (a *RedisAdapter) Close(_ context.Context, _ *Session) error {
	return nil
}

// ListResources returns a ResourceNode for each Redis database (0-15) with
// their key counts. The sidebar renders these as db0, db1, … db15.
func (a *RedisAdapter) ListResources(ctx context.Context, sess *Session, parent *ResourceRef) ([]ResourceNode, error) {
	cl, ok := sess.Handle.(redis.UniversalClient)
	if !ok {
		return nil, fmt.Errorf("redis adapter: unexpected handle type %T", sess.Handle)
	}

	ctx, cancel := context.WithTimeout(ctx, timeoutDuration(a.timeoutSeconds, 30*time.Second))
	defer cancel()

	dbCount := 16
	cfgRes, err := cl.ConfigGet(ctx, "databases").Result()
	if err == nil {
		if v, ok := cfgRes["databases"]; ok {
			if n, err2 := strconv.Atoi(v); err2 == nil && n > 0 {
				dbCount = n
			}
		}
	}

	nodes := make([]ResourceNode, dbCount)
	for i := range dbCount {
		path := strconv.Itoa(i)
		nodes[i] = ResourceNode{
			ID:   fmt.Sprintf("%s:redis:db:%d", sess.ProfileID, i),
			Name: fmt.Sprintf("db%d", i),
			Path: path,
			Type: NodeRedisDB,
		}
	}
	return nodes, nil
}

// InspectResource reports that Redis resource details are unavailable.
func (a *RedisAdapter) InspectResource(_ context.Context, _ *Session, _ ResourceRef) (*ResourceDetail, error) {
	return nil, fmt.Errorf("redis adapter: InspectResource unavailable")
}
