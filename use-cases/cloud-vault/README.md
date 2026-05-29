# Markdown Cloud Vault

A single-binary Go + React application for managing Markdown documents with object storage versioning.

## Technology Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26 + Plumego |
| Database | SQLite (modernc.org/sqlite) |
| Object Storage | Local filesystem / Qiniu Kodo |
| ID Generation | ULID (oklog/ulid/v2) |
| Markdown Parsing | goldmark |
| Frontend | React + Vite + TypeScript |
| Editor | CodeMirror 6 |
| Styling | Tailwind CSS + shadcn/ui |
| Deployment | Single Go binary with embedded frontend |

## Local Development

### Prerequisites

- Go 1.26+
- Node.js 20+ and pnpm

### Start the Go backend (local storage, no frontend build needed)

```bash
go run ./cmd/server
```

Server starts at `http://localhost:8080`. The API is available at `/api/v1/`.

### Start the full stack with live reload

**Terminal 1 — Backend:**
```bash
go run ./cmd/server
```

**Terminal 2 — Frontend dev server:**
```bash
make dev
# or: cd web && pnpm dev
```

The Vite dev server proxies `/api/*` to `:8080`. Open `http://localhost:5173`.

## Configuration

Configuration is loaded from environment variables and an optional `.env` file (default: `.env`).

```bash
cp env.example .env
# edit .env as needed
```

| Variable | Default | Description |
|---|---|---|
| `APP_ADDR` | `:8080` | HTTP listen address |
| `DB_PATH` | `./data/app.db` | SQLite database path |
| `STORAGE_PROVIDER` | `local` | `local` or `qiniu` |
| `LOCAL_ROOT` | `./data/objects` | Root for local object storage |
| `APP_MAX_UPLOAD_SIZE_MB` | `10` | Maximum upload size in MB |
| `QINIU_ACCESS_KEY` | — | Qiniu access key |
| `QINIU_SECRET_KEY` | — | Qiniu secret key |
| `QINIU_BUCKET` | — | Qiniu bucket name |
| `QINIU_DOMAIN` | — | Qiniu CDN domain (e.g. `https://example.com`) |
| `QINIU_REGION` | `z0` | Qiniu region: `z0` 华东, `z1` 华北, `z2` 华南, `na0` 北美, `as0` 新加坡 |
| `QINIU_USE_HTTPS` | `true` | Whether to use HTTPS for Qiniu API calls |

## Using LocalStorage (default)

No configuration needed. Files are stored under `./data/objects/`:

```
./data/objects/docs/{doc_id}/current.md
./data/objects/docs/{doc_id}/versions/000001.md
```

## Using Qiniu Kodo

1. Create a bucket in [Qiniu Console](https://portal.qiniu.com/).
2. Set the following variables (in `.env` or environment):

```bash
STORAGE_PROVIDER=qiniu
QINIU_ACCESS_KEY=your-ak
QINIU_SECRET_KEY=your-sk
QINIU_BUCKET=your-bucket
QINIU_DOMAIN=https://your-cdn-domain.com
QINIU_REGION=z0
```

> **Note:** For private buckets, `Get` uses signed URLs valid for 1 hour. Public buckets work out of the box.

## API Reference

All endpoints are under `/api/v1`. Successful responses use `{"data": ...}`, errors use `{"error": {"code": "...", "message": "..."}}`.

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/documents` | List documents (`?q=`, `?limit=`, `?offset=`) |
| POST | `/api/v1/documents` | Create document |
| GET | `/api/v1/documents/:id` | Get document with content |
| PUT | `/api/v1/documents/:id` | Update document |
| DELETE | `/api/v1/documents/:id` | Soft-delete document |
| GET | `/api/v1/documents/:id/versions` | List document versions |
| GET | `/api/v1/documents/:id/versions/:version` | Get specific version content |

### Version conflict

`PUT /api/v1/documents/:id` requires `base_version` matching the current version. If the document was updated concurrently, a `409 Conflict` is returned:

```json
{"error": {"code": "DOCUMENT_VERSION_CONFLICT", "message": "document has been updated by another session"}}
```

## Building a Single Binary

```bash
make build
# or step by step:
make web-build     # builds frontend into internal/web/static/
make server-build  # embeds frontend and compiles Go binary → bin/markdown-vault

./bin/markdown-vault
```

## Database Schema

| Table | Purpose |
|---|---|
| `documents` | Document metadata, current version pointer |
| `document_versions` | Version history (storage key + hash per version) |
| `tags` | Tag definitions (reserved for V0.2) |
| `document_tags` | Document↔tag associations (reserved for V0.2) |
| `sync_jobs` | Background sync job queue (reserved for V0.2) |

## Object Storage Key Format

```
docs/{doc_id}/current.md           ← always points to latest version
docs/{doc_id}/versions/000001.md   ← immutable version snapshots
docs/{doc_id}/versions/000002.md
...
```

## V0.3: Full-Text Search & Knowledge Recall

V0.3 adds SQLite FTS5-powered full-text search across all document content.

### Features

- **Full-text search** — titles, headings, summaries, original paths, and cleaned body text are all searchable
- **Highlight snippets** — search results include `<mark>`-tagged excerpts from the best-matching passage
- **Advanced filters** — tag, status, source type, import job, favorites, date range
- **Paginated results** — `limit` / `offset` with a maximum of 100 per page
- **Search history** — recent queries stored and displayed; clearable
- **Background indexer** — a goroutine polls `document_index_status` every N seconds and indexes pending documents
- **Inline indexing** — non-import document saves are indexed synchronously (no delay)
- **Index status page** — shows totals by state (indexed / pending / failed / stale)
- **Reindex operations** — trigger reindexing for all, failed, stale, or a single document

### Search API

```
GET /api/v1/search?q=clickhouse+async_insert&tag=<tag_id>&limit=20&offset=0
```

Response:
```json
{
  "data": {
    "items": [
      {
        "id": "01...",
        "title": "ClickHouse 高频写入优化",
        "summary": "分析 async_insert、parts、merge...",
        "highlights": ["... <mark>async_insert</mark>=1 可以减少客户端小批量写入 ..."],
        "score": -1.23,
        "tags": ["ClickHouse", "数据库"],
        "original_path": "/notes/clickhouse/write.md",
        "updated_at": "2026-05-29T12:00:00Z"
      }
    ],
    "total": 42,
    "limit": 20,
    "offset": 0
  }
}
```

Query parameters:

| Parameter | Description |
|---|---|
| `q` | Full-text search query (space-separated terms = implicit AND) |
| `tag` | Filter by tag ID |
| `status` | `active` (default) / `archived` / `all` |
| `source_type` | `manual` / `imported` |
| `review_status` | `pending` / `reviewed` |
| `import_job_id` | Filter by import job |
| `is_favorite` | `1` / `0` |
| `from` / `to` | ISO datetime bounds on `updated_at` |
| `sort` | `relevance` (default) / `updated_at` / `title` |
| `order` | `asc` / `desc` |

### Index Management API

```
GET    /api/v1/search/index-status     # aggregate stats
POST   /api/v1/search/reindex          # trigger rebuild: {"scope":"all|failed|stale|document","document_id":"..."}
GET    /api/v1/search/history          # recent searches
DELETE /api/v1/search/history          # clear all history
```

### How to Rebuild the Index

From the UI — click the **Index** tab → **Reindex All**.

Via API:
```bash
curl -X POST http://localhost:8080/api/v1/search/reindex \
  -H 'Content-Type: application/json' \
  -d '{"scope":"all"}'
```

The background indexer will process all documents within `SEARCH_INDEX_INTERVAL_SECONDS` (default 5 s).

### Index Status

| State | Meaning |
|---|---|
| `indexed` | Content is current in the FTS index |
| `pending` | Awaiting background indexer |
| `stale` | Document was updated after last indexing |
| `failed` | Indexer encountered an error (stored in `error_message`) |

### Search Configuration

All configuration uses environment variables (or `.env` file):

| Variable | Default | Description |
|---|---|---|
| `SEARCH_ENABLED` | `true` | Enable FTS search |
| `SEARCH_INDEX_ON_SAVE` | `true` | Index document immediately on create/update |
| `SEARCH_INDEX_ON_IMPORT` | `background` | `inline` / `background` / `disabled` |
| `SEARCH_INDEX_BATCH_SIZE` | `100` | Documents processed per indexer tick |
| `SEARCH_INDEX_INTERVAL_SECONDS` | `5` | Background indexer polling interval |
| `SEARCH_MAX_CONTENT_SIZE_MB` | `5` | Files larger than this are indexed by title/headings only |
| `SEARCH_SNIPPET_TOKENS` | `20` | Tokens in each highlight snippet |
| `SEARCH_HISTORY_LIMIT` | `100` | Maximum search history entries retained |

### Search-to-Editor Workflow

Clicking **Open** on any search result switches the UI to the **Vault** tab, loads the document, and highlights every occurrence of the search query inside the CodeMirror editor:

1. The **Search** page calls `onOpenDocument(id, query)` when the user clicks Open.
2. `App.tsx` stores `{ id, query }` in `openDoc` state and switches `page` to `"vault"`.
3. `VaultPage` receives `initialDocId` and `highlightQuery` as props.
4. A `useEffect` loads the document and stores the query in `activeQuery` state.
5. `MarkdownEditor` receives `highlightQuery={activeQuery}` and applies a `ViewPlugin` that scans the document text and wraps every case-insensitive match with a `cm-search-highlight` decoration (yellow background, `#fef08a`).
6. A second `useEffect` inside the editor forces a decoration redraw and calls `EditorView.scrollIntoView` on the position of the first match, centering it in the viewport.

The highlight is non-destructive — it never modifies document content and disappears when `highlightQuery` is cleared (e.g. when the user navigates away or opens a different document from the sidebar).

### Current Limitations (V0.3)

- Uses SQLite FTS5 with `unicode61` tokenizer — Chinese word segmentation is character-level, not word-level; search for individual characters or short phrases works, multi-word Chinese phrase search may have lower recall
- No semantic / vector search
- No AI question-answering
- No similar-document recommendations
- For very large corpora (100k+ documents) consider tuning `SEARCH_INDEX_BATCH_SIZE` and `SEARCH_INDEX_INTERVAL_SECONDS` based on available I/O

## Database Schema

| Table | Purpose |
|---|---|
| `documents` | Document metadata, current version pointer |
| `document_versions` | Version history (storage key + hash per version) |
| `tags` | Tag definitions |
| `document_tags` | Document↔tag associations |
| `import_jobs` | Batch import job tracking |
| `import_job_items` | Per-file import progress |
| `document_metadata` | Extracted metadata (headings JSON, code languages) |
| `document_fts` | SQLite FTS5 virtual table for full-text search |
| `document_index_status` | Per-document indexing state and hash |
| `search_history` | Recent search queries |

## V0.4: Organize, Deduplicate & Collections

V0.4 adds rule-based document organization: duplicate detection, similarity analysis, tag suggestions, topic clustering, quality scoring, prompt candidate identification, and a collection system.

No AI, no vector database, no external search engine — all features use SQLite + lightweight heuristics.

### Features

- **Exact duplicate detection** — groups documents sharing the same content hash; user chooses which to keep and what to do with the rest (archive / mark duplicate / ignore)
- **Near-duplicate & similarity detection** — Jaccard similarity on text shingles within buckets (same tag, import job, directory, title prefix); configurable thresholds
- **Similar documents sidebar** — per-document similar doc list; ignore or confirm each pair
- **Collections** — named groups of document references; add, remove, reorder, note; create from search results or topic conversion
- **Tag suggestions** — rule-based candidates extracted from path, title, headings; user confirms each suggestion
- **Topic clustering** — rule-based topics from existing tags and path directories; convert any topic to a collection
- **Review queue** — unified inbox for duplicates, similar pairs, tag suggestions, prompt candidates, low-quality docs; bulk Run All
- **Prompt candidate detection** — keyword heuristics to flag documents that look like LLM prompts
- **Quality scoring** — rule-based score (0–100) based on title length, word count, headings, code blocks, favorite status, duplicate status

### Duplicate Detection

```bash
# Via UI: Duplicates tab → Detect Duplicates
# Via API:
curl -X POST http://localhost:8080/api/v1/organize/detect-duplicates
curl http://localhost:8080/api/v1/organize/duplicates

# Resolve: keep one, archive the rest
curl -X POST http://localhost:8080/api/v1/organize/duplicates/resolve \
  -H 'Content-Type: application/json' \
  -d '{"keep_document_id":"01...","duplicate_document_ids":["01..."],"action":"archive"}'
```

`action` values: `archive` | `mark_duplicate` | `ignore`

### Similarity Detection

```bash
curl -X POST http://localhost:8080/api/v1/organize/detect-similarity
curl http://localhost:8080/api/v1/documents/{id}/similar
curl -X POST http://localhost:8080/api/v1/organize/similarity/{id}/ignore
curl -X POST http://localhost:8080/api/v1/organize/similarity/{id}/confirm
```

### Collections API

```
GET    /api/v1/collections
POST   /api/v1/collections
GET    /api/v1/collections/:id
PUT    /api/v1/collections/:id
DELETE /api/v1/collections/:id
POST   /api/v1/collections/:id/documents
DELETE /api/v1/collections/:id/documents/:document_id
PUT    /api/v1/collections/:id/documents/reorder
POST   /api/v1/collections/from-search
```

### Tag Suggestions API

```
POST /api/v1/organize/suggest-tags
GET  /api/v1/documents/:id/tag-suggestions
POST /api/v1/tag-suggestions/:id/accept
POST /api/v1/tag-suggestions/:id/reject
POST /api/v1/tag-suggestions/batch/accept
```

### Topics API

```
POST /api/v1/organize/build-topics
GET  /api/v1/topics
GET  /api/v1/topics/:id
```

To convert a topic to a collection, call `POST /api/v1/collections/from-search` with the topic's document IDs.

### Review Queue API

```
GET /api/v1/review/queue?type=duplicates|similar|tag_suggestions|prompt_candidates|low_quality
```

### Organize Configuration

| Variable | Default | Description |
|---|---|---|
| `ORGANIZE_DUPLICATE_DETECTION` | `true` | Enable exact duplicate detection |
| `ORGANIZE_SIMILARITY_DETECTION` | `true` | Enable near-duplicate detection |
| `ORGANIZE_TAG_SUGGESTION` | `true` | Enable tag suggestions |
| `ORGANIZE_TOPIC_BUILD` | `true` | Enable topic building |
| `ORGANIZE_NEAR_DUPLICATE_THRESHOLD` | `0.85` | Jaccard threshold for near-duplicate |
| `ORGANIZE_RELATED_THRESHOLD` | `0.70` | Jaccard threshold for related |
| `ORGANIZE_MAX_COMPARE_PER_BUCKET` | `1000` | Max pairs compared per bucket |
| `ORGANIZE_AUTO_ARCHIVE_DUPLICATES` | `false` | Never auto-archives without user confirmation |
| `ORGANIZE_AUTO_APPLY_TAG_SUGGESTIONS` | `false` | Never auto-applies tags without confirmation |
| `ORGANIZE_PROMPT_CANDIDATE_DETECTION` | `true` | Enable prompt candidate identification |

### V0.4 Database Tables

| Table | Purpose |
|---|---|
| `document_similarity` | Detected similarity pairs (exact / near / related) |
| `collections` | Named document collections |
| `collection_documents` | Collection membership with sort order |
| `document_sources` | Document provenance links |
| `tag_suggestions` | Pending / accepted / rejected tag proposals |
| `topics` | Rule-derived topic clusters |
| `topic_documents` | Topic membership with score |
| `organize_jobs` | Long-running organize operation history |
| `document_fingerprints` | Lightweight text fingerprints for similarity |

New columns: `documents.quality_score`, `document_metadata.is_prompt_candidate`, `document_metadata.prompt_score`

### V0.4 Current Limitations

- V0.4 does **not** use AI, large language models, or vector databases
- Near-duplicate detection uses Jaccard similarity on text shingles — may produce false positives and false negatives; tune thresholds via config
- Similarity is only computed within buckets (same tag/import job/directory/title prefix) to avoid O(N²) comparisons
- System **never** automatically deletes documents
- Tag suggestions and similarity resolutions always require explicit user confirmation
- For very large corpora (50k+ documents) tune `ORGANIZE_MAX_COMPARE_PER_BUCKET` and `ORGANIZE_SIMILARITY_BATCH_SIZE`

## V0.5: AI-Assisted Organization, Q&A, and Knowledge Reconstruction

V0.5 adds an AI task queue, per-document summaries, document-grounded Q&A, and prompt extraction. All AI operations are opt-in and traceable to their source documents.

### Privacy Constraints (Hard Rules)

1. **No whole-library chat** — Q&A answers are grounded only in explicitly selected documents
2. **Documents not sent by default** — the AI receives only documents the user explicitly selects
3. **No content in logs** — full Markdown content is never written to server logs
4. **No hardcoded API keys** — keys come from `AI_API_KEY` env var only
5. **AI off by default** — `AI_ENABLED=false` until explicitly turned on
6. **Source tracking** — every AI-generated document records its source document IDs in `document_sources`
7. **Saveable output** — all AI outputs are saved as Markdown documents in the vault

### AI Features

- **Document Summary** — enqueue from the Vault tab; result saved as a new Markdown document with source link
- **Q&A** — ask a question grounded in selected documents; answer with citations; result saved as Markdown
- **Prompt Extraction** — extract a reusable LLM prompt from any document; saved to the Prompt Library
- **Prompt Library** — browse, copy, filter by scenario, delete extracted prompts
- **AI Task Queue** — track pending/running/completed/failed tasks; cancel pending tasks

### AI Configuration

| Variable | Default | Description |
|---|---|---|
| `AI_ENABLED` | `false` | Enable AI features (must be explicitly set to `true`) |
| `AI_PROVIDER` | `local_mock` | `local_mock` \| `openai_compatible` |
| `AI_BASE_URL` | — | OpenAI-compatible endpoint (e.g. `https://api.openai.com/v1`) |
| `AI_API_KEY` | — | API key for the provider |
| `AI_MODEL` | `gpt-4o-mini` | Model name to use |
| `AI_MAX_CONTEXT_TOKENS` | `8000` | Max tokens to include in context window |
| `AI_MAX_RETRIES` | `2` | Max task retries before marking failed |
| `AI_TASK_WORKERS` | `2` | Number of background worker goroutines |
| `AI_SUMMARY_ENABLED` | `true` | Enable document summary tasks |
| `AI_QA_ENABLED` | `true` | Enable Q&A tasks |
| `AI_PROMPT_EXTRACT_ENABLED` | `true` | Enable prompt extraction tasks |

### AI API

```
POST /api/v1/ai/tasks/summary          # enqueue summary: {"document_id":"..."}
POST /api/v1/ai/tasks/qa               # enqueue Q&A: {"question":"...","document_ids":["..."]}
POST /api/v1/ai/tasks/prompt-extract   # enqueue prompt extraction: {"document_id":"..."}
GET  /api/v1/ai/tasks                  # list tasks (?status=pending|running|completed|failed)
GET  /api/v1/ai/tasks/:id              # get task
POST /api/v1/ai/tasks/:id/cancel       # cancel pending task
GET  /api/v1/ai/documents/:id/summary  # get AI summary for a document
GET  /api/v1/ai/prompts                # list prompt library (?scenario=...)
GET  /api/v1/ai/prompts/:id            # get prompt
DELETE /api/v1/ai/prompts/:id          # delete prompt
```

### V0.5 Database Tables

| Table | Purpose |
|---|---|
| `ai_tasks` | AI task queue (pending → running → completed/failed/cancelled) |
| `document_ai_summaries` | Structured per-document AI summaries |
| `prompts` | Prompt library (extracted or manually created) |
| `document_chunks` | Heading-split chunks for context assembly |

### V0.5 Current Limitations

- No vector search or semantic similarity — context is assembled by selecting documents explicitly
- No streaming responses — tasks are asynchronous; poll task status or refresh the AI Tasks page
- No whole-library chat — by design; always select specific documents
- `local_mock` provider returns deterministic stub responses; switch to `openai_compatible` for real AI

## V0.6: System Observability, Testing & Benchmarking

V0.6 focuses on production hardening: system health monitoring, consistency checks, comprehensive testing, and performance benchmarking.

### New Features

#### System Observability

- **GET `/api/v1/system/health`** — Overall system health status
  - Database connectivity
  - Storage availability
  - Search index status
  - AI provider status

- **GET `/api/v1/system/stats`** — Aggregate statistics
  - Document, collection, tag counts
  - Storage usage
  - AI task queue depth
  - Import job statistics

- **POST `/api/v1/system/doctor`** — Consistency checks
  - Storage object integrity (missing files)
  - Document version consistency
  - Content hash verification
  - FTS index coverage
  - Tag reference integrity
  - Collection reference integrity
  - Source document integrity
  - Import job consistency
  - AI task consistency

#### Testing Infrastructure

- **Go tests** — 7 test suites covering core functionality
  - `internal/database/migrate_test.go` — Migration idempotency and table creation
  - `internal/storage/local_test.go` — Local storage operations
  - `internal/document/service_test.go` — Document CRUD and versioning
  - `internal/search/index_test.go` — FTS indexing and search
  - `internal/organize/duplicate_test.go` — Duplicate detection
  - `internal/ai/task_test.go` — AI task queue and processing
  - `internal/system/service_test.go` — System health and doctor checks

- **Test fixtures** — 11 Markdown files in `e2e/testdata/markdown/`
  - Simple, complex, frontmatter, code blocks
  - Links, images, tables
  - Duplicate pairs for testing
  - Similar documents for testing
  - Large documents for performance testing

- **E2E tests** — Playwright smoke tests
  - Navigation to all pages
  - API endpoint verification
  - Health, stats, doctor endpoints

#### Performance Benchmarking

- **`cmd/bench/main.go`** — Benchmark tool
  - Configurable document count
  - Search query performance
  - Indexing throughput
  - JSON output for automation

```bash
# Run benchmark with 1000 documents and 20 search queries
go run ./cmd/bench --docs 1000 --queries 20

# JSON output for CI
go run ./cmd/bench --docs 500 --queries 10 --json
```

### Makefile Targets

```bash
make test          # Run all tests
make test-go       # Run Go tests
make test-web      # Run frontend type-check and build
make test-e2e      # Run Playwright E2E tests (requires: cd e2e && pnpm install && pnpm exec playwright install)
make doctor        # Run doctor check against running server
make bench         # Run benchmark (1000 docs, 20 queries)
```

### System Health API

Check system health:

```bash
curl http://localhost:8080/api/v1/system/health
```

Response:
```json
{
  "status": "ok",
  "database": "ok",
  "storage": "ok",
  "search": "ok",
  "ai": "disabled"
}
```

### System Stats API

Get aggregate statistics:

```bash
curl http://localhost:8080/api/v1/system/stats
```

Response:
```json
{
  "documents": 150,
  "versions": 420,
  "collections": 12,
  "tags": 45,
  "import_jobs": 3,
  "indexed_documents": 150,
  "pending_indexes": 0,
  "failed_indexes": 0,
  "ai_tasks": 5,
  "prompts": 8
}
```

### Doctor API

Run consistency checks:

```bash
# Run all checks
curl -X POST http://localhost:8080/api/v1/system/doctor \
  -H 'Content-Type: application/json' \
  -d '{"checks":[]}'

# Run specific checks
curl -X POST http://localhost:8080/api/v1/system/doctor \
  -H 'Content-Type: application/json' \
  -d '{"checks":["storage_objects","document_versions","fts_index"]}'
```

Response:
```json
{
  "status": "ok",
  "checks": [
    {
      "name": "storage_objects",
      "status": "ok",
      "total": 150,
      "failed": 0,
      "items": []
    },
    {
      "name": "document_versions",
      "status": "ok",
      "total": 150,
      "failed": 0,
      "items": []
    }
  ]
}
```

### Running Tests

```bash
# Run all Go tests
go test ./...

# Run specific package tests
go test ./internal/document
go test ./internal/search
go test ./internal/system

# Run with coverage
go test -cover ./...

# Run E2E tests
cd e2e
pnpm install
pnpm exec playwright install
pnpm test
```

### Benchmark Results

Example benchmark output:

```
Cloud Vault Benchmark Results
============================
Documents created: 1000
Create duration: 2.345s (2.35 ms/doc)
Search queries: 20
Total search duration: 156ms
Average search time: 7.8ms
Indexed documents: 1000
Index coverage: 100.0%
```

JSON output:
```json
{
  "doc_count": 1000,
  "create_duration_ms": 2345,
  "search_queries": 20,
  "search_duration_ms": 156,
  "avg_search_ms": 7,
  "indexed_docs": 1000
}
```

### V0.6 Current Limitations

- E2E tests require manual Playwright installation
- Benchmark uses local storage only (no cloud storage benchmarking)
- Doctor checks are read-only (no auto-repair)
- No scheduled doctor runs (must be triggered manually)
- No metrics export (Prometheus/Graphite)

## Current Limitations

- No multi-user collaboration
- No authentication / authorization
- No semantic or vector search
