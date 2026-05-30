package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/spcent/plumego/health"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/connection"
	"dbadmin/internal/esmanager"
	"dbadmin/internal/mongomanager"
	"dbadmin/internal/redismanager"
)

// sqlChecker implements health.ComponentChecker for SQL connections
type sqlChecker struct {
	manager *dbmanager.Manager
	connID  string
}

func (c *sqlChecker) Name() string {
	return "sql:" + c.connID
}

func (c *sqlChecker) Check(ctx context.Context) error {
	db := c.manager.Get(c.connID)
	if db == nil {
		return fmt.Errorf("connection not found")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return db.PingContext(pingCtx)
}

// redisChecker implements health.ComponentChecker for Redis connections
type redisChecker struct {
	manager *redismanager.Manager
	connID  string
	dbIndex int
}

func (c *redisChecker) Name() string {
	return fmt.Sprintf("redis:%s:db%d", c.connID, c.dbIndex)
}

func (c *redisChecker) Check(ctx context.Context) error {
	client, err := c.manager.Get(c.connID, c.dbIndex)
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return client.Ping(pingCtx).Err()
}

// mongoChecker implements health.ComponentChecker for MongoDB connections
type mongoChecker struct {
	manager *mongomanager.Manager
	connID  string
}

func (c *mongoChecker) Name() string {
	return "mongo:" + c.connID
}

func (c *mongoChecker) Check(ctx context.Context) error {
	client := c.manager.Get(c.connID)
	if client == nil {
		return fmt.Errorf("connection not found")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return client.Ping(pingCtx, nil)
}

// esChecker implements health.ComponentChecker for Elasticsearch connections
type esChecker struct {
	manager *esmanager.Manager
	connID  string
}

func (c *esChecker) Name() string {
	return "es:" + c.connID
}

func (c *esChecker) Check(ctx context.Context) error {
	client := c.manager.Get(c.connID)
	if client == nil {
		return fmt.Errorf("connection not found")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	res, err := client.Info(client.Info.WithContext(pingCtx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("ping failed: %s", res.Status())
	}
	return nil
}

// BuildHealthCheckers creates health checkers from a connection store.
// It iterates all connections and creates appropriate checkers based on driver type.
func BuildHealthCheckers(
	connStore *connection.Store,
	dbMgr *dbmanager.Manager,
	redisMgr *redismanager.Manager,
	mongoMgr *mongomanager.Manager,
	esMgr *esmanager.Manager,
) ([]health.ComponentChecker, error) {
	var checkers []health.ComponentChecker

	connections, err := connStore.List()
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}

	for _, conn := range connections {
		switch conn.Driver {
		case connection.DriverMySQL, connection.DriverSQLite:
			checkers = append(checkers, &sqlChecker{
				manager: dbMgr,
				connID:  conn.ID,
			})
		case connection.DriverRedis:
			checkers = append(checkers, &redisChecker{
				manager: redisMgr,
				connID:  conn.ID,
				dbIndex: conn.RedisDBIndex,
			})
		case connection.DriverMongoDB:
			checkers = append(checkers, &mongoChecker{
				manager: mongoMgr,
				connID:  conn.ID,
			})
		case connection.DriverElasticsearch:
			checkers = append(checkers, &esChecker{
				manager: esMgr,
				connID:  conn.ID,
			})
		}
	}

	return checkers, nil
}
