// Package storage defines the guardus Store contract and constructs a
// concrete backend (memory or sqlite) per Config.
//
// Suite-related methods from upstream gatus are intentionally absent; suites
// are out of scope for v1. Triggered-alert persistence is no-op on the memory
// backend and durable on the sqlite backend.
package storage

import (
	"context"
	"errors"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
	"guardus/internal/storage/common/paging"
	"guardus/internal/storage/memory"
	"guardus/internal/storage/sql"

	"time"
)

// ErrUnknownStorageType is returned when Config.Type is not recognized.
var ErrUnknownStorageType = errors.New("unknown storage type")

// Store is the persistence contract used by guardus.
type Store interface {
	// GetAllEndpointStatuses returns endpoint statuses with paginated event/result windows.
	GetAllEndpointStatuses(params *paging.EndpointStatusParams) ([]*endpoint.Status, error)

	// GetEndpointStatus returns the status for the (group, name) pair.
	GetEndpointStatus(groupName, endpointName string, params *paging.EndpointStatusParams) (*endpoint.Status, error)

	// GetEndpointStatusByKey returns the status for a precomputed endpoint key.
	GetEndpointStatusByKey(key string, params *paging.EndpointStatusParams) (*endpoint.Status, error)

	// GetUptimeByKey returns the uptime ratio (0..1) for the time range.
	GetUptimeByKey(key string, from, to time.Time) (float64, error)

	// GetAverageResponseTimeByKey returns the average response time in ms.
	GetAverageResponseTimeByKey(key string, from, to time.Time) (int, error)

	// GetHourlyAverageResponseTimeByKey returns hourly average response times in ms.
	GetHourlyAverageResponseTimeByKey(key string, from, to time.Time) (map[int64]int, error)

	// InsertEndpointResult records a probe outcome for an endpoint.
	InsertEndpointResult(ep *endpoint.Endpoint, result *endpoint.Result) error

	// DeleteAllEndpointStatusesNotInKeys removes statuses whose keys are not in the keepers list.
	DeleteAllEndpointStatusesNotInKeys(keys []string) int

	// HasEndpointStatusNewerThan reports whether any result is newer than timestamp.
	HasEndpointStatusNewerThan(key string, timestamp time.Time) (bool, error)

	// GetTriggeredEndpointAlert returns whether an alert is currently triggered along with metadata.
	GetTriggeredEndpointAlert(ep *endpoint.Endpoint, a *alert.Alert) (exists bool, resolveKey string, numberOfSuccessesInARow int, err error)

	// UpsertTriggeredEndpointAlert persists or refreshes a triggered alert record.
	UpsertTriggeredEndpointAlert(ep *endpoint.Endpoint, triggeredAlert *alert.Alert) error

	// DeleteTriggeredEndpointAlert removes a triggered alert record.
	DeleteTriggeredEndpointAlert(ep *endpoint.Endpoint, triggeredAlert *alert.Alert) error

	// DeleteAllTriggeredAlertsNotInChecksumsByEndpoint cleans up triggered alerts whose configuration is gone.
	DeleteAllTriggeredAlertsNotInChecksumsByEndpoint(ep *endpoint.Endpoint, checksums []string) int

	// ListEndpointConfigs returns the user-managed endpoint definitions.
	// Endpoints and ExternalEndpoints are returned in separate slices because
	// they have distinct shapes (HTTP probe vs push-based heartbeat).
	ListEndpointConfigs(ctx context.Context) ([]*endpoint.Endpoint, []*endpoint.ExternalEndpoint, error)

	// UpsertEndpointConfig persists ep, replacing any prior row with the same key.
	UpsertEndpointConfig(ctx context.Context, ep *endpoint.Endpoint) error

	// UpsertExternalEndpointConfig persists ee, replacing any prior row with the same key.
	UpsertExternalEndpointConfig(ctx context.Context, ee *endpoint.ExternalEndpoint) error

	// Clear removes everything.
	Clear()

	// Save flushes any in-memory state to its backing store, when applicable.
	Save() error

	// Close releases any underlying resources.
	Close()
}

// Validate interface implementations at compile time.
var (
	_ Store = (*memory.Store)(nil)
	_ Store = (*sql.Store)(nil)
)

// New constructs a Store from the validated Config. Caller owns lifecycle and
// must call Close on shutdown. logger is forwarded to backends that surface
// non-fatal persistence errors (currently only the sqlite backend); it must
// not be nil.
func New(cfg *Config, logger plumelog.StructuredLogger) (Store, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	if err := cfg.ValidateAndSetDefaults(); err != nil {
		return nil, err
	}
	switch cfg.Type {
	case TypeMemory:
		return memory.NewStore(cfg.MaximumNumberOfResults, cfg.MaximumNumberOfEvents), nil
	case TypeSQLite:
		return sql.NewStore("sqlite", cfg.Path, cfg.Caching, cfg.MaximumNumberOfResults, cfg.MaximumNumberOfEvents, logger)
	case TypeMySQL:
		return sql.NewStore("mysql", cfg.Path, cfg.Caching, cfg.MaximumNumberOfResults, cfg.MaximumNumberOfEvents, logger)
	default:
		return nil, ErrUnknownStorageType
	}
}
