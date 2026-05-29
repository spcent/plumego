-- V0.4 organize: duplicate detection, similarity, collections, topics, tag suggestions, quality scoring.

-- Similarity pairs between documents.
CREATE TABLE IF NOT EXISTS document_similarity (
  id               TEXT PRIMARY KEY,
  document_id_a    TEXT NOT NULL,
  document_id_b    TEXT NOT NULL,
  similarity_type  TEXT NOT NULL,   -- exact_duplicate | near_duplicate | related
  similarity_score REAL NOT NULL,
  reason           TEXT,
  status           TEXT NOT NULL DEFAULT 'pending',  -- pending | ignored | confirmed
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

-- Named collections of documents.
CREATE TABLE IF NOT EXISTS collections (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  description TEXT,
  type        TEXT NOT NULL DEFAULT 'manual',  -- manual | search | topic
  status      TEXT NOT NULL DEFAULT 'active',
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_collections_type ON collections(type);

-- Documents belonging to a collection.
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

-- Document provenance / source relationships.
CREATE TABLE IF NOT EXISTS document_sources (
  document_id        TEXT NOT NULL,
  source_document_id TEXT NOT NULL,
  source_type        TEXT NOT NULL DEFAULT 'related',
  created_at         TEXT NOT NULL,
  PRIMARY KEY(document_id, source_document_id),
  FOREIGN KEY(document_id)        REFERENCES documents(id),
  FOREIGN KEY(source_document_id) REFERENCES documents(id)
);

-- Proposed tags for documents (require user confirmation).
CREATE TABLE IF NOT EXISTS tag_suggestions (
  id          TEXT  PRIMARY KEY,
  document_id TEXT  NOT NULL,
  tag_id      TEXT,
  tag_name    TEXT  NOT NULL,
  source      TEXT  NOT NULL,        -- path | title | heading | keyword | sibling
  confidence  REAL  NOT NULL DEFAULT 0,
  status      TEXT  NOT NULL DEFAULT 'pending',  -- pending | accepted | rejected
  created_at  TEXT  NOT NULL,
  updated_at  TEXT  NOT NULL,
  FOREIGN KEY(document_id) REFERENCES documents(id),
  FOREIGN KEY(tag_id)      REFERENCES tags(id)
);

CREATE INDEX IF NOT EXISTS idx_tag_suggestions_document ON tag_suggestions(document_id);
CREATE INDEX IF NOT EXISTS idx_tag_suggestions_status   ON tag_suggestions(status);

-- Topic clusters built from tags and path keywords.
CREATE TABLE IF NOT EXISTS topics (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  description TEXT,
  source      TEXT NOT NULL DEFAULT 'rule',  -- rule | manual
  status      TEXT NOT NULL DEFAULT 'active',
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
);

-- Documents assigned to topics.
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

-- Background organize job queue.
CREATE TABLE IF NOT EXISTS organize_jobs (
  id              TEXT PRIMARY KEY,
  job_type        TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'pending',  -- pending | running | done | failed
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

-- Lightweight fingerprints for similarity bucketing.
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

-- quality_score on documents
-- is_prompt_candidate and prompt_score on document_metadata
-- Added via addColumnIfNotExists in migrate004().
