package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/domain/endpoint"
)

// BootstrapFile is the JSON shape accepted by GUARDUS_BOOTSTRAP_FILE: a
// single document with optional `endpoints` and `external-endpoints` arrays.
// Either may be omitted; absent arrays mean "nothing to import for that
// kind".
type BootstrapFile struct {
	Endpoints         []*endpoint.Endpoint         `json:"endpoints,omitempty"`
	ExternalEndpoints []*endpoint.ExternalEndpoint `json:"external-endpoints,omitempty"`
}

// EndpointUpserter is the subset of storage.Store needed for bootstrap. It is
// declared here (rather than in storage) so this package does not depend on
// the storage package's full interface — that would create an import cycle
// once the storage package depends on config types.
type EndpointUpserter interface {
	UpsertEndpointConfig(ctx context.Context, ep *endpoint.Endpoint) error
	UpsertExternalEndpointConfig(ctx context.Context, ee *endpoint.ExternalEndpoint) error
}

// MaybeBootstrapEndpoints reads path (when non-empty) and upserts every entry
// into store. Missing path is a no-op so the call is safe to make
// unconditionally on startup. Returns the number of endpoints upserted.
//
// Endpoints are NOT validated here — validation runs after the full set
// (bootstrap + DB) is loaded, so a bootstrap file that adds a partial
// endpoint can rely on later defaults from ValidateAndSetDefaults.
func MaybeBootstrapEndpoints(ctx context.Context, store EndpointUpserter, path string, logger plumelog.StructuredLogger) (int, error) {
	if path == "" {
		return 0, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("MaybeBootstrapEndpoints: bootstrap file not found; skipping", plumelog.Fields{"path": path})
			return 0, nil
		}
		return 0, fmt.Errorf("read bootstrap file %s: %w", path, err)
	}
	var doc BootstrapFile
	if err := json.Unmarshal(data, &doc); err != nil {
		return 0, fmt.Errorf("parse bootstrap file %s: %w", path, err)
	}
	count := 0
	for _, ep := range doc.Endpoints {
		if err := store.UpsertEndpointConfig(ctx, ep); err != nil {
			return count, fmt.Errorf("upsert endpoint %s: %w", ep.Key(), err)
		}
		count++
	}
	for _, ee := range doc.ExternalEndpoints {
		if err := store.UpsertExternalEndpointConfig(ctx, ee); err != nil {
			return count, fmt.Errorf("upsert external endpoint %s: %w", ee.Key(), err)
		}
		count++
	}
	logger.Info("MaybeBootstrapEndpoints: imported endpoints", plumelog.Fields{"path": path, "count": count})
	return count, nil
}
