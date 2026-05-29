package system

import (
	"context"
	"database/sql"

	"cloud-vault/internal/config"
	"cloud-vault/internal/storage"
)

func checkHealth(ctx context.Context, db *sql.DB, store storage.ObjectStorage, aiCfg config.AIConfig) HealthResult {
	r := HealthResult{
		Database: checkDatabase(ctx, db),
		Storage:  checkStorage(ctx, store),
		Search:   checkSearch(ctx, db),
		AI:       checkAI(aiCfg),
	}
	r.Status = overallStatus(r)
	return r
}

func checkDatabase(ctx context.Context, db *sql.DB) string {
	if err := db.PingContext(ctx); err != nil {
		return StatusError
	}
	return StatusOK
}

func checkStorage(ctx context.Context, store storage.ObjectStorage) string {
	// Probe with a key that should not exist; error on Exists itself = problem.
	_, err := store.Exists(ctx, ".health-probe")
	if err != nil {
		return StatusWarning
	}
	return StatusOK
}

func checkSearch(ctx context.Context, db *sql.DB) string {
	var n int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM document_fts").Scan(&n)
	if err != nil {
		return StatusError
	}
	return StatusOK
}

func checkAI(cfg config.AIConfig) string {
	if !cfg.Enabled {
		return StatusDisabled
	}
	return StatusOK
}

// overallStatus returns the worst status across database, storage, and search.
// AI disabled does not count as a problem.
func overallStatus(r HealthResult) string {
	components := []string{r.Database, r.Storage, r.Search}
	worst := StatusOK
	for _, s := range components {
		switch {
		case s == StatusError:
			return StatusError
		case s == StatusWarning && worst == StatusOK:
			worst = StatusWarning
		}
	}
	return worst
}
