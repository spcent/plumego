-- V0.5 AI: tasks, summaries, prompts, chunks.
-- document_sources already created in migration 004.

-- AI task queue.
CREATE TABLE IF NOT EXISTS ai_tasks (
  id                 TEXT PRIMARY KEY,
  task_type          TEXT NOT NULL,   -- document_summary | topic_summary | search_answer | prompt_extract
  status             TEXT NOT NULL DEFAULT 'pending',  -- pending | running | completed | failed | cancelled
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

-- Per-document AI summaries.
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

-- Prompt library.
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

-- Document chunks for context building.
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
