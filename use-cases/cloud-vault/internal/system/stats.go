package system

import (
	"context"
	"database/sql"
)

func collectStats(ctx context.Context, db *sql.DB) (StatsResult, error) {
	var s StatsResult
	type countQuery struct {
		dest *int64
		sql  string
	}
	queries := []countQuery{
		{&s.Documents, "SELECT COUNT(*) FROM documents WHERE status != 'deleted'"},
		{&s.Versions, "SELECT COUNT(*) FROM document_versions"},
		{&s.Collections, "SELECT COUNT(*) FROM collections WHERE status != 'deleted'"},
		{&s.Tags, "SELECT COUNT(*) FROM tags"},
		{&s.ImportJobs, "SELECT COUNT(*) FROM import_jobs"},
		{&s.IndexedDocuments, "SELECT COUNT(*) FROM document_index_status WHERE status = 'indexed'"},
		{&s.PendingIndexes, "SELECT COUNT(*) FROM document_index_status WHERE status IN ('pending', 'stale')"},
		{&s.FailedIndexes, "SELECT COUNT(*) FROM document_index_status WHERE status = 'failed'"},
		{&s.AITasks, "SELECT COUNT(*) FROM ai_tasks"},
		{&s.Prompts, "SELECT COUNT(*) FROM prompts"},
	}
	for _, q := range queries {
		if err := db.QueryRowContext(ctx, q.sql).Scan(q.dest); err != nil {
			return s, err
		}
	}
	return s, nil
}
