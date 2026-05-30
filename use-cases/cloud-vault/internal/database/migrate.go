package database

import (
	"database/sql"
	"fmt"
	"time"
)

type migration struct {
	version int
	up      func(tx *sql.Tx) error
}

var migrations = []migration{
	{version: 1, up: migrate001},
	{version: 2, up: migrate002},
	{version: 3, up: migrate003},
	{version: 4, up: migrate004},
	{version: 5, up: migrate005},
	{version: 6, up: migrate006},
	{version: 7, up: migrate007},
}

// Migrate applies all pending migrations in ascending version order.
func (db *DB) Migrate() error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT    NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		applied, err := isMigrationApplied(db.DB, m.version)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if applied {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.version, err)
		}

		if err := m.up(tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("run migration %d: %w", m.version, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
			m.version, time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}

	return nil
}

func isMigrationApplied(db *sql.DB, version int) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	return count > 0, err
}

// addColumnIfNotExists adds a column only when absent.
// SQLite lacks ADD COLUMN IF NOT EXISTS, so we check PRAGMA table_info first.
func addColumnIfNotExists(tx *sql.Tx, table, column, definition string) error {
	rows, err := tx.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("pragma table_info %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid, notNull, pk int
		var name, typ string
		var dfltValue *string
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition))
	return err
}

// migrate001 applies the V0.1 initial schema.
// All CREATE TABLE IF NOT EXISTS — safe to reapply on any database state.
func migrate001(tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS documents (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  slug TEXT,
  original_path TEXT,
  storage_key TEXT NOT NULL,
  current_version INTEGER NOT NULL DEFAULT 1,
  content_hash TEXT NOT NULL,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  word_count INTEGER NOT NULL DEFAULT 0,
  line_count INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  sync_status TEXT NOT NULL DEFAULT 'synced',
  is_favorite INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  uploaded_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_documents_updated_at  ON documents(updated_at);
CREATE INDEX IF NOT EXISTS idx_documents_status      ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_sync_status ON documents(sync_status);
CREATE INDEX IF NOT EXISTS idx_documents_title       ON documents(title);
CREATE TABLE IF NOT EXISTS document_versions (
  id TEXT PRIMARY KEY,
  document_id TEXT NOT NULL,
  version INTEGER NOT NULL,
  storage_key TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  note TEXT,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_document_versions_doc_version
  ON document_versions(document_id, version);
CREATE INDEX IF NOT EXISTS idx_document_versions_document_id
  ON document_versions(document_id);
CREATE TABLE IF NOT EXISTS tags (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS document_tags (
  document_id TEXT NOT NULL,
  tag_id TEXT NOT NULL,
  PRIMARY KEY(document_id, tag_id),
  FOREIGN KEY(document_id) REFERENCES documents(id),
  FOREIGN KEY(tag_id) REFERENCES tags(id)
);
CREATE TABLE IF NOT EXISTS sync_jobs (
  id TEXT PRIMARY KEY,
  document_id TEXT,
  job_type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  error_message TEXT,
  retry_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_status ON sync_jobs(status);
`)
	return err
}

// migrate002 adds V0.2 import and management columns/tables.
func migrate002(tx *sql.Tx) error {
	docCols := []struct{ name, def string }{
		{"source_type", "TEXT NOT NULL DEFAULT 'manual'"},
		{"import_job_id", "TEXT"},
		{"imported_at", "TEXT"},
		{"summary", "TEXT"},
		{"heading_text", "TEXT"},
		{"review_status", "TEXT NOT NULL DEFAULT 'pending'"},
	}
	for _, col := range docCols {
		if err := addColumnIfNotExists(tx, "documents", col.name, col.def); err != nil {
			return fmt.Errorf("add documents.%s: %w", col.name, err)
		}
	}

	tagCols := []struct{ name, def string }{
		{"color", "TEXT"},
		{"source", "TEXT"},
	}
	for _, col := range tagCols {
		if err := addColumnIfNotExists(tx, "tags", col.name, col.def); err != nil {
			return fmt.Errorf("add tags.%s: %w", col.name, err)
		}
	}

	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS import_jobs (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  source_path TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  total_count INTEGER NOT NULL DEFAULT 0,
  processed_count INTEGER NOT NULL DEFAULT 0,
  success_count INTEGER NOT NULL DEFAULT 0,
  failed_count INTEGER NOT NULL DEFAULT 0,
  skipped_count INTEGER NOT NULL DEFAULT 0,
  error_message TEXT,
  started_at TEXT,
  completed_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_import_jobs_status ON import_jobs(status);
CREATE TABLE IF NOT EXISTS import_job_items (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  file_path TEXT NOT NULL,
  document_id TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  error_message TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(job_id) REFERENCES import_jobs(id)
);
CREATE INDEX IF NOT EXISTS idx_import_job_items_job_id ON import_job_items(job_id);
CREATE INDEX IF NOT EXISTS idx_import_job_items_status ON import_job_items(status);
CREATE TABLE IF NOT EXISTS document_metadata (
  id TEXT PRIMARY KEY,
  document_id TEXT NOT NULL UNIQUE,
  headings TEXT,
  code_languages TEXT,
  code_block_count INTEGER NOT NULL DEFAULT 0,
  link_count INTEGER NOT NULL DEFAULT 0,
  image_count INTEGER NOT NULL DEFAULT 0,
  extracted_at TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_document_metadata_doc_id ON document_metadata(document_id);
CREATE INDEX IF NOT EXISTS idx_documents_source_type   ON documents(source_type);
CREATE INDEX IF NOT EXISTS idx_documents_import_job_id ON documents(import_job_id);
CREATE INDEX IF NOT EXISTS idx_documents_review_status ON documents(review_status);
CREATE INDEX IF NOT EXISTS idx_documents_is_favorite   ON documents(is_favorite);
`)
	return err
}

// migrate004 adds V0.4 organize: similarity, collections, topics, tag suggestions, quality scoring.
func migrate004(tx *sql.Tx) error {
	docCols := []struct{ name, def string }{
		{"quality_score", "REAL NOT NULL DEFAULT 0"},
	}
	for _, col := range docCols {
		if err := addColumnIfNotExists(tx, "documents", col.name, col.def); err != nil {
			return fmt.Errorf("add documents.%s: %w", col.name, err)
		}
	}
	metaCols := []struct{ name, def string }{
		{"is_prompt_candidate", "INTEGER NOT NULL DEFAULT 0"},
		{"prompt_score", "REAL NOT NULL DEFAULT 0"},
	}
	for _, col := range metaCols {
		if err := addColumnIfNotExists(tx, "document_metadata", col.name, col.def); err != nil {
			return fmt.Errorf("add document_metadata.%s: %w", col.name, err)
		}
	}
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS document_similarity (
  id               TEXT PRIMARY KEY,
  document_id_a    TEXT NOT NULL,
  document_id_b    TEXT NOT NULL,
  similarity_type  TEXT NOT NULL,
  similarity_score REAL NOT NULL,
  reason           TEXT,
  status           TEXT NOT NULL DEFAULT 'pending',
  created_at       TEXT NOT NULL,
  updated_at       TEXT NOT NULL,
  FOREIGN KEY(document_id_a) REFERENCES documents(id),
  FOREIGN KEY(document_id_b) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_document_similarity_a
  ON document_similarity(document_id_a);
CREATE INDEX IF NOT EXISTS idx_document_similarity_b
  ON document_similarity(document_id_b);
CREATE INDEX IF NOT EXISTS idx_document_similarity_score
  ON document_similarity(similarity_score DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_document_similarity_pair
  ON document_similarity(document_id_a, document_id_b, similarity_type);
CREATE TABLE IF NOT EXISTS collections (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  description TEXT,
  type        TEXT NOT NULL DEFAULT 'manual',
  status      TEXT NOT NULL DEFAULT 'active',
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_collections_type ON collections(type);
CREATE TABLE IF NOT EXISTS collection_documents (
  collection_id TEXT    NOT NULL,
  document_id   TEXT    NOT NULL,
  sort_order    INTEGER NOT NULL DEFAULT 0,
  note          TEXT,
  created_at    TEXT    NOT NULL,
  PRIMARY KEY(collection_id, document_id),
  FOREIGN KEY(collection_id) REFERENCES collections(id),
  FOREIGN KEY(document_id)   REFERENCES documents(id)
);
CREATE TABLE IF NOT EXISTS document_sources (
  document_id        TEXT NOT NULL,
  source_document_id TEXT NOT NULL,
  source_type        TEXT NOT NULL DEFAULT 'related',
  created_at         TEXT NOT NULL,
  PRIMARY KEY(document_id, source_document_id),
  FOREIGN KEY(document_id)        REFERENCES documents(id),
  FOREIGN KEY(source_document_id) REFERENCES documents(id)
);
CREATE TABLE IF NOT EXISTS tag_suggestions (
  id          TEXT PRIMARY KEY,
  document_id TEXT NOT NULL,
  tag_id      TEXT,
  tag_name    TEXT NOT NULL,
  source      TEXT NOT NULL,
  confidence  REAL NOT NULL DEFAULT 0,
  status      TEXT NOT NULL DEFAULT 'pending',
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id),
  FOREIGN KEY(tag_id)      REFERENCES tags(id)
);
CREATE INDEX IF NOT EXISTS idx_tag_suggestions_document ON tag_suggestions(document_id);
CREATE INDEX IF NOT EXISTS idx_tag_suggestions_status   ON tag_suggestions(status);
CREATE TABLE IF NOT EXISTS topics (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  description TEXT,
  source      TEXT NOT NULL DEFAULT 'rule',
  status      TEXT NOT NULL DEFAULT 'active',
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS topic_documents (
  topic_id    TEXT NOT NULL,
  document_id TEXT NOT NULL,
  score       REAL NOT NULL DEFAULT 0,
  source      TEXT NOT NULL DEFAULT 'rule',
  created_at  TEXT NOT NULL,
  PRIMARY KEY(topic_id, document_id),
  FOREIGN KEY(topic_id)    REFERENCES topics(id),
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE TABLE IF NOT EXISTS organize_jobs (
  id              TEXT PRIMARY KEY,
  job_type        TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'pending',
  total_items     INTEGER NOT NULL DEFAULT 0,
  processed_items INTEGER NOT NULL DEFAULT 0,
  failed_items    INTEGER NOT NULL DEFAULT 0,
  error_message   TEXT,
  started_at      TEXT,
  finished_at     TEXT,
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_organize_jobs_status ON organize_jobs(status);
CREATE TABLE IF NOT EXISTS document_fingerprints (
  document_id  TEXT PRIMARY KEY,
  title_norm   TEXT,
  simhash      TEXT,
  heading_hash TEXT,
  keyword_hash TEXT,
  created_at   TEXT NOT NULL,
  updated_at   TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
`)
	return err
}

// migrate005 adds V0.5 AI tables: task queue, summaries, prompts, chunks.
func migrate005(tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS ai_tasks (
  id                 TEXT PRIMARY KEY,
  task_type          TEXT NOT NULL,
  status             TEXT NOT NULL DEFAULT 'pending',
  input_json         TEXT NOT NULL,
  output_document_id TEXT,
  output_json        TEXT,
  provider           TEXT,
  model              TEXT,
  error_message      TEXT,
  retry_count        INTEGER NOT NULL DEFAULT 0,
  started_at         TEXT,
  finished_at        TEXT,
  created_at         TEXT NOT NULL,
  updated_at         TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_status ON ai_tasks(status);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_type   ON ai_tasks(task_type);
CREATE TABLE IF NOT EXISTS document_ai_summaries (
  id              TEXT PRIMARY KEY,
  document_id     TEXT NOT NULL,
  summary         TEXT NOT NULL,
  key_points_json TEXT,
  actions_json    TEXT,
  code_refs_json  TEXT,
  provider        TEXT,
  model           TEXT,
  created_at      TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_document_ai_summaries_document_id
  ON document_ai_summaries(document_id);
CREATE TABLE IF NOT EXISTS prompts (
  id                 TEXT PRIMARY KEY,
  title              TEXT NOT NULL,
  content            TEXT NOT NULL,
  source_document_id TEXT,
  model_hint         TEXT,
  scenario           TEXT,
  tags_json          TEXT,
  quality_score      REAL NOT NULL DEFAULT 0,
  created_at         TEXT NOT NULL,
  updated_at         TEXT NOT NULL,
  FOREIGN KEY(source_document_id) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_prompts_scenario ON prompts(scenario);
CREATE TABLE IF NOT EXISTS document_chunks (
  id           TEXT PRIMARY KEY,
  document_id  TEXT NOT NULL,
  chunk_index  INTEGER NOT NULL,
  heading_path TEXT,
  content      TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  token_count  INTEGER NOT NULL DEFAULT 0,
  created_at   TEXT NOT NULL,
  updated_at   TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_document_chunks_document_id
  ON document_chunks(document_id);
`)
	return err
}

// migrate003 adds V0.3 full-text search tables.
func migrate003(tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE VIRTUAL TABLE IF NOT EXISTS document_fts USING fts5(
  document_id UNINDEXED,
  title,
  original_path,
  summary,
  headings,
  content,
  tokenize = 'unicode61'
);
CREATE TABLE IF NOT EXISTS document_index_status (
  document_id     TEXT PRIMARY KEY,
  content_hash    TEXT NOT NULL,
  indexed_version INTEGER NOT NULL DEFAULT 0,
  status          TEXT NOT NULL DEFAULT 'pending',
  error_message   TEXT,
  indexed_at      TEXT,
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);
CREATE INDEX IF NOT EXISTS idx_document_index_status_status
  ON document_index_status(status);
CREATE TABLE IF NOT EXISTS search_history (
  id           TEXT PRIMARY KEY,
  query        TEXT NOT NULL,
  filters_json TEXT,
  result_count INTEGER NOT NULL DEFAULT 0,
  created_at   TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_search_history_created_at
  ON search_history(created_at DESC);
`)
	return err
}

// migrate006 adds V0.7 authentication tables.
func migrate006(tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS users (
  id                  TEXT PRIMARY KEY,
  username            TEXT NOT NULL UNIQUE,
  email               TEXT UNIQUE,
  display_name        TEXT,
  password_hash       TEXT NOT NULL,
  password_algo       TEXT NOT NULL DEFAULT 'pbkdf2-sha512',
  role                TEXT NOT NULL DEFAULT 'admin',
  status              TEXT NOT NULL DEFAULT 'active',
  locale              TEXT NOT NULL DEFAULT 'zh-CN',
  theme               TEXT NOT NULL DEFAULT 'system',
  last_login_at       TEXT,
  password_changed_at TEXT,
  created_at          TEXT NOT NULL,
  updated_at          TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email    ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status   ON users(status);
CREATE TABLE IF NOT EXISTS user_sessions (
  id           TEXT PRIMARY KEY,
  user_id      TEXT NOT NULL,
  session_hash TEXT NOT NULL UNIQUE,
  user_agent   TEXT,
  ip_address   TEXT,
  expires_at   TEXT NOT NULL,
  revoked_at   TEXT,
  created_at   TEXT NOT NULL,
  updated_at   TEXT NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id    ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_hash       ON user_sessions(session_hash);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE TABLE IF NOT EXISTS login_attempts (
  id              TEXT PRIMARY KEY,
  login_identifier TEXT NOT NULL,
  ip_address      TEXT,
  success         INTEGER NOT NULL DEFAULT 0,
  reason          TEXT,
  created_at      TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_login_attempts_identifier_created
  ON login_attempts(login_identifier, created_at);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_created
  ON login_attempts(ip_address, created_at);
CREATE TABLE IF NOT EXISTS security_events (
  id          TEXT PRIMARY KEY,
  user_id     TEXT,
  event_type  TEXT NOT NULL,
  ip_address  TEXT,
  user_agent  TEXT,
  detail_json TEXT,
  created_at  TEXT NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_security_events_user_id    ON security_events(user_id);
CREATE INDEX IF NOT EXISTS idx_security_events_type       ON security_events(event_type);
CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events(created_at);
`)
	return err
}

// migrate007 fixes document_ai_summaries schema: document_id must have UNIQUE constraint
// for ON CONFLICT(document_id) to work in UpsertSummary.
func migrate007(tx *sql.Tx) error {
	_, err := tx.Exec(`
-- Create new table with UNIQUE constraint on document_id
CREATE TABLE IF NOT EXISTS document_ai_summaries_new (
  id              TEXT PRIMARY KEY,
  document_id     TEXT NOT NULL UNIQUE,
  summary         TEXT NOT NULL,
  key_points_json TEXT,
  actions_json    TEXT,
  code_refs_json  TEXT,
  provider        TEXT,
  model           TEXT,
  created_at      TEXT NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id)
);

-- Copy existing data (keep latest per document_id if duplicates exist)
INSERT OR IGNORE INTO document_ai_summaries_new
  SELECT * FROM document_ai_summaries
  WHERE id IN (
    SELECT id FROM document_ai_summaries
    GROUP BY document_id
    ORDER BY created_at DESC
  );

-- Drop old table and rename new one
DROP TABLE IF EXISTS document_ai_summaries;
ALTER TABLE document_ai_summaries_new RENAME TO document_ai_summaries;

-- Recreate index
CREATE INDEX IF NOT EXISTS idx_document_ai_summaries_document_id
  ON document_ai_summaries(document_id);
`)
	return err
}
