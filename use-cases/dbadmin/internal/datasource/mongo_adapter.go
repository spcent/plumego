package datasource

import (
	"context"
	"fmt"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/mongomanager"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoConfig adapts *connection.Connection to the ConnectionConfig interface
// for MongoDB connections. Mirrors the SQLConfig / RedisConfig pattern.
type MongoConfig struct {
	Conn *connection.Connection
}

// DSType implements ConnectionConfig.
func (c MongoConfig) DSType() DataSourceType { return TypeMongoDB }

// MongoAdapter implements DataSourceDriver for MongoDB.
// The ResourceExplorer uses it to list databases and collections in the sidebar.
type MongoAdapter struct {
	manager *mongomanager.Manager
}

// NewMongoAdapter creates a MongoAdapter backed by the given manager.
func NewMongoAdapter(mgr *mongomanager.Manager) *MongoAdapter {
	return &MongoAdapter{manager: mgr}
}

// Type returns TypeMongoDB.
func (a *MongoAdapter) Type() DataSourceType { return TypeMongoDB }

// Test verifies connectivity to the MongoDB server without caching the client.
func (a *MongoAdapter) Test(ctx context.Context, cfg ConnectionConfig) error {
	mc, ok := cfg.(MongoConfig)
	if !ok {
		return fmt.Errorf("mongo adapter: unexpected config type %T", cfg)
	}
	return a.manager.Test(ctx, mc.Conn)
}

// Open connects to the configured MongoDB server and returns a Session.
func (a *MongoAdapter) Open(ctx context.Context, cfg ConnectionConfig) (*Session, error) {
	mc, ok := cfg.(MongoConfig)
	if !ok {
		return nil, fmt.Errorf("mongo adapter: unexpected config type %T", cfg)
	}
	client, err := a.manager.Open(mc.Conn)
	if err != nil {
		return nil, err
	}
	return &Session{
		ProfileID: mc.Conn.ID,
		Type:      TypeMongoDB,
		Handle:    client,
	}, nil
}

// Close is a no-op: the Manager owns the client lifecycle.
func (a *MongoAdapter) Close(_ context.Context, _ *Session) error {
	return nil
}

// ListResources returns ResourceNodes for the given parent.
// A nil parent returns top-level nodes (databases); a database path returns collections.
func (a *MongoAdapter) ListResources(ctx context.Context, sess *Session, parent *ResourceRef) ([]ResourceNode, error) {
	client, ok := sess.Handle.(*mongo.Client)
	if !ok {
		return nil, fmt.Errorf("mongo adapter: unexpected handle type %T", sess.Handle)
	}

	if parent == nil || parent.Path == "" {
		// List databases
		return a.listDatabases(ctx, client, sess.ProfileID)
	}

	// List collections under a database
	return a.listCollections(ctx, client, sess.ProfileID, parent.Path)
}

// InspectResource reports that MongoDB resource details are unavailable in the preparation phase.
func (a *MongoAdapter) InspectResource(_ context.Context, _ *Session, _ ResourceRef) (*ResourceDetail, error) {
	return nil, fmt.Errorf("mongo adapter: InspectResource unavailable")
}

// listDatabases returns mongo_database nodes for all user databases.
// System databases (admin, local, config) are excluded from the listing.
func (a *MongoAdapter) listDatabases(ctx context.Context, client *mongo.Client, profileID string) ([]ResourceNode, error) {
	result, err := client.ListDatabases(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}

	systemDBs := map[string]bool{
		"admin":  true,
		"local":  true,
		"config": true,
	}

	var nodes []ResourceNode
	for _, db := range result.Databases {
		if systemDBs[db.Name] {
			continue
		}
		nodes = append(nodes, ResourceNode{
			ID:             fmt.Sprintf("%s:mongo:db:%s", profileID, db.Name),
			Name:           db.Name,
			Type:           NodeMongoDatabase,
			DataSourceType: TypeMongoDB,
			Path:           db.Name,
			Meta: map[string]any{
				"size_on_disk": db.SizeOnDisk,
				"empty":        db.Empty,
			},
		})
	}
	return nodes, nil
}

// listCollections returns mongo_collection nodes for the given database.
func (a *MongoAdapter) listCollections(ctx context.Context, client *mongo.Client, profileID, dbName string) ([]ResourceNode, error) {
	db := client.Database(dbName)

	cursor, err := db.ListCollections(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var nodes []ResourceNode
	for cursor.Next(ctx) {
		var coll struct {
			Name string `bson:"name"`
			Type string `bson:"type"` // "collection" or "view"
		}
		if err := cursor.Decode(&coll); err != nil {
			continue
		}
		// Skip system collections (e.g., system.views)
		if len(coll.Name) > 7 && coll.Name[:7] == "system." {
			continue
		}
		nodes = append(nodes, ResourceNode{
			ID:             fmt.Sprintf("%s:mongo:coll:%s/%s", profileID, dbName, coll.Name),
			Name:           coll.Name,
			Type:           NodeMongoCollection,
			ParentID:       dbName,
			DataSourceType: TypeMongoDB,
			Path:           fmt.Sprintf("%s/%s", dbName, coll.Name),
			Meta: map[string]any{
				"collection_type": coll.Type,
			},
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor iteration: %w", err)
	}
	return nodes, nil
}
