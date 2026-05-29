// Package memory implements an in-process Store with no durability.
//
// Unlike upstream gatus this does not depend on gocache; the data set is
// small (one *endpoint.Status per configured endpoint) and a typed map
// guarded by sync.RWMutex is sufficient.
package memory

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
	"guardus/internal/domain/key"
	"guardus/internal/storage/common"
	"guardus/internal/storage/common/paging"
)

// Store keeps endpoint statuses in memory. Triggered-alert persistence is a
// no-op because the data does not survive process restart.
type Store struct {
	mu sync.RWMutex

	statuses          map[string]*endpoint.Status
	endpointConfigs   map[string]*endpoint.Endpoint
	externalConfigs   map[string]*endpoint.ExternalEndpoint
	endpointConfigsMu sync.RWMutex

	maximumNumberOfResults int
	maximumNumberOfEvents  int
}

// NewStore returns an in-memory Store with the given retention limits.
func NewStore(maximumNumberOfResults, maximumNumberOfEvents int) *Store {
	return &Store{
		statuses:               make(map[string]*endpoint.Status),
		endpointConfigs:        make(map[string]*endpoint.Endpoint),
		externalConfigs:        make(map[string]*endpoint.ExternalEndpoint),
		maximumNumberOfResults: maximumNumberOfResults,
		maximumNumberOfEvents:  maximumNumberOfEvents,
	}
}

// GetAllEndpointStatuses returns a paginated copy of every known endpoint
// status, sorted by key for deterministic output.
func (s *Store) GetAllEndpointStatuses(params *paging.EndpointStatusParams) ([]*endpoint.Status, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*endpoint.Status, 0, len(s.statuses))
	for _, status := range s.statuses {
		out = append(out, ShallowCopyEndpointStatus(status, params))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// GetEndpointStatus returns the status for (groupName, endpointName).
func (s *Store) GetEndpointStatus(groupName, endpointName string, params *paging.EndpointStatusParams) (*endpoint.Status, error) {
	return s.GetEndpointStatusByKey(key.ConvertGroupAndNameToKey(groupName, endpointName), params)
}

// GetEndpointStatusByKey returns the status for the given endpoint key.
func (s *Store) GetEndpointStatusByKey(k string, params *paging.EndpointStatusParams) (*endpoint.Status, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status, ok := s.statuses[k]
	if !ok {
		return nil, common.ErrEndpointNotFound
	}
	return ShallowCopyEndpointStatus(status, params), nil
}

// GetUptimeByKey returns the uptime ratio over the [from, to] range.
func (s *Store) GetUptimeByKey(k string, from, to time.Time) (float64, error) {
	if from.After(to) {
		return 0, common.ErrInvalidTimeRange
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	status, ok := s.statuses[k]
	if !ok || status.Uptime == nil {
		return 0, common.ErrEndpointNotFound
	}
	var successful, total uint64
	current := from
	for !current.After(to) {
		ts := current.Truncate(time.Hour).Unix()
		if hs := status.Uptime.HourlyStatistics[ts]; hs != nil && hs.TotalExecutions > 0 {
			successful += hs.SuccessfulExecutions
			total += hs.TotalExecutions
		}
		current = current.Add(time.Hour)
	}
	if total == 0 {
		return 0, nil
	}
	return float64(successful) / float64(total), nil
}

// GetAverageResponseTimeByKey returns the average response time in ms over the range.
func (s *Store) GetAverageResponseTimeByKey(k string, from, to time.Time) (int, error) {
	if from.After(to) {
		return 0, common.ErrInvalidTimeRange
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	status, ok := s.statuses[k]
	if !ok || status.Uptime == nil {
		return 0, common.ErrEndpointNotFound
	}
	var totalExec, totalRT uint64
	current := from
	for !current.After(to) {
		ts := current.Truncate(time.Hour).Unix()
		if hs := status.Uptime.HourlyStatistics[ts]; hs != nil && hs.TotalExecutions > 0 {
			totalExec += hs.TotalExecutions
			totalRT += hs.TotalExecutionsResponseTime
		}
		current = current.Add(time.Hour)
	}
	if totalExec == 0 {
		return 0, nil
	}
	return int(float64(totalRT) / float64(totalExec)), nil
}

// GetHourlyAverageResponseTimeByKey returns hourly average response times in ms.
func (s *Store) GetHourlyAverageResponseTimeByKey(k string, from, to time.Time) (map[int64]int, error) {
	if from.After(to) {
		return nil, common.ErrInvalidTimeRange
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	status, ok := s.statuses[k]
	if !ok || status.Uptime == nil {
		return nil, common.ErrEndpointNotFound
	}
	out := make(map[int64]int)
	current := from
	for !current.After(to) {
		ts := current.Truncate(time.Hour).Unix()
		if hs := status.Uptime.HourlyStatistics[ts]; hs != nil && hs.TotalExecutions > 0 {
			out[ts] = int(float64(hs.TotalExecutionsResponseTime) / float64(hs.TotalExecutions))
		}
		current = current.Add(time.Hour)
	}
	return out, nil
}

// InsertEndpointResult records a probe outcome.
func (s *Store) InsertEndpointResult(ep *endpoint.Endpoint, result *endpoint.Result) error {
	endpointKey := ep.Key()
	s.mu.Lock()
	defer s.mu.Unlock()
	status, exists := s.statuses[endpointKey]
	if !exists {
		status = endpoint.NewStatus(ep.Group, ep.Name)
		status.Events = append(status.Events, &endpoint.Event{
			Type:      endpoint.EventStart,
			Timestamp: time.Now(),
		})
	}
	AddResult(status, result, s.maximumNumberOfResults, s.maximumNumberOfEvents)
	s.statuses[endpointKey] = status
	return nil
}

// DeleteAllEndpointStatusesNotInKeys drops statuses whose key is not in keep.
func (s *Store) DeleteAllEndpointStatusesNotInKeys(keep []string) int {
	keepSet := make(map[string]struct{}, len(keep))
	for _, k := range keep {
		keepSet[k] = struct{}{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	deleted := 0
	for k := range s.statuses {
		if _, ok := keepSet[k]; !ok {
			delete(s.statuses, k)
			deleted++
		}
	}
	return deleted
}

// GetTriggeredEndpointAlert always reports no triggered alert; the in-memory
// store does not persist alert state across restarts.
func (s *Store) GetTriggeredEndpointAlert(_ *endpoint.Endpoint, _ *alert.Alert) (bool, string, int, error) {
	return false, "", 0, nil
}

// UpsertTriggeredEndpointAlert is a no-op for the in-memory store.
func (s *Store) UpsertTriggeredEndpointAlert(_ *endpoint.Endpoint, _ *alert.Alert) error {
	return nil
}

// DeleteTriggeredEndpointAlert is a no-op for the in-memory store.
func (s *Store) DeleteTriggeredEndpointAlert(_ *endpoint.Endpoint, _ *alert.Alert) error {
	return nil
}

// DeleteAllTriggeredAlertsNotInChecksumsByEndpoint is a no-op for the in-memory store.
func (s *Store) DeleteAllTriggeredAlertsNotInChecksumsByEndpoint(_ *endpoint.Endpoint, _ []string) int {
	return 0
}

// HasEndpointStatusNewerThan reports whether the status has any result newer than ts.
func (s *Store) HasEndpointStatusNewerThan(k string, ts time.Time) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status, ok := s.statuses[k]
	if !ok {
		return false, nil
	}
	for _, result := range status.Results {
		if result.Timestamp.After(ts) {
			return true, nil
		}
	}
	return false, nil
}

// ListEndpointConfigs returns the in-memory endpoint definitions, sorted by key.
func (s *Store) ListEndpointConfigs(_ context.Context) ([]*endpoint.Endpoint, []*endpoint.ExternalEndpoint, error) {
	s.endpointConfigsMu.RLock()
	defer s.endpointConfigsMu.RUnlock()
	endpoints := make([]*endpoint.Endpoint, 0, len(s.endpointConfigs))
	for _, ep := range s.endpointConfigs {
		endpoints = append(endpoints, ep)
	}
	sort.Slice(endpoints, func(i, j int) bool { return endpoints[i].Key() < endpoints[j].Key() })
	external := make([]*endpoint.ExternalEndpoint, 0, len(s.externalConfigs))
	for _, ee := range s.externalConfigs {
		external = append(external, ee)
	}
	sort.Slice(external, func(i, j int) bool { return external[i].Key() < external[j].Key() })
	return endpoints, external, nil
}

// UpsertEndpointConfig stores ep keyed by ep.Key().
func (s *Store) UpsertEndpointConfig(_ context.Context, ep *endpoint.Endpoint) error {
	if ep == nil {
		return errors.New("endpoint is nil")
	}
	s.endpointConfigsMu.Lock()
	defer s.endpointConfigsMu.Unlock()
	s.endpointConfigs[ep.Key()] = ep
	delete(s.externalConfigs, ep.Key())
	return nil
}

// UpsertExternalEndpointConfig stores ee keyed by ee.Key().
func (s *Store) UpsertExternalEndpointConfig(_ context.Context, ee *endpoint.ExternalEndpoint) error {
	if ee == nil {
		return errors.New("external endpoint is nil")
	}
	s.endpointConfigsMu.Lock()
	defer s.endpointConfigsMu.Unlock()
	s.externalConfigs[ee.Key()] = ee
	delete(s.endpointConfigs, ee.Key())
	return nil
}

// Clear deletes everything.
func (s *Store) Clear() {
	s.mu.Lock()
	s.statuses = make(map[string]*endpoint.Status)
	s.mu.Unlock()
	s.endpointConfigsMu.Lock()
	s.endpointConfigs = make(map[string]*endpoint.Endpoint)
	s.externalConfigs = make(map[string]*endpoint.ExternalEndpoint)
	s.endpointConfigsMu.Unlock()
}

// Save is a no-op; the in-memory store has no durable backing.
func (s *Store) Save() error { return nil }

// Close is a no-op.
func (s *Store) Close() {}
