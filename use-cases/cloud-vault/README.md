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

## Current Limitations

- No multi-user collaboration
- No authentication / authorization
- No semantic or vector search
