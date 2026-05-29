package system

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"

	"cloud-vault/internal/storage"
)

const defaultMaxItems = 100
const defaultSampleSize = 100

// runCheck dispatches to the correct check function by name.
func runCheck(ctx context.Context, name string, db *sql.DB, store storage.ObjectStorage, sampleSize int) CheckResult {
	switch name {
	case "storage_objects":
		return checkStorageObjects(ctx, db, store)
	case "document_versions":
		return checkDocumentVersions(ctx, db)
	case "document_hash":
		return checkDocumentHash(ctx, db, store, sampleSize)
	case "fts_index":
		return checkFTSIndex(ctx, db)
	case "tags":
		return checkTags(ctx, db)
	case "collections":
		return checkCollections(ctx, db)
	case "sources":
		return checkSources(ctx, db)
	case "imports":
		return checkImports(ctx, db)
	case "ai":
		return checkAIConsistency(ctx, db)
	default:
		return CheckResult{Name: name, Status: StatusError, Items: []IssueItem{{Issue: "unknown_check"}}}
	}
}

// checkStorageObjects verifies every active document's storage_key exists in object storage.
func checkStorageObjects(ctx context.Context, db *sql.DB, store storage.ObjectStorage) CheckResult {
	r := CheckResult{Name: "storage_objects", Status: StatusOK}

	rows, err := db.QueryContext(ctx,
		`SELECT id, storage_key FROM documents WHERE status != 'deleted' LIMIT 1000`)
	if err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "query_failed", Detail: err.Error()}}
		return r
	}
	defer rows.Close()

	for rows.Next() {
		var docID, key string
		if err := rows.Scan(&docID, &key); err != nil {
			continue
		}
		r.Total++
		ok, err := store.Exists(ctx, key)
		if err != nil || !ok {
			r.Failed++
			issue := "storage_object_missing"
			if err != nil {
				issue = "storage_read_failed"
			}
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{
					DocumentID: docID,
					Issue:      issue,
					Detail:     key,
				})
			}
		}
	}
	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkDocumentVersions verifies version integrity.
func checkDocumentVersions(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "document_versions", Status: StatusOK}

	// Check current_version matches max version in document_versions.
	rows, err := db.QueryContext(ctx, `
		SELECT d.id, d.current_version, COALESCE(MAX(dv.version), 0) as max_v
		FROM documents d
		LEFT JOIN document_versions dv ON dv.document_id = d.id
		WHERE d.status != 'deleted'
		GROUP BY d.id
		HAVING d.current_version != max_v
		LIMIT 1000`)
	if err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "query_failed", Detail: err.Error()}}
		return r
	}
	defer rows.Close()

	for rows.Next() {
		var docID string
		var cur, maxV int
		if err := rows.Scan(&docID, &cur, &maxV); err != nil {
			continue
		}
		r.Total++
		r.Failed++
		if len(r.Items) < defaultMaxItems {
			r.Items = append(r.Items, IssueItem{
				DocumentID: docID,
				Issue:      "version_mismatch",
				Detail:     fmt.Sprintf("current_version=%d max_version=%d", cur, maxV),
			})
		}
	}
	rows.Close()

	// Count total non-deleted docs for Total field.
	var total int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents WHERE status != 'deleted'`).Scan(&total)
	r.Total = total

	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkDocumentHash spot-checks content_hash by reading from storage.
func checkDocumentHash(ctx context.Context, db *sql.DB, store storage.ObjectStorage, sampleSize int) CheckResult {
	r := CheckResult{Name: "document_hash", Status: StatusOK}
	if sampleSize <= 0 {
		sampleSize = defaultSampleSize
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, storage_key, content_hash FROM documents WHERE status != 'deleted' LIMIT ?`,
		sampleSize)
	if err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "query_failed", Detail: err.Error()}}
		return r
	}
	defer rows.Close()

	for rows.Next() {
		var docID, key, storedHash string
		if err := rows.Scan(&docID, &key, &storedHash); err != nil {
			continue
		}
		r.Total++

		rc, err := store.Get(ctx, key)
		if err != nil {
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{
					DocumentID: docID, Issue: "storage_read_failed", Detail: err.Error(),
				})
			}
			continue
		}
		h := sha256.New()
		_, copyErr := io.Copy(h, rc)
		rc.Close()
		if copyErr != nil {
			continue
		}
		actual := fmt.Sprintf("%x", h.Sum(nil))
		if actual != storedHash {
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{
					DocumentID: docID, Issue: "content_hash_mismatch",
					Detail: fmt.Sprintf("stored=%s actual=%s", storedHash[:8], actual[:8]),
				})
			}
		}
	}
	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkFTSIndex verifies FTS index coverage.
func checkFTSIndex(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "fts_index", Status: StatusOK}

	var total int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents WHERE status = 'active'`).Scan(&total)
	r.Total = total

	// Active docs missing from document_index_status.
	rows, err := db.QueryContext(ctx, `
		SELECT d.id FROM documents d
		WHERE d.status = 'active'
		AND NOT EXISTS (SELECT 1 FROM document_index_status dis WHERE dis.document_id = d.id)
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var docID string
			if err := rows.Scan(&docID); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{DocumentID: docID, Issue: "missing_index"})
			}
		}
		rows.Close()
	}

	// Indexed docs with stale content hash.
	staleRows, err := db.QueryContext(ctx, `
		SELECT d.id FROM documents d
		JOIN document_index_status dis ON dis.document_id = d.id
		WHERE d.status = 'active'
		AND dis.status = 'indexed'
		AND dis.content_hash != d.content_hash
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer staleRows.Close()
		for staleRows.Next() {
			var docID string
			if err := staleRows.Scan(&docID); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{DocumentID: docID, Issue: "stale_index"})
			}
		}
	}

	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkTags verifies document_tags referential integrity.
func checkTags(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "tags", Status: StatusOK}

	var total int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM document_tags`).Scan(&total)
	r.Total = total

	addDangling := func(query, issue string) {
		rows, err := db.QueryContext(ctx, query+" LIMIT ?", defaultMaxItems)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{EntityID: id, Issue: issue})
			}
		}
	}

	addDangling(`SELECT dt.document_id FROM document_tags dt
		WHERE NOT EXISTS (SELECT 1 FROM documents d WHERE d.id = dt.document_id)`,
		"dangling_document_id")
	addDangling(`SELECT dt.tag_id FROM document_tags dt
		WHERE NOT EXISTS (SELECT 1 FROM tags t WHERE t.id = dt.tag_id)`,
		"dangling_tag_id")

	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkCollections verifies collection_documents referential integrity.
func checkCollections(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "collections", Status: StatusOK}

	var total int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM collection_documents`).Scan(&total)
	r.Total = total

	addDangling := func(query, issue string) {
		rows, err := db.QueryContext(ctx, query+" LIMIT ?", defaultMaxItems)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{EntityID: id, Issue: issue})
			}
		}
	}

	addDangling(`SELECT cd.collection_id FROM collection_documents cd
		WHERE NOT EXISTS (SELECT 1 FROM collections c WHERE c.id = cd.collection_id)`,
		"dangling_collection_id")
	addDangling(`SELECT cd.document_id FROM collection_documents cd
		WHERE NOT EXISTS (SELECT 1 FROM documents d WHERE d.id = cd.document_id)`,
		"dangling_document_id")

	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkSources verifies document_sources referential integrity.
func checkSources(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "sources", Status: StatusOK}

	var total int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM document_sources`).Scan(&total)
	r.Total = total

	addDangling := func(query, issue string) {
		rows, err := db.QueryContext(ctx, query+" LIMIT ?", defaultMaxItems)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{EntityID: id, Issue: issue})
			}
		}
	}

	addDangling(`SELECT ds.document_id FROM document_sources ds
		WHERE NOT EXISTS (SELECT 1 FROM documents d WHERE d.id = ds.document_id)`,
		"dangling_document_id")
	addDangling(`SELECT ds.source_document_id FROM document_sources ds
		WHERE NOT EXISTS (SELECT 1 FROM documents d WHERE d.id = ds.source_document_id)`,
		"dangling_source_id")

	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkImports verifies import_job_items referential integrity.
func checkImports(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "imports", Status: StatusOK}

	var total int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM import_job_items`).Scan(&total)
	r.Total = total

	// Items with document_id pointing to non-existent document.
	rows, err := db.QueryContext(ctx, `
		SELECT iji.id, iji.document_id FROM import_job_items iji
		WHERE iji.document_id IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM documents d WHERE d.id = iji.document_id)
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var itemID, docID string
			if err := rows.Scan(&itemID, &docID); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{
					EntityID: itemID, Issue: "missing_document", Detail: docID,
				})
			}
		}
		rows.Close()
	}

	// Imported items without a document_id.
	rows2, err := db.QueryContext(ctx, `
		SELECT iji.id FROM import_job_items iji
		WHERE iji.status = 'imported' AND iji.document_id IS NULL
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var itemID string
			if err := rows2.Scan(&itemID); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{EntityID: itemID, Issue: "imported_without_doc"})
			}
		}
	}

	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkAIConsistency verifies AI table referential integrity.
func checkAIConsistency(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "ai", Status: StatusOK}

	// ai_tasks with missing output_document_id.
	rows1, err := db.QueryContext(ctx, `
		SELECT at.id, at.output_document_id FROM ai_tasks at
		WHERE at.output_document_id IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM documents d WHERE d.id = at.output_document_id)
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer rows1.Close()
		for rows1.Next() {
			var id, docID string
			if err := rows1.Scan(&id, &docID); err != nil {
				continue
			}
			r.Total++
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{EntityID: id, Issue: "missing_output_document", Detail: docID})
			}
		}
		rows1.Close()
	}

	// prompts with missing source_document_id.
	rows2, err := db.QueryContext(ctx, `
		SELECT p.id, p.source_document_id FROM prompts p
		WHERE p.source_document_id IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM documents d WHERE d.id = p.source_document_id)
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var id, docID string
			if err := rows2.Scan(&id, &docID); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{EntityID: id, Issue: "missing_source_document", Detail: docID})
			}
		}
		rows2.Close()
	}

	// document_ai_summaries with missing document.
	rows3, err := db.QueryContext(ctx, `
		SELECT das.id, das.document_id FROM document_ai_summaries das
		WHERE NOT EXISTS (SELECT 1 FROM documents d WHERE d.id = das.document_id)
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var id, docID string
			if err := rows3.Scan(&id, &docID); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{EntityID: id, Issue: "missing_summary_document", Detail: docID})
			}
		}
	}

	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// worstStatus returns the worst status across all check results.
func worstStatus(checks []CheckResult) string {
	worst := StatusOK
	for _, c := range checks {
		switch {
		case c.Status == StatusError:
			return StatusError
		case c.Status == StatusWarning && worst == StatusOK:
			worst = StatusWarning
		}
	}
	return worst
}
