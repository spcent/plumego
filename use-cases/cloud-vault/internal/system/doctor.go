package system

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud-vault/internal/config"
	"cloud-vault/internal/storage"
)

const defaultMaxItems = 100
const defaultSampleSize = 100

// runCheck dispatches to the correct check function by name.
func runCheck(ctx context.Context, name string, db *sql.DB, store storage.ObjectStorage, cfg config.Config, sampleSize int) CheckResult {
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
	case "auth":
		return checkAuth(ctx, db)
	case "config_check":
		return checkConfig(ctx, cfg)
	case "data_dir_check":
		return checkDataDir(ctx, cfg)
	case "storage_writable_check":
		return checkStorageWritable(ctx, store, cfg)
	case "auth_security_check":
		return checkAuthSecurity(ctx, db, cfg)
	case "cookie_security_check":
		return checkCookieSecurity(ctx, cfg)
	case "qiniu_config_check":
		return checkQiniuConfig(ctx, store, cfg)
	case "backup_check":
		return checkBackup(ctx, cfg)
	case "migration_check":
		return checkMigration(ctx, db)
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

// checkAuth verifies authentication table integrity.
func checkAuth(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "auth", Status: StatusOK}

	// Check users table exists and has reasonable data
	var userCount int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&userCount)
	if err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "users_table_missing", Detail: err.Error()}}
		return r
	}
	r.Total = userCount

	// Check for users with empty password hashes
	rows, err := db.QueryContext(ctx, `
		SELECT id, username FROM users
		WHERE password_hash = '' OR password_hash IS NULL
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var userID, username string
			if err := rows.Scan(&userID, &username); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{
					EntityID: userID,
					Issue:    "empty_password_hash",
					Detail:   username,
				})
			}
		}
		rows.Close()
	}

	// Check for expired sessions that should be cleaned up
	var expiredSessions int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_sessions
		WHERE expires_at < datetime('now') AND revoked_at IS NULL
	`).Scan(&expiredSessions)
	if err == nil && expiredSessions > 0 {
		r.Items = append(r.Items, IssueItem{
			Issue:  "expired_sessions_pending_cleanup",
			Detail: fmt.Sprintf("%d expired sessions awaiting cleanup", expiredSessions),
		})
		r.Status = StatusWarning
	}

	// Check for orphaned sessions (user_id doesn't exist)
	orphanedRows, err := db.QueryContext(ctx, `
		SELECT us.id, us.user_id FROM user_sessions us
		WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = us.user_id)
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer orphanedRows.Close()
		for orphanedRows.Next() {
			var sessionID, userID string
			if err := orphanedRows.Scan(&sessionID, &userID); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{
					EntityID: sessionID,
					Issue:    "orphaned_session",
					Detail:   fmt.Sprintf("user_id=%s", userID),
				})
			}
		}
	}

	if r.Failed > 0 {
		r.Status = StatusWarning
	}
	return r
}

// checkConfig validates the configuration.
func checkConfig(ctx context.Context, cfg config.Config) CheckResult {
	r := CheckResult{Name: "config_check", Status: StatusOK}
	if err := config.ValidateConfig(cfg); err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "config_validation_failed", Detail: err.Error()}}
	}
	return r
}

// checkDataDir verifies the data directory exists and is writable.
func checkDataDir(ctx context.Context, cfg config.Config) CheckResult {
	r := CheckResult{Name: "data_dir_check", Status: StatusOK}

	dataDir := cfg.DB.Path
	if dataDir != "" {
		dataDir = filepath.Dir(dataDir)
	}

	if dataDir == "" {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "database_path_empty"}}
		return r
	}

	info, err := os.Stat(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create it
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				r.Status = StatusError
				r.Items = []IssueItem{{Issue: "data_dir_missing", Detail: err.Error()}}
				return r
			}
			r.Status = StatusOK
			r.Items = []IssueItem{{Issue: "data_dir_created", Detail: dataDir}}
			return r
		}
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "data_dir_stat_failed", Detail: err.Error()}}
		return r
	}

	if !info.IsDir() {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "data_path_not_directory", Detail: dataDir}}
		return r
	}

	// Check writability
	testFile := filepath.Join(dataDir, ".write_test")
	if f, err := os.Create(testFile); err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "data_dir_not_writable", Detail: err.Error()}}
	} else {
		f.Close()
		os.Remove(testFile)
	}

	return r
}

// checkStorageWritable verifies storage can write and read objects.
func checkStorageWritable(ctx context.Context, store storage.ObjectStorage, cfg config.Config) CheckResult {
	r := CheckResult{Name: "storage_writable_check", Status: StatusOK}

	testKey := ".doctor_write_test"
	testData := []byte("test")

	// Write
	if err := store.Put(ctx, testKey, bytes.NewReader(testData), int64(len(testData)), "text/plain"); err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "storage_write_failed", Detail: err.Error()}}
		return r
	}

	// Read
	exists, err := store.Exists(ctx, testKey)
	if err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "storage_exists_failed", Detail: err.Error()}}
		store.Delete(ctx, testKey)
		return r
	}
	if !exists {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "storage_object_not_found_after_write"}}
		return r
	}

	// Delete
	if err := store.Delete(ctx, testKey); err != nil {
		r.Status = StatusWarning
		r.Items = []IssueItem{{Issue: "storage_delete_failed", Detail: err.Error()}}
	}

	return r
}

// checkAuthSecurity verifies authentication security settings.
func checkAuthSecurity(ctx context.Context, db *sql.DB, cfg config.Config) CheckResult {
	r := CheckResult{Name: "auth_security_check", Status: StatusOK}

	if !cfg.Auth.Enabled {
		r.Status = StatusWarning
		r.Items = []IssueItem{{Issue: "auth_disabled", Detail: "authentication is disabled"}}
		return r
	}

	// Check for at least one admin user
	var adminCount int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&adminCount)
	if err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "admin_query_failed", Detail: err.Error()}}
		return r
	}
	if adminCount == 0 {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "no_admin_users", Detail: "at least one admin user required"}}
	}

	// Check for disabled users with active sessions
	rows, err := db.QueryContext(ctx, `
		SELECT us.id, us.user_id FROM user_sessions us
		JOIN users u ON u.id = us.user_id
		WHERE u.disabled = true
		AND us.revoked_at IS NULL
		AND us.expires_at > datetime('now')
		LIMIT ?`, defaultMaxItems)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sessionID, userID string
			if err := rows.Scan(&sessionID, &userID); err != nil {
				continue
			}
			r.Failed++
			if len(r.Items) < defaultMaxItems {
				r.Items = append(r.Items, IssueItem{
					EntityID: sessionID,
					Issue:    "disabled_user_active_session",
					Detail:   fmt.Sprintf("user_id=%s", userID),
				})
			}
		}
	}

	// Check for expired sessions pending cleanup
	var expiredSessions int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_sessions
		WHERE expires_at < datetime('now') AND revoked_at IS NULL
	`).Scan(&expiredSessions)
	if err == nil && expiredSessions > 0 {
		r.Items = append(r.Items, IssueItem{
			Issue:  "expired_sessions_pending_cleanup",
			Detail: fmt.Sprintf("%d expired sessions", expiredSessions),
		})
		if r.Status == StatusOK {
			r.Status = StatusWarning
		}
	}

	if r.Failed > 0 && r.Status != StatusError {
		r.Status = StatusWarning
	}
	return r
}

// checkCookieSecurity verifies cookie security settings.
func checkCookieSecurity(ctx context.Context, cfg config.Config) CheckResult {
	r := CheckResult{Name: "cookie_security_check", Status: StatusOK}

	if !cfg.Auth.Enabled {
		r.Status = StatusDisabled
		return r
	}

	if cfg.Auth.CookieName == "" {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "cookie_name_empty"}}
		return r
	}

	if !cfg.Auth.SecureCookie {
		r.Status = StatusWarning
		r.Items = []IssueItem{{Issue: "secure_cookie_disabled", Detail: "set secure_cookie=true for production"}}
	}

	return r
}

// checkQiniuConfig verifies Qiniu configuration when provider is qiniu.
func checkQiniuConfig(ctx context.Context, store storage.ObjectStorage, cfg config.Config) CheckResult {
	r := CheckResult{Name: "qiniu_config_check", Status: StatusOK}

	if cfg.Storage.Provider != "qiniu" {
		r.Status = StatusDisabled
		return r
	}

	// Check all required fields
	var missing []string
	if cfg.Qiniu.AccessKey == "" {
		missing = append(missing, "access_key")
	}
	if cfg.Qiniu.SecretKey == "" {
		missing = append(missing, "secret_key")
	}
	if cfg.Qiniu.Bucket == "" {
		missing = append(missing, "bucket")
	}
	if cfg.Qiniu.Domain == "" {
		missing = append(missing, "domain")
	}
	if cfg.Qiniu.Region == "" {
		missing = append(missing, "region")
	}

	if len(missing) > 0 {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "qiniu_fields_missing", Detail: fmt.Sprintf("missing: %v", missing)}}
		return r
	}

	// Try to probe storage (Exists call)
	_, err := store.Exists(ctx, ".qiniu_probe_test")
	if err != nil {
		r.Status = StatusWarning
		r.Items = []IssueItem{{Issue: "qiniu_probe_failed", Detail: err.Error()}}
	}

	return r
}

// checkBackup verifies backup directory and recent backups.
func checkBackup(ctx context.Context, cfg config.Config) CheckResult {
	r := CheckResult{Name: "backup_check", Status: StatusOK}

	backupDir := filepath.Join(filepath.Dir(cfg.DB.Path), "backups")

	info, err := os.Stat(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create it
			if err := os.MkdirAll(backupDir, 0755); err != nil {
				r.Status = StatusError
				r.Items = []IssueItem{{Issue: "backup_dir_missing", Detail: err.Error()}}
				return r
			}
			r.Items = []IssueItem{{Issue: "backup_dir_created", Detail: backupDir}}
			return r
		}
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "backup_dir_stat_failed", Detail: err.Error()}}
		return r
	}

	if !info.IsDir() {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "backup_path_not_directory", Detail: backupDir}}
		return r
	}

	// Check writability
	testFile := filepath.Join(backupDir, ".write_test")
	if f, err := os.Create(testFile); err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "backup_dir_not_writable", Detail: err.Error()}}
	} else {
		f.Close()
		os.Remove(testFile)
	}

	// Check for recent backups (within 7 days)
	entries, err := os.ReadDir(backupDir)
	if err == nil {
		var latest time.Time
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".zip" {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latest) {
				latest = info.ModTime()
			}
		}

		if latest.IsZero() {
			r.Items = append(r.Items, IssueItem{Issue: "no_backups_found"})
			if r.Status == StatusOK {
				r.Status = StatusWarning
			}
		} else if time.Since(latest) > 7*24*time.Hour {
			r.Items = append(r.Items, IssueItem{
				Issue:  "backup_stale",
				Detail: fmt.Sprintf("last backup: %s", latest.Format("2006-01-02")),
			})
			if r.Status == StatusOK {
				r.Status = StatusWarning
			}
		}
	}

	return r
}

// checkMigration verifies database schema is up to date.
func checkMigration(ctx context.Context, db *sql.DB) CheckResult {
	r := CheckResult{Name: "migration_check", Status: StatusOK}

	// Check schema_migrations table exists
	var version string
	err := db.QueryRowContext(ctx, `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
	if err != nil {
		r.Status = StatusError
		r.Items = []IssueItem{{Issue: "schema_migrations_missing", Detail: err.Error()}}
		return r
	}

	// Expected latest version (V0.8)
	expectedVersion := "008"
	if version != expectedVersion {
		r.Status = StatusWarning
		r.Items = []IssueItem{{
			Issue:  "schema_version_drift",
			Detail: fmt.Sprintf("current=%s expected=%s", version, expectedVersion),
		}}
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
