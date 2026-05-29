package datasource

import (
	"context"
	"encoding/json"
	"fmt"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/esmanager"

	"github.com/elastic/go-elasticsearch/v8"
)

// ESConfig adapts *connection.Connection to the ConnectionConfig interface
// for Elasticsearch connections. Mirrors the SQLConfig / RedisConfig / MongoConfig pattern.
type ESConfig struct {
	Conn *connection.Connection
}

// DSType implements ConnectionConfig.
func (c ESConfig) DSType() DataSourceType { return TypeElasticsearch }

// ESAdapter implements DataSourceDriver for Elasticsearch.
// The ResourceExplorer uses it to list indices in the sidebar.
type ESAdapter struct {
	manager *esmanager.Manager
}

// NewESAdapter creates an ESAdapter backed by the given manager.
func NewESAdapter(mgr *esmanager.Manager) *ESAdapter {
	return &ESAdapter{manager: mgr}
}

// Type returns TypeElasticsearch.
func (a *ESAdapter) Type() DataSourceType { return TypeElasticsearch }

// Test verifies connectivity to the Elasticsearch cluster without caching the client.
func (a *ESAdapter) Test(ctx context.Context, cfg ConnectionConfig) error {
	ec, ok := cfg.(ESConfig)
	if !ok {
		return fmt.Errorf("es adapter: unexpected config type %T", cfg)
	}
	return a.manager.Test(ctx, ec.Conn)
}

// Open connects to the configured Elasticsearch cluster and returns a Session.
func (a *ESAdapter) Open(ctx context.Context, cfg ConnectionConfig) (*Session, error) {
	ec, ok := cfg.(ESConfig)
	if !ok {
		return nil, fmt.Errorf("es adapter: unexpected config type %T", cfg)
	}
	client, err := a.manager.Open(ec.Conn)
	if err != nil {
		return nil, err
	}
	return &Session{
		ProfileID: ec.Conn.ID,
		Type:      TypeElasticsearch,
		Handle:    client,
	}, nil
}

// Close is a no-op: the Manager owns the client lifecycle.
func (a *ESAdapter) Close(_ context.Context, _ *Session) error {
	return nil
}

// ListResources returns ResourceNodes for the given parent.
// A nil parent returns top-level nodes (indices); ES indices don't have children in the resource tree.
func (a *ESAdapter) ListResources(ctx context.Context, sess *Session, parent *ResourceRef) ([]ResourceNode, error) {
	client, ok := sess.Handle.(*elasticsearch.Client)
	if !ok {
		return nil, fmt.Errorf("es adapter: unexpected handle type %T", sess.Handle)
	}

	// Only top-level listing (no parent hierarchy for ES)
	if parent != nil && parent.Path != "" {
		return nil, nil // ES indices don't have children in the resource tree
	}

	return a.listIndices(ctx, client, sess.ProfileID)
}

// InspectResource reports that Elasticsearch resource details are unavailable in the preparation phase.
func (a *ESAdapter) InspectResource(_ context.Context, _ *Session, _ ResourceRef) (*ResourceDetail, error) {
	return nil, fmt.Errorf("es adapter: InspectResource unavailable")
}

// listIndices returns es_index nodes for all user indices.
// System indices (starting with '.') are excluded from the listing.
func (a *ESAdapter) listIndices(ctx context.Context, client *elasticsearch.Client, profileID string) ([]ResourceNode, error) {
	res, err := client.Cat.Indices(client.Cat.Indices.WithFormat("json"))
	if err != nil {
		return nil, fmt.Errorf("list indices: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("list indices: %s", res.Status())
	}

	var indices []struct {
		Index     string `json:"index"`
		Health    string `json:"health"`
		Status    string `json:"status"`
		DocsCount string `json:"docs.count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, fmt.Errorf("decode indices: %w", err)
	}

	var nodes []ResourceNode
	for _, idx := range indices {
		// Skip system indices
		if len(idx.Index) > 0 && idx.Index[0] == '.' {
			continue
		}
		nodes = append(nodes, ResourceNode{
			ID:             fmt.Sprintf("%s:es:index:%s", profileID, idx.Index),
			Name:           idx.Index,
			Type:           NodeESIndex,
			DataSourceType: TypeElasticsearch,
			Path:           idx.Index,
			Meta: map[string]any{
				"health":     idx.Health,
				"status":     idx.Status,
				"docs_count": idx.DocsCount,
			},
		})
	}
	return nodes, nil
}
