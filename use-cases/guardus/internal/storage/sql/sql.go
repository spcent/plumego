// Package sql implements the guardus Store on top of database/sql, supporting
// SQLite (via modernc.org/sqlite) and MySQL (via go-sql-driver/mysql).
// The schema is embedded and created on first open. Dialect differences
// (placeholders, upsert syntax, RETURNING vs LastInsertId) are abstracted
// behind the dialect interface.
//
// Suite-related code from upstream gatus has been removed; only endpoint-scoped
// data is persisted.
package sql

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	databasesql "database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"

	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/store/db"

	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
	"guardus/internal/domain/key"
	"guardus/internal/storage/common"
	"guardus/internal/storage/common/paging"
)

//////////////////////////////////////////////////////////////////////////////////////////////////
// Note that only exported functions in this file may create, commit, or rollback a transaction //
//////////////////////////////////////////////////////////////////////////////////////////////////

const (
	// arraySeparator joins multiple error strings into one column value.
	arraySeparator = "|~|"

	eventsAboveMaximumCleanUpThreshold  = 10
	resultsAboveMaximumCleanUpThreshold = 10

	uptimeTotalEntriesMergeThreshold = 100
	uptimeAgeCleanUpThreshold        = 32 * 24 * time.Hour
	uptimeRetention                  = 30 * 24 * time.Hour
	uptimeHourlyBuffer               = 48 * time.Hour
)

//go:embed schema_sqlite.sql
var schemaSQLite string

//go:embed schema_mysql.sql
var schemaMySQL string

var (
	// ErrPathNotSpecified is returned when the path parameter is empty.
	ErrPathNotSpecified = errors.New("path cannot be empty")

	// ErrDatabaseDriverNotSpecified is returned when the driver parameter is empty.
	ErrDatabaseDriverNotSpecified = errors.New("database driver cannot be empty")

	errNoRowsReturned = errors.New("expected a row to be returned, but none was")
)

// Store persists endpoint statuses, results, events, uptime, and triggered
// alerts to a SQL database.
type Store struct {
	driver, path string

	db  *databasesql.DB
	log plumelog.StructuredLogger

	// dialect abstracts SQL syntax differences between backends.
	d dialect
	// schemaSQL holds the embedded DDL for the active backend.
	schemaSQL string

	// writeThroughCache (when caching is enabled) maps a generated cache key
	// to a *endpoint.Status to avoid round-tripping the database for repeated
	// reads. Entries are invalidated on InsertEndpointResult and on Clear.
	cacheEnabled bool
	cacheMu      sync.RWMutex
	cache        map[string]*endpoint.Status

	maximumNumberOfResults int
	maximumNumberOfEvents  int
}

// NewStore opens path with the given driver, creates the schema if necessary,
// and returns a *Store. Supported drivers: "sqlite" and "mysql". For sqlite,
// path is a file path; for mysql, path is a DSN string.
// logger must not be nil.
func NewStore(driver, path string, caching bool, maximumNumberOfResults, maximumNumberOfEvents int, logger plumelog.StructuredLogger) (*Store, error) {
	if len(driver) == 0 {
		return nil, ErrDatabaseDriverNotSpecified
	}
	if len(path) == 0 {
		return nil, ErrPathNotSpecified
	}

	var d dialect
	var schema string
	switch driver {
	case "sqlite":
		d = sqliteDialect{}
		schema = schemaSQLite
	case "mysql":
		d = mysqlDialect{}
		schema = schemaMySQL
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	cfg := db.DefaultConfig(driver, path)
	if driver == "sqlite" {
		// SQLite + WAL needs a single writer; one connection avoids "database is locked".
		cfg.MaxOpenConns = 1
		cfg.MaxIdleConns = 1
	}
	handle, err := db.Open(cfg)
	if err != nil {
		return nil, err
	}
	if err := handle.Ping(); err != nil {
		_ = handle.Close()
		return nil, err
	}
	if err := d.initPragmas(handle); err != nil {
		_ = handle.Close()
		return nil, err
	}
	store := &Store{
		driver:                 driver,
		path:                   path,
		db:                     handle,
		log:                    logger,
		d:                      d,
		schemaSQL:              schema,
		maximumNumberOfResults: maximumNumberOfResults,
		maximumNumberOfEvents:  maximumNumberOfEvents,
	}
	if err := store.createSchema(); err != nil {
		_ = handle.Close()
		return nil, err
	}
	if caching {
		store.cacheEnabled = true
		store.cache = make(map[string]*endpoint.Status)
	}
	return store, nil
}

// createSchema runs the embedded schema inside a single transaction.
func (s *Store) createSchema() error {
	return db.WithTransaction(context.Background(), s.db, nil, func(tx *databasesql.Tx) error {
		for _, stmt := range splitSQLStatements(s.schemaSQL) {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				return fmt.Errorf("apply schema statement: %w", err)
			}
		}
		return nil
	})
}

func splitSQLStatements(s string) []string {
	parts := strings.Split(s, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return out
}

// GetAllEndpointStatuses returns every endpoint's status, with paginated
// events/results.
func (s *Store) GetAllEndpointStatuses(params *paging.EndpointStatusParams) ([]*endpoint.Status, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	keys, err := s.getAllEndpointKeys(tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	out := make([]*endpoint.Status, 0, len(keys))
	for _, k := range keys {
		status, err := s.getEndpointStatusByKey(tx, k, params)
		if err != nil {
			continue
		}
		out = append(out, status)
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
	}
	return out, err
}

// GetEndpointStatus returns the status for the (group, name) pair.
func (s *Store) GetEndpointStatus(groupName, endpointName string, params *paging.EndpointStatusParams) (*endpoint.Status, error) {
	return s.GetEndpointStatusByKey(key.ConvertGroupAndNameToKey(groupName, endpointName), params)
}

// GetEndpointStatusByKey returns the status for the given key.
func (s *Store) GetEndpointStatusByKey(k string, params *paging.EndpointStatusParams) (*endpoint.Status, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	status, err := s.getEndpointStatusByKey(tx, k, params)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
	}
	return status, err
}

// GetUptimeByKey returns the uptime ratio over [from, to].
func (s *Store) GetUptimeByKey(k string, from, to time.Time) (float64, error) {
	if from.After(to) {
		return 0, common.ErrInvalidTimeRange
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	endpointID, _, _, err := s.getEndpointIDGroupAndNameByKey(tx, k)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	uptime, _, err := s.getEndpointUptime(tx, endpointID, from, to)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
	}
	return uptime, nil
}

// GetAverageResponseTimeByKey returns the average response time in ms over [from, to].
func (s *Store) GetAverageResponseTimeByKey(k string, from, to time.Time) (int, error) {
	if from.After(to) {
		return 0, common.ErrInvalidTimeRange
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	endpointID, _, _, err := s.getEndpointIDGroupAndNameByKey(tx, k)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	avg, err := s.getEndpointAverageResponseTime(tx, endpointID, from, to)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
	}
	return avg, nil
}

// GetHourlyAverageResponseTimeByKey returns hourly average response times in ms.
func (s *Store) GetHourlyAverageResponseTimeByKey(k string, from, to time.Time) (map[int64]int, error) {
	if from.After(to) {
		return nil, common.ErrInvalidTimeRange
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	endpointID, _, _, err := s.getEndpointIDGroupAndNameByKey(tx, k)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	out, err := s.getEndpointHourlyAverageResponseTimes(tx, endpointID, from, to)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
	}
	return out, nil
}

// InsertEndpointResult records a probe outcome and updates events/uptime.
func (s *Store) InsertEndpointResult(ep *endpoint.Endpoint, result *endpoint.Result) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	endpointID, err := s.getEndpointID(tx, ep)
	if err != nil {
		if errors.Is(err, common.ErrEndpointNotFound) {
			if endpointID, err = s.insertEndpoint(tx, ep); err != nil {
				_ = tx.Rollback()
				s.log.Error("InsertEndpointResult: failed to create endpoint", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
				return err
			}
		} else {
			_ = tx.Rollback()
			s.log.Error("InsertEndpointResult: failed to retrieve endpoint id", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
			return err
		}
	}
	numberOfEvents, err := s.getNumberOfEventsByEndpointID(tx, endpointID)
	if err != nil {
		s.log.Error("InsertEndpointResult: failed to retrieve total number of events", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
	}
	if numberOfEvents == 0 {
		err = s.insertEndpointEvent(tx, endpointID, &endpoint.Event{
			Type:      endpoint.EventStart,
			Timestamp: result.Timestamp.Add(-50 * time.Millisecond),
		})
		if err != nil {
			s.log.Error("InsertEndpointResult: failed to insert event", plumelog.Fields{"op": "sql.InsertEndpointResult", "event": string(endpoint.EventStart), "key": ep.Key(), "err": err.Error()})
		}
		event := endpoint.NewEventFromResult(result)
		if err = s.insertEndpointEvent(tx, endpointID, event); err != nil {
			s.log.Error("InsertEndpointResult: failed to insert event", plumelog.Fields{"op": "sql.InsertEndpointResult", "event": string(event.Type), "key": ep.Key(), "err": err.Error()})
		}
	} else {
		var lastResultSuccess bool
		if lastResultSuccess, err = s.getLastEndpointResultSuccessValue(tx, endpointID); err != nil {
			s.log.Error("InsertEndpointResult: failed to retrieve outcome of previous result", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
		} else if lastResultSuccess != result.Success {
			event := endpoint.NewEventFromResult(result)
			if err = s.insertEndpointEvent(tx, endpointID, event); err != nil {
				s.log.Error("InsertEndpointResult: failed to insert event", plumelog.Fields{"op": "sql.InsertEndpointResult", "event": string(event.Type), "key": ep.Key(), "err": err.Error()})
			}
		}
		if numberOfEvents > int64(s.maximumNumberOfEvents+eventsAboveMaximumCleanUpThreshold) {
			if err = s.deleteOldEndpointEvents(tx, endpointID); err != nil {
				s.log.Error("InsertEndpointResult: failed to delete old events", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
			}
		}
	}
	if err = s.insertEndpointResult(tx, endpointID, result); err != nil {
		s.log.Error("InsertEndpointResult: failed to insert result", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
		_ = tx.Rollback()
		return err
	}
	numberOfResults, err := s.getNumberOfResultsByEndpointID(tx, endpointID)
	if err != nil {
		s.log.Error("InsertEndpointResult: failed to retrieve total number of results", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
	} else if numberOfResults > int64(s.maximumNumberOfResults+resultsAboveMaximumCleanUpThreshold) {
		if err = s.deleteOldEndpointResults(tx, endpointID); err != nil {
			s.log.Error("InsertEndpointResult: failed to delete old results", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
		}
	}
	if err = s.updateEndpointUptime(tx, endpointID, result); err != nil {
		s.log.Error("InsertEndpointResult: failed to update uptime", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
	}
	numberOfUptimeEntries, err := s.getNumberOfUptimeEntriesByEndpointID(tx, endpointID)
	if err != nil {
		s.log.Error("InsertEndpointResult: failed to retrieve total number of uptime entries", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
	} else if numberOfUptimeEntries >= uptimeTotalEntriesMergeThreshold {
		s.log.Info("InsertEndpointResult: merging hourly uptime entries", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key()})
		if err = s.mergeHourlyUptimeEntriesOlderThanMergeThresholdIntoDailyUptimeEntries(tx, endpointID); err != nil {
			s.log.Error("InsertEndpointResult: failed to merge hourly uptime entries", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
		}
	}
	ageOfOldestUptimeEntry, err := s.getAgeOfOldestEndpointUptimeEntry(tx, endpointID)
	if err != nil {
		if !errors.Is(err, errNoRowsReturned) {
			s.log.Error("InsertEndpointResult: failed to retrieve oldest endpoint uptime entry", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
		}
	} else if ageOfOldestUptimeEntry > uptimeAgeCleanUpThreshold {
		if err = s.deleteOldUptimeEntries(tx, endpointID, time.Now().Add(-(uptimeRetention + time.Hour))); err != nil {
			s.log.Error("InsertEndpointResult: failed to delete old uptime entries", plumelog.Fields{"op": "sql.InsertEndpointResult", "key": ep.Key(), "err": err.Error()})
		}
	}
	if s.cacheEnabled {
		s.invalidateCacheByPrefix(ep.Key())
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
	}
	return err
}

// DeleteAllEndpointStatusesNotInKeys deletes endpoints (and cascading rows)
// whose endpoint_key is not in keys.
func (s *Store) DeleteAllEndpointStatusesNotInKeys(keys []string) int {
	var (
		err    error
		result databasesql.Result
	)
	if len(keys) == 0 {
		result, err = s.db.Exec("DELETE FROM endpoints")
	} else {
		args := make([]interface{}, 0, len(keys))
		query := "DELETE FROM endpoints WHERE endpoint_key NOT IN ("
		query += s.d.placeholders(1, len(keys))
		query += ")"
		for i := range keys {
			args = append(args, keys[i])
		}
		result, err = s.db.Exec(query, args...)
	}
	if err != nil {
		s.log.Error("DeleteAllEndpointStatusesNotInKeys: failed to delete endpoints", plumelog.Fields{"op": "sql.DeleteAllEndpointStatusesNotInKeys", "err": err.Error()})
		return 0
	}
	if s.cacheEnabled {
		s.cacheMu.Lock()
		s.cache = make(map[string]*endpoint.Status)
		s.cacheMu.Unlock()
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected)
}

// GetTriggeredEndpointAlert reports whether a triggered alert exists.
func (s *Store) GetTriggeredEndpointAlert(ep *endpoint.Endpoint, a *alert.Alert) (exists bool, resolveKey string, numberOfSuccessesInARow int, err error) {
	endpointID, err := s.lookupEndpointID(ep.Key())
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) || errors.Is(err, common.ErrEndpointNotFound) {
			return false, "", 0, nil
		}
		return false, "", 0, err
	}
	err = s.db.QueryRow(
		fmt.Sprintf("SELECT resolve_key, number_of_successes_in_a_row FROM endpoint_alerts_triggered WHERE endpoint_id = %s AND configuration_checksum = %s",
			s.d.placeholder(1), s.d.placeholder(2)),
		endpointID,
		a.Checksum(),
	).Scan(&resolveKey, &numberOfSuccessesInARow)
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) {
			return false, "", 0, nil
		}
		return false, "", 0, err
	}
	return true, resolveKey, numberOfSuccessesInARow, nil
}

// UpsertTriggeredEndpointAlert persists a triggered alert.
func (s *Store) UpsertTriggeredEndpointAlert(ep *endpoint.Endpoint, triggeredAlert *alert.Alert) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	endpointID, err := s.getEndpointID(tx, ep)
	if err != nil {
		if errors.Is(err, common.ErrEndpointNotFound) {
			if endpointID, err = s.insertEndpoint(tx, ep); err != nil {
				_ = tx.Rollback()
				s.log.Error("UpsertTriggeredEndpointAlert: failed to create endpoint", plumelog.Fields{"op": "sql.UpsertTriggeredEndpointAlert", "key": ep.Key(), "err": err.Error()})
				return err
			}
		} else {
			_ = tx.Rollback()
			s.log.Error("UpsertTriggeredEndpointAlert: failed to retrieve endpoint id", plumelog.Fields{"op": "sql.UpsertTriggeredEndpointAlert", "key": ep.Key(), "err": err.Error()})
			return err
		}
	}
	upsertClause := s.d.upsertClause("endpoint_id, configuration_checksum", []string{"resolve_key", "number_of_successes_in_a_row"})
	_, err = tx.Exec(
		fmt.Sprintf(
			`INSERT INTO endpoint_alerts_triggered (endpoint_id, configuration_checksum, resolve_key, number_of_successes_in_a_row)
			 VALUES (%s, %s, %s, %s) %s`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3), s.d.placeholder(4),
			upsertClause,
		),
		endpointID,
		triggeredAlert.Checksum(),
		triggeredAlert.ResolveKey,
		ep.NumberOfSuccessesInARow,
	)
	if err != nil {
		_ = tx.Rollback()
		s.log.Error("UpsertTriggeredEndpointAlert: failed to persist triggered alert", plumelog.Fields{"op": "sql.UpsertTriggeredEndpointAlert", "key": ep.Key(), "err": err.Error()})
		return err
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
	}
	return nil
}

// DeleteTriggeredEndpointAlert removes a triggered alert.
func (s *Store) DeleteTriggeredEndpointAlert(ep *endpoint.Endpoint, triggeredAlert *alert.Alert) error {
	endpointID, err := s.lookupEndpointID(ep.Key())
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) || errors.Is(err, common.ErrEndpointNotFound) {
			return nil
		}
		return err
	}
	_, err = s.db.Exec(
		fmt.Sprintf("DELETE FROM endpoint_alerts_triggered WHERE configuration_checksum = %s AND endpoint_id = %s",
			s.d.placeholder(1), s.d.placeholder(2)),
		triggeredAlert.Checksum(), endpointID,
	)
	return err
}

// DeleteAllTriggeredAlertsNotInChecksumsByEndpoint cleans up triggered alerts whose
// configuration is no longer present.
func (s *Store) DeleteAllTriggeredAlertsNotInChecksumsByEndpoint(ep *endpoint.Endpoint, checksums []string) int {
	endpointID, err := s.lookupEndpointID(ep.Key())
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) || errors.Is(err, common.ErrEndpointNotFound) {
			return 0
		}
		s.log.Error("DeleteAllTriggeredAlertsNotInChecksumsByEndpoint: lookup failed", plumelog.Fields{"op": "sql.DeleteAllTriggeredAlertsNotInChecksumsByEndpoint", "key": ep.Key(), "err": err.Error()})
		return 0
	}
	var (
		dbErr  error
		result databasesql.Result
	)
	if len(checksums) == 0 {
		result, dbErr = s.db.Exec(
			fmt.Sprintf("DELETE FROM endpoint_alerts_triggered WHERE endpoint_id = %s",
				s.d.placeholder(1)),
			endpointID,
		)
	} else {
		args := make([]interface{}, 0, len(checksums)+1)
		args = append(args, endpointID)
		query := `DELETE FROM endpoint_alerts_triggered
			WHERE endpoint_id = ` + s.d.placeholder(1) + `
			  AND configuration_checksum NOT IN (` + s.d.placeholders(2, len(checksums)) + `)`
		for i := range checksums {
			args = append(args, checksums[i])
		}
		result, dbErr = s.db.Exec(query, args...)
	}
	if dbErr != nil {
		s.log.Error("DeleteAllTriggeredAlertsNotInChecksumsByEndpoint: failed", plumelog.Fields{"op": "sql.DeleteAllTriggeredAlertsNotInChecksumsByEndpoint", "key": ep.Key(), "err": dbErr.Error()})
		return 0
	}
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected)
}

// HasEndpointStatusNewerThan reports whether a result newer than ts exists.
func (s *Store) HasEndpointStatusNewerThan(k string, ts time.Time) (bool, error) {
	if ts.IsZero() {
		return false, errors.New("timestamp is zero")
	}
	var count int
	err := s.db.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM endpoint_results WHERE endpoint_id = (SELECT endpoint_id FROM endpoints WHERE endpoint_key = %s LIMIT 1) AND timestamp > %s",
			s.d.placeholder(1), s.d.placeholder(2)),
		k,
		ts.UTC(),
	).Scan(&count)
	if err != nil {
		return false, nil
	}
	return count > 0, nil
}

// Clear deletes everything.
func (s *Store) Clear() {
	_, _ = s.db.Exec("DELETE FROM endpoints")
	if s.cacheEnabled {
		s.cacheMu.Lock()
		s.cache = make(map[string]*endpoint.Status)
		s.cacheMu.Unlock()
	}
}

// ListEndpointConfigs returns the persisted endpoint definitions, separating
// regular Endpoints from ExternalEndpoints by the is_external flag.
func (s *Store) ListEndpointConfigs(ctx context.Context) ([]*endpoint.Endpoint, []*endpoint.ExternalEndpoint, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT endpoint_key, is_external, payload_json FROM endpoint_configs ORDER BY endpoint_key",
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var endpoints []*endpoint.Endpoint
	var external []*endpoint.ExternalEndpoint
	for rows.Next() {
		var (
			k          string
			isExternal int
			payload    string
		)
		if err := rows.Scan(&k, &isExternal, &payload); err != nil {
			return nil, nil, err
		}
		if isExternal == 1 {
			ee := &endpoint.ExternalEndpoint{}
			if err := json.Unmarshal([]byte(payload), ee); err != nil {
				s.log.Warn("ListEndpointConfigs: skipping malformed external endpoint", plumelog.Fields{"endpoint_key": k, "err": err.Error()})
				continue
			}
			external = append(external, ee)
			continue
		}
		ep := &endpoint.Endpoint{}
		if err := json.Unmarshal([]byte(payload), ep); err != nil {
			s.log.Warn("ListEndpointConfigs: skipping malformed endpoint", plumelog.Fields{"endpoint_key": k, "err": err.Error()})
			continue
		}
		endpoints = append(endpoints, ep)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return endpoints, external, nil
}

// UpsertEndpointConfig persists ep keyed by ep.Key().
func (s *Store) UpsertEndpointConfig(ctx context.Context, ep *endpoint.Endpoint) error {
	if ep == nil {
		return errors.New("endpoint is nil")
	}
	payload, err := json.Marshal(ep)
	if err != nil {
		return fmt.Errorf("marshal endpoint: %w", err)
	}
	return s.upsertEndpointConfig(ctx, ep.Key(), 0, payload)
}

// UpsertExternalEndpointConfig persists ee keyed by ee.Key().
func (s *Store) UpsertExternalEndpointConfig(ctx context.Context, ee *endpoint.ExternalEndpoint) error {
	if ee == nil {
		return errors.New("external endpoint is nil")
	}
	payload, err := json.Marshal(ee)
	if err != nil {
		return fmt.Errorf("marshal external endpoint: %w", err)
	}
	return s.upsertEndpointConfig(ctx, ee.Key(), 1, payload)
}

func (s *Store) upsertEndpointConfig(ctx context.Context, k string, isExternal int, payload []byte) error {
	upsertClause := s.d.upsertClause("endpoint_key", []string{"is_external", "payload_json", "updated_at"})
	return db.WithTransaction(ctx, s.db, nil, func(tx *databasesql.Tx) error {
		_, err := tx.ExecContext(ctx,
			fmt.Sprintf(
				`INSERT INTO endpoint_configs (endpoint_key, is_external, payload_json, updated_at)
				 VALUES (%s, %s, %s, %s) %s`,
				s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3), s.d.placeholder(4),
				upsertClause,
			),
			k, isExternal, string(payload), time.Now().UTC())
		return err
	})
}

// Save is a no-op; the database persists writes immediately.
func (s *Store) Save() error { return nil }

// Close closes the underlying database handle.
func (s *Store) Close() {
	_ = s.db.Close()
	if s.cacheEnabled {
		s.cacheMu.Lock()
		s.cache = make(map[string]*endpoint.Status)
		s.cacheMu.Unlock()
	}
}

// ---------------------------------------------------------------------------
// internal helpers (only exported wrappers may begin/commit/rollback transactions)
// ---------------------------------------------------------------------------

// lookupEndpointID fetches the endpoint_id for a key without a subquery.
// Used to work around MySQL's restriction on LIMIT in subqueries referencing
// the same table being modified.
func (s *Store) lookupEndpointID(k string) (int64, error) {
	var id int64
	err := s.db.QueryRow(
		fmt.Sprintf("SELECT endpoint_id FROM endpoints WHERE endpoint_key = %s LIMIT 1", s.d.placeholder(1)),
		k,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) {
			return 0, common.ErrEndpointNotFound
		}
		return 0, err
	}
	return id, nil
}

func (s *Store) insertEndpoint(tx *databasesql.Tx, ep *endpoint.Endpoint) (int64, error) {
	returning := s.d.returningID("endpoint_id")
	query := fmt.Sprintf(
		"INSERT INTO endpoints (endpoint_key, endpoint_name, endpoint_group) VALUES (%s, %s, %s)%s",
		s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3),
		returning,
	)
	if s.driver == "mysql" {
		result, err := tx.Exec(query, ep.Key(), ep.Name, ep.Group)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}
	var id int64
	err := tx.QueryRow(query, ep.Key(), ep.Name, ep.Group).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) insertEndpointEvent(tx *databasesql.Tx, endpointID int64, event *endpoint.Event) error {
	_, err := tx.Exec(
		fmt.Sprintf("INSERT INTO endpoint_events (endpoint_id, event_type, event_timestamp) VALUES (%s, %s, %s)",
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID, event.Type, event.Timestamp.UTC(),
	)
	return err
}

func (s *Store) insertEndpointResult(tx *databasesql.Tx, endpointID int64, result *endpoint.Result) error {
	returning := s.d.returningID("endpoint_result_id")
	query := fmt.Sprintf(
		`INSERT INTO endpoint_results (endpoint_id, success, errors, connected, status, dns_rcode, certificate_expiration, domain_expiration, hostname, ip, duration, timestamp)
		 VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)%s`,
		s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3), s.d.placeholder(4),
		s.d.placeholder(5), s.d.placeholder(6), s.d.placeholder(7), s.d.placeholder(8),
		s.d.placeholder(9), s.d.placeholder(10), s.d.placeholder(11), s.d.placeholder(12),
		returning,
	)
	var endpointResultID int64
	if s.driver == "mysql" {
		execResult, err := tx.Exec(query,
			endpointID,
			result.Success,
			strings.Join(result.Errors, arraySeparator),
			result.Connected,
			result.HTTPStatus,
			result.DNSRCode,
			result.CertificateExpiration,
			result.DomainExpiration,
			result.Hostname,
			result.IP,
			result.Duration,
			result.Timestamp.UTC(),
		)
		if err != nil {
			return err
		}
		endpointResultID, err = execResult.LastInsertId()
		if err != nil {
			return err
		}
	} else {
		err := tx.QueryRow(query,
			endpointID,
			result.Success,
			strings.Join(result.Errors, arraySeparator),
			result.Connected,
			result.HTTPStatus,
			result.DNSRCode,
			result.CertificateExpiration,
			result.DomainExpiration,
			result.Hostname,
			result.IP,
			result.Duration,
			result.Timestamp.UTC(),
		).Scan(&endpointResultID)
		if err != nil {
			return err
		}
	}
	return s.insertConditionResults(tx, endpointResultID, result.ConditionResults)
}

func (s *Store) insertConditionResults(tx *databasesql.Tx, endpointResultID int64, conditionResults []*endpoint.ConditionResult) error {
	for _, cr := range conditionResults {
		if _, err := tx.Exec(
			fmt.Sprintf("INSERT INTO endpoint_result_conditions (endpoint_result_id, condition, success) VALUES (%s, %s, %s)",
				s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
			endpointResultID, cr.Condition, cr.Success,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) updateEndpointUptime(tx *databasesql.Tx, endpointID int64, result *endpoint.Result) error {
	unixTimestampFlooredAtHour := result.Timestamp.Truncate(time.Hour).Unix()
	successfulExecutions := 0
	if result.Success {
		successfulExecutions = 1
	}
	// This upsert ADDS to existing totals, so we need dialect-specific expressions.
	var conflictClause string
	if s.driver == "mysql" {
		conflictClause = `ON DUPLICATE KEY UPDATE
				total_executions = total_executions + VALUES(total_executions),
				successful_executions = successful_executions + VALUES(successful_executions),
				total_response_time = total_response_time + VALUES(total_response_time)`
	} else {
		conflictClause = `ON CONFLICT(endpoint_id, hour_unix_timestamp) DO UPDATE SET
				total_executions = excluded.total_executions + endpoint_uptimes.total_executions,
				successful_executions = excluded.successful_executions + endpoint_uptimes.successful_executions,
				total_response_time = excluded.total_response_time + endpoint_uptimes.total_response_time`
	}
	_, err := tx.Exec(
		fmt.Sprintf(
			`INSERT INTO endpoint_uptimes (endpoint_id, hour_unix_timestamp, total_executions, successful_executions, total_response_time)
			 VALUES (%s, %s, %s, %s, %s) %s`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3), s.d.placeholder(4), s.d.placeholder(5),
			conflictClause,
		),
		endpointID,
		unixTimestampFlooredAtHour,
		1,
		successfulExecutions,
		result.Duration.Milliseconds(),
	)
	return err
}

func (s *Store) getAllEndpointKeys(tx *databasesql.Tx) (keys []string, err error) {
	rows, err := tx.Query(`SELECT endpoint_key FROM endpoints ORDER BY endpoint_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		_ = rows.Scan(&k)
		keys = append(keys, k)
	}
	return
}

func (s *Store) getEndpointStatusByKey(tx *databasesql.Tx, k string, parameters *paging.EndpointStatusParams) (*endpoint.Status, error) {
	if s.cacheEnabled {
		cacheKey := generateCacheKey(k, parameters)
		s.cacheMu.RLock()
		cached, ok := s.cache[cacheKey]
		s.cacheMu.RUnlock()
		if ok {
			return cached, nil
		}
	}
	endpointID, group, name, err := s.getEndpointIDGroupAndNameByKey(tx, k)
	if err != nil {
		return nil, err
	}
	status := endpoint.NewStatus(group, name)
	if parameters.EventsPageSize > 0 {
		if status.Events, err = s.getEndpointEventsByEndpointID(tx, endpointID, parameters.EventsPage, parameters.EventsPageSize); err != nil {
			s.log.Error("getEndpointStatusByKey: failed to retrieve events", plumelog.Fields{"op": "sql.getEndpointStatusByKey", "key": k, "err": err.Error()})
		}
	}
	if parameters.ResultsPageSize > 0 {
		if status.Results, err = s.getEndpointResultsByEndpointID(tx, endpointID, parameters.ResultsPage, parameters.ResultsPageSize); err != nil {
			s.log.Error("getEndpointStatusByKey: failed to retrieve results", plumelog.Fields{"op": "sql.getEndpointStatusByKey", "key": k, "err": err.Error()})
		}
	}
	if s.cacheEnabled {
		cacheKey := generateCacheKey(k, parameters)
		s.cacheMu.Lock()
		s.cache[cacheKey] = status
		s.cacheMu.Unlock()
	}
	return status, nil
}

func (s *Store) getEndpointIDGroupAndNameByKey(tx *databasesql.Tx, k string) (id int64, group, name string, err error) {
	err = tx.QueryRow(
		fmt.Sprintf("SELECT endpoint_id, endpoint_group, endpoint_name FROM endpoints WHERE endpoint_key = %s LIMIT 1",
			s.d.placeholder(1)),
		k,
	).Scan(&id, &group, &name)
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) {
			return 0, "", "", common.ErrEndpointNotFound
		}
		return 0, "", "", err
	}
	return
}

func (s *Store) getEndpointEventsByEndpointID(tx *databasesql.Tx, endpointID int64, page, pageSize int) (events []*endpoint.Event, err error) {
	rows, err := tx.Query(
		fmt.Sprintf(
			`SELECT event_type, event_timestamp
			 FROM (
				SELECT event_type, event_timestamp, endpoint_event_id
				FROM endpoint_events
				WHERE endpoint_id = %s
				ORDER BY endpoint_event_id DESC
				LIMIT %s OFFSET %s
			 ) AS recent_events
			 ORDER BY endpoint_event_id ASC`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID,
		pageSize,
		(page-1)*pageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		event := &endpoint.Event{}
		_ = rows.Scan(&event.Type, &event.Timestamp)
		events = append(events, event)
	}
	return
}

func (s *Store) getEndpointResultsByEndpointID(tx *databasesql.Tx, endpointID int64, page, pageSize int) (results []*endpoint.Result, err error) {
	rows, err := tx.Query(
		fmt.Sprintf(
			`SELECT endpoint_result_id, success, errors, connected, status, dns_rcode, certificate_expiration, domain_expiration, hostname, ip, duration, timestamp
			 FROM endpoint_results
			 WHERE endpoint_id = %s
			 ORDER BY endpoint_result_id DESC
			 LIMIT %s OFFSET %s`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID,
		pageSize,
		(page-1)*pageSize,
	)
	if err != nil {
		return nil, err
	}
	idResultMap := make(map[int64]*endpoint.Result)
	for rows.Next() {
		result := &endpoint.Result{}
		var id int64
		var joinedErrors string
		err = rows.Scan(&id, &result.Success, &joinedErrors, &result.Connected, &result.HTTPStatus, &result.DNSRCode, &result.CertificateExpiration, &result.DomainExpiration, &result.Hostname, &result.IP, &result.Duration, &result.Timestamp)
		if err != nil {
			s.log.Error("getEndpointResultsByEndpointID: silently failed to retrieve endpoint result", plumelog.Fields{"op": "sql.getEndpointResultsByEndpointID", "endpointID": endpointID, "err": err.Error()})
			err = nil
		}
		if joinedErrors != "" {
			result.Errors = strings.Split(joinedErrors, arraySeparator)
		}
		results = append([]*endpoint.Result{result}, results...)
		idResultMap[id] = result
	}
	rows.Close()
	if len(idResultMap) == 0 {
		return
	}
	args := make([]interface{}, 0, len(idResultMap))
	query := `SELECT endpoint_result_id, condition, success
				FROM endpoint_result_conditions
				WHERE endpoint_result_id IN (`
	index := 1
	for endpointResultID := range idResultMap {
		if index > 1 {
			query += ","
		}
		query += s.d.placeholder(index)
		args = append(args, endpointResultID)
		index++
	}
	query += ")"
	rows2, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		cr := &endpoint.ConditionResult{}
		var endpointResultID int64
		if err = rows2.Scan(&endpointResultID, &cr.Condition, &cr.Success); err != nil {
			return
		}
		idResultMap[endpointResultID].ConditionResults = append(idResultMap[endpointResultID].ConditionResults, cr)
	}
	return
}

func (s *Store) getEndpointUptime(tx *databasesql.Tx, endpointID int64, from, to time.Time) (uptime float64, avgResponseTime time.Duration, err error) {
	rows, err := tx.Query(
		fmt.Sprintf(
			`SELECT SUM(total_executions), SUM(successful_executions), SUM(total_response_time)
			 FROM endpoint_uptimes
			 WHERE endpoint_id = %s
				AND hour_unix_timestamp >= %s
				AND hour_unix_timestamp <= %s`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID, from.Unix(), to.Unix(),
	)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	var totalExecutions, totalSuccessfulExecutions, totalResponseTime int
	for rows.Next() {
		_ = rows.Scan(&totalExecutions, &totalSuccessfulExecutions, &totalResponseTime)
	}
	if totalExecutions > 0 {
		uptime = float64(totalSuccessfulExecutions) / float64(totalExecutions)
		avgResponseTime = time.Duration(float64(totalResponseTime)/float64(totalExecutions)) * time.Millisecond
	}
	return
}

func (s *Store) getEndpointAverageResponseTime(tx *databasesql.Tx, endpointID int64, from, to time.Time) (int, error) {
	rows, err := tx.Query(
		fmt.Sprintf(
			`SELECT SUM(total_executions), SUM(total_response_time)
			 FROM endpoint_uptimes
			 WHERE endpoint_id = %s
				AND total_executions > 0
				AND hour_unix_timestamp >= %s
				AND hour_unix_timestamp <= %s`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID, from.Unix(), to.Unix(),
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var totalExecutions, totalResponseTime int
	for rows.Next() {
		_ = rows.Scan(&totalExecutions, &totalResponseTime)
	}
	if totalExecutions == 0 {
		return 0, nil
	}
	return int(float64(totalResponseTime) / float64(totalExecutions)), nil
}

func (s *Store) getEndpointHourlyAverageResponseTimes(tx *databasesql.Tx, endpointID int64, from, to time.Time) (map[int64]int, error) {
	rows, err := tx.Query(
		fmt.Sprintf(
			`SELECT hour_unix_timestamp, total_executions, total_response_time
			 FROM endpoint_uptimes
			 WHERE endpoint_id = %s
				AND total_executions > 0
				AND hour_unix_timestamp >= %s
				AND hour_unix_timestamp <= %s`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID, from.Unix(), to.Unix(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64]int)
	var totalExecutions, totalResponseTime int
	var ts int64
	for rows.Next() {
		_ = rows.Scan(&ts, &totalExecutions, &totalResponseTime)
		out[ts] = int(float64(totalResponseTime) / float64(totalExecutions))
	}
	return out, nil
}

func (s *Store) getEndpointID(tx *databasesql.Tx, ep *endpoint.Endpoint) (int64, error) {
	var id int64
	err := tx.QueryRow(
		fmt.Sprintf("SELECT endpoint_id FROM endpoints WHERE endpoint_key = %s", s.d.placeholder(1)),
		ep.Key(),
	).Scan(&id)
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) {
			return 0, common.ErrEndpointNotFound
		}
		return 0, err
	}
	return id, nil
}

func (s *Store) getNumberOfEventsByEndpointID(tx *databasesql.Tx, endpointID int64) (int64, error) {
	var n int64
	err := tx.QueryRow(
		fmt.Sprintf("SELECT COUNT(1) FROM endpoint_events WHERE endpoint_id = %s", s.d.placeholder(1)),
		endpointID,
	).Scan(&n)
	return n, err
}

func (s *Store) getNumberOfResultsByEndpointID(tx *databasesql.Tx, endpointID int64) (int64, error) {
	var n int64
	err := tx.QueryRow(
		fmt.Sprintf("SELECT COUNT(1) FROM endpoint_results WHERE endpoint_id = %s", s.d.placeholder(1)),
		endpointID,
	).Scan(&n)
	return n, err
}

func (s *Store) getNumberOfUptimeEntriesByEndpointID(tx *databasesql.Tx, endpointID int64) (int64, error) {
	var n int64
	err := tx.QueryRow(
		fmt.Sprintf("SELECT COUNT(1) FROM endpoint_uptimes WHERE endpoint_id = %s", s.d.placeholder(1)),
		endpointID,
	).Scan(&n)
	return n, err
}

func (s *Store) getAgeOfOldestEndpointUptimeEntry(tx *databasesql.Tx, endpointID int64) (time.Duration, error) {
	rows, err := tx.Query(
		fmt.Sprintf("SELECT hour_unix_timestamp FROM endpoint_uptimes WHERE endpoint_id = %s ORDER BY hour_unix_timestamp LIMIT 1",
			s.d.placeholder(1)),
		endpointID,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var oldest int64
	var found bool
	for rows.Next() {
		_ = rows.Scan(&oldest)
		found = true
	}
	if !found {
		return 0, errNoRowsReturned
	}
	return time.Since(time.Unix(oldest, 0)), nil
}

func (s *Store) getLastEndpointResultSuccessValue(tx *databasesql.Tx, endpointID int64) (bool, error) {
	var success bool
	err := tx.QueryRow(
		fmt.Sprintf("SELECT success FROM endpoint_results WHERE endpoint_id = %s ORDER BY endpoint_result_id DESC LIMIT 1",
			s.d.placeholder(1)),
		endpointID,
	).Scan(&success)
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) {
			return false, errNoRowsReturned
		}
		return false, err
	}
	return success, nil
}

func (s *Store) deleteOldEndpointEvents(tx *databasesql.Tx, endpointID int64) error {
	_, err := tx.Exec(
		fmt.Sprintf(
			`DELETE FROM endpoint_events
			 WHERE endpoint_id = %s
				AND endpoint_event_id NOT IN (
					SELECT endpoint_event_id
					FROM endpoint_events
					WHERE endpoint_id = %s
					ORDER BY endpoint_event_id DESC
					LIMIT %s
				)`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID, endpointID, s.maximumNumberOfEvents,
	)
	return err
}

func (s *Store) deleteOldEndpointResults(tx *databasesql.Tx, endpointID int64) error {
	_, err := tx.Exec(
		fmt.Sprintf(
			`DELETE FROM endpoint_results
			 WHERE endpoint_id = %s
				AND endpoint_result_id NOT IN (
					SELECT endpoint_result_id
					FROM endpoint_results
					WHERE endpoint_id = %s
					ORDER BY endpoint_result_id DESC
					LIMIT %s
				)`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID, endpointID, s.maximumNumberOfResults,
	)
	return err
}

func (s *Store) deleteOldUptimeEntries(tx *databasesql.Tx, endpointID int64, maxAge time.Time) error {
	_, err := tx.Exec(
		fmt.Sprintf("DELETE FROM endpoint_uptimes WHERE endpoint_id = %s AND hour_unix_timestamp < %s",
			s.d.placeholder(1), s.d.placeholder(2)),
		endpointID, maxAge.Unix(),
	)
	return err
}

func (s *Store) mergeHourlyUptimeEntriesOlderThanMergeThresholdIntoDailyUptimeEntries(tx *databasesql.Tx, endpointID int64) error {
	now := time.Now()
	minThreshold := now.Add(-uptimeHourlyBuffer)
	minThreshold = time.Date(minThreshold.Year(), minThreshold.Month(), minThreshold.Day(), 0, 0, 0, 0, minThreshold.Location())
	maxThreshold := now.Add(-uptimeRetention)
	rows, err := tx.Query(
		fmt.Sprintf(
			`SELECT hour_unix_timestamp, total_executions, successful_executions, total_response_time
			 FROM endpoint_uptimes
			 WHERE endpoint_id = %s
				AND hour_unix_timestamp < %s
				AND hour_unix_timestamp >= %s`,
			s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3)),
		endpointID, minThreshold.Unix(), maxThreshold.Unix(),
	)
	if err != nil {
		return err
	}
	type entry struct {
		totalExecutions      int
		successfulExecutions int
		totalResponseTime    int
	}
	dailyEntries := make(map[int64]*entry)
	for rows.Next() {
		var ts int64
		e := entry{}
		if err = rows.Scan(&ts, &e.totalExecutions, &e.successfulExecutions, &e.totalResponseTime); err != nil {
			rows.Close()
			return err
		}
		t := time.Unix(ts, 0)
		dayTS := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
		if d := dailyEntries[dayTS]; d == nil {
			dailyEntries[dayTS] = &e
		} else {
			d.totalExecutions += e.totalExecutions
			d.successfulExecutions += e.successfulExecutions
			d.totalResponseTime += e.totalResponseTime
		}
	}
	rows.Close()
	if _, err = tx.Exec(
		fmt.Sprintf("DELETE FROM endpoint_uptimes WHERE endpoint_id = %s AND hour_unix_timestamp < %s",
			s.d.placeholder(1), s.d.placeholder(2)),
		endpointID, minThreshold.Unix(),
	); err != nil {
		return err
	}
	upsertClause := s.d.upsertClause("endpoint_id, hour_unix_timestamp", []string{
		"total_executions",
		"successful_executions",
		"total_response_time",
	})
	for ts, e := range dailyEntries {
		_, err = tx.Exec(
			fmt.Sprintf(
				`INSERT INTO endpoint_uptimes (endpoint_id, hour_unix_timestamp, total_executions, successful_executions, total_response_time)
				 VALUES (%s, %s, %s, %s, %s) %s`,
				s.d.placeholder(1), s.d.placeholder(2), s.d.placeholder(3), s.d.placeholder(4), s.d.placeholder(5),
				upsertClause,
			),
			endpointID, ts, e.totalExecutions, e.successfulExecutions, e.totalResponseTime,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateCacheKey(endpointKey string, p *paging.EndpointStatusParams) string {
	return fmt.Sprintf("%s-%d-%d-%d-%d", endpointKey, p.EventsPage, p.EventsPageSize, p.ResultsPage, p.ResultsPageSize)
}

func (s *Store) invalidateCacheByPrefix(prefix string) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	for k := range s.cache {
		if strings.HasPrefix(k, prefix) {
			delete(s.cache, k)
		}
	}
}

