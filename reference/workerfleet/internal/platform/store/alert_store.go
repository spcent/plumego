package store

import (
	"sort"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

type AlertFilter struct {
	WorkerID  domain.WorkerID
	AlertType domain.AlertType
	Status    domain.AlertStatus
}

type AlertRecord = domain.AlertRecord

type AlertStore interface {
	AppendAlert(record AlertRecord) error
	ListAlerts(filter AlertFilter) ([]AlertRecord, error)
	ListAlertRecords() ([]AlertRecord, error)
}

func (s *MemoryStore) AppendAlert(record AlertRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cloned := cloneAlertRecord(record)
	s.alerts = append(s.alerts, cloned)
	sort.Slice(s.alerts, func(i, j int) bool {
		return s.alerts[i].TriggeredAt.Before(s.alerts[j].TriggeredAt)
	})
	return nil
}

func (s *MemoryStore) ListAlerts(filter AlertFilter) ([]AlertRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]AlertRecord, 0, len(s.alerts))
	for _, alert := range s.alerts {
		if filter.WorkerID != "" && alert.WorkerID != filter.WorkerID {
			continue
		}
		if filter.AlertType != "" && alert.AlertType != filter.AlertType {
			continue
		}
		if filter.Status != "" && alert.Status != filter.Status {
			continue
		}
		out = append(out, cloneAlertRecord(alert))
	}
	return out, nil
}

func (s *MemoryStore) ListAlertRecords() ([]AlertRecord, error) {
	return s.ListAlerts(AlertFilter{})
}

func cloneAlertRecord(record AlertRecord) AlertRecord {
	record.Details = cloneStringMap(record.Details)
	return record
}

var _ AlertStore = (*MemoryStore)(nil)
