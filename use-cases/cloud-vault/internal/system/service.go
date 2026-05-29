package system

import (
	"context"
	"database/sql"

	"cloud-vault/internal/config"
	"cloud-vault/internal/storage"
)

// Service provides system observability operations.
type Service struct {
	db    *sql.DB
	store storage.ObjectStorage
	aiCfg config.AIConfig
}

func NewService(db *sql.DB, store storage.ObjectStorage, aiCfg config.AIConfig) *Service {
	return &Service{db: db, store: store, aiCfg: aiCfg}
}

func (s *Service) Health(ctx context.Context) HealthResult {
	return checkHealth(ctx, s.db, s.store, s.aiCfg)
}

func (s *Service) Stats(ctx context.Context) (StatsResult, error) {
	return collectStats(ctx, s.db)
}

func (s *Service) Doctor(ctx context.Context, req DoctorRequest) DoctorResult {
	checks := req.Checks
	if len(checks) == 0 {
		checks = AllChecks
	}

	sampleSize := req.SampleSize
	if sampleSize <= 0 {
		sampleSize = defaultSampleSize
	}

	var results []CheckResult
	for _, name := range checks {
		results = append(results, runCheck(ctx, name, s.db, s.store, sampleSize))
	}

	return DoctorResult{
		Status: worstStatus(results),
		Checks: results,
	}
}
