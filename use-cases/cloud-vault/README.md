# Markdown Cloud Vault

A local-first Markdown knowledge vault for your AI-generated notes.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0-green.svg)](docs/release-notes.md)

[Download Desktop App](https://vault.birdor.com/download) | [Deploy Server](docs/server-deploy.md) | [Documentation](docs/)

---

## Features

- **Local-first storage** — Your Markdown stays on your machine (or your own cloud storage)
- **Bulk import** — Import hundreds of Markdown files with scan preview, duplicate detection, and retry
- **Full-text search** — SQLite FTS5 powers instant search across titles, content, and metadata
- **Smart organization** — Collections, tags, review queue, and AI-powered topic grouping
- **Optional AI** — Summarize, question, and extract insights (configure your own AI provider)
- **Backup & restore** — Built-in backup with one-click restore
- **Desktop & server** — Native desktop app (macOS/Windows/Linux) or self-hosted server

## Quick Start

### Desktop App

1. Download the installer for your platform:
   - [macOS (Apple Silicon)](https://vault.birdor.com/downloads/Cloud-Vault-v1.0.0-macOS-arm64.dmg)
   - [macOS (Intel)](https://vault.birdor.com/downloads/Cloud-Vault-v1.0.0-macOS-amd64.dmg)
   - [Windows](https://vault.birdor.com/downloads/Cloud-Vault-v1.0.0-windows-amd64.exe)
   - [Linux](https://vault.birdor.com/downloads/Cloud-Vault-v1.0.0-linux-amd64.AppImage)

2. Install and launch the application
3. On first run, create an admin account
4. Import your Markdown files: System → Import → Select Directory
5. Start searching and organizing!

See [Desktop Guide](docs/desktop.md) for detailed instructions.

### Server Deployment

1. Download the server binary:
   ```bash
   curl -L -o cloud-vault https://vault.birdor.com/downloads/cloud-vault-v1.0.0-linux-amd64
   chmod +x cloud-vault
   ```

2. Create config file:
   ```bash
   cp config.example.toml config.toml
   # Edit config.toml as needed
   ```

3. Run the server:
   ```bash
   ./cloud-vault
   ```

4. Open http://localhost:8080 and create an admin account

See [Server Deployment Guide](docs/server-deploy.md) for Docker, systemd, and HTTPS setup.

## Configuration

Configuration is loaded from `config.toml`, `.env` file, or environment variables (in that order, env vars take precedence).

Key options:
- `server.addr` — HTTP listen address (default: `:8080`)
- `database.path` — SQLite database path (default: `./data/app.db`)
- `storage.provider` — `local` or `qiniu` (default: `local`)
- `auth.enabled` — Enable authentication (default: `false`)
- `ai.enabled` — Enable AI features (default: `false`)

See [Configuration Guide](docs/configuration.md) for all options.

## Importing Markdown

1. Navigate to System → Import
2. Select a directory containing Markdown files
3. Review scan preview (file count, size, duplicates)
4. Click "Start Import"
5. Monitor progress and retry failures

See [Import Guide](docs/import.md) for advanced options.

## Search & Organize

- **Search** — Use the search bar to find documents by title, content, or metadata
- **Collections** — Group related documents into collections
- **Tags** — Add tags for flexible categorization
- **Review queue** — Process unprocessed imports
- **AI topics** — Get AI-powered topic suggestions (if enabled)

See [Search Guide](docs/search.md) and [Organize Guide](docs/organize.md).

## AI Features (Optional)

AI features are disabled by default. To enable:

1. Configure an OpenAI-compatible provider in `config.toml`:
   ```toml
   [ai]
   enabled = true
   provider = "openai_compatible"
   base_url = "https://api.openai.com/v1"
   api_key = "your-api-key"
   model = "gpt-4"
   ```

2. Use AI features:
   - Summarize documents
   - Ask questions about selected documents
   - Extract prompts from conversations
   - Generate new documents from AI output

**Privacy**: Only documents you explicitly select are sent to the AI provider.

See [AI Guide](docs/ai.md) for details.

## Backup & Restore

### Create backup
```bash
curl -X POST http://localhost:8080/api/v1/system/backup \
  -H 'Content-Type: application/json' \
  -b cookie.txt
```

Or use the web UI: System → Backup → Create Backup

### Restore from backup
```bash
./cloud-vault restore --file backups/backup-20260530-120000.zip --data-dir ./data
```

See [Backup & Restore Guide](docs/backup-restore.md).

## Error Reporting & Diagnostics

If you encounter issues, export a diagnostic bundle:

1. Navigate to Settings → Diagnostics
2. Click "Export Diagnostic Bundle"
3. Download the generated zip file
4. Share with support (if applicable)

**Privacy**: Diagnostic bundles include redacted config, logs, and system info. They **do not** include Markdown content, database files, API keys, session tokens, or passwords.

See [Diagnostics Guide](docs/diagnostics.md).

## Privacy & Security

- **LocalStorage mode** — All data stays on your machine
- **QiniuStorage mode** — Data uploaded to your own Qiniu bucket (you control access)
- **AI features** — Optional, only send documents you explicitly select
- **Session cookies** — HttpOnly flag, secure flag recommended for production
- **Diagnostic bundles** — Exclude Markdown content and secrets

See [Privacy Policy](docs/privacy.md) and [Security Guide](docs/security.md).

## Development

### Prerequisites
- Go 1.26+
- Node.js 20+ and pnpm
- (Optional) Wails CLI for desktop builds: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Build from source

```bash
# Clone repository
git clone https://github.com/example/cloud-vault.git
cd cloud-vault

# Build server
make build

# Run server
./bin/markdown-vault

# Build desktop app (requires Wails CLI)
make desktop-build
```

### Run tests

```bash
# Go tests
make test

# Frontend tests
cd web && pnpm test

# E2E tests (requires Playwright)
make test-e2e
```

See [Development Guide](docs/development.md).

## License

[MIT](LICENSE)

## Current Limitations

- V1.0 does not support multi-user real-time collaboration
- V1.0 does not support cloud sync accounts
- Auto-update check only notifies, does not silently install
- AI features require user-configured model provider
- Approximate duplicate and topic detection are suggestions, may not be fully accurate
- QiniuStorage requires user to manage bucket permissions and backup strategy
- macOS/Windows installers may show security warnings if not signed (see [Installation Guide](docs/installation.md))

## Links

- [Documentation](docs/)
- [Release Notes](docs/release-notes.md)
- [Changelog](CHANGELOG.md)
- [Privacy Policy](docs/privacy.md)
- [Security](docs/security.md)
- [Landing Page Draft](landing/index.html)

---

# Developer Documentation

The sections below contain detailed developer documentation for all versions.

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
| `AUTH_ENABLED` | `false` | Enable authentication (see V0.7) |
| `AUTH_SESSION_TTL_HOURS` | `720` | Session TTL in hours (default 30 days) |
| `AUTH_COOKIE_NAME` | `cv_session` | Session cookie name |
| `AUTH_SECURE_COOKIE` | `false` | Use Secure flag on cookies (HTTPS only) |
| `AUTH_MAX_LOGIN_FAILURES` | `5` | Max failures before lockout |
| `AUTH_PASSWORD_MIN_LENGTH` | `12` | Minimum password length (≥ 8) |
| `AUTH_BOOTSTRAP_ADMIN_ENABLED` | `false` | Auto-create admin on first startup |
| `AUTH_BOOTSTRAP_ADMIN_USERNAME` | — | Bootstrap admin username |
| `AUTH_BOOTSTRAP_ADMIN_EMAIL` | — | Bootstrap admin email |
| `AUTH_BOOTSTRAP_ADMIN_PASSWORD` | — | Bootstrap admin password (set via env for security) |

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
| POST | `/api/v1/auth/login` | Login (public, returns session cookie) |
| POST | `/api/v1/auth/logout` | Logout and revoke session (protected) |
| POST | `/api/v1/auth/setup` | Create first admin user (public, one-time) |
| GET | `/api/v1/auth/status` | Check if system is initialized (public) |
| GET | `/api/v1/auth/me` | Get current user profile (protected) |
| PUT | `/api/v1/auth/me` | Update user profile (protected) |
| POST | `/api/v1/auth/change-password` | Change password (protected) |
| GET | `/api/v1/auth/sessions` | List active sessions (protected) |
| DELETE | `/api/v1/auth/sessions/:id` | Revoke specific session (protected) |
| POST | `/api/v1/auth/sessions/revoke-all` | Revoke all other sessions (protected) |
| GET | `/api/v1/auth/security-events` | List security events (protected) |
| POST | `/api/v1/system/backup` | Create backup (protected) |
| GET | `/api/v1/system/backups` | List backups (protected) |
| GET | `/api/v1/system/backups/:name/download` | Download backup (protected) |
| DELETE | `/api/v1/system/backups/:name` | Delete backup (protected) |
| POST | `/api/v1/system/restore` | Restore from backup (protected, requires `confirm=RESTORE`) |
| GET | `/api/v1/system/settings` | Get non-sensitive system settings (protected) |

**Note:** Protected endpoints require a valid session cookie. See V0.7 documentation for authentication details.

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
| `users` | User accounts with profile and preferences (V0.7) |
| `user_sessions` | Session tokens with metadata (V0.7) |
| `login_attempts` | Failed login tracking for rate limiting (V0.7) |
| `security_events` | Audit log for authentication actions (V0.7) |

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

## V0.7: Authentication, i18n & Theme System

V0.7 adds foundational productization features: session-based authentication, internationalization (3 locales), theme system (light/dark/system), and account management.

### Features

- **Session-based authentication** — secure cookie-based sessions with automatic renewal
- **Admin setup flow** — first-time setup wizard to create initial admin account
- **Login rate limiting** — configurable failed login attempt limits with automatic lockout
- **Password management** — secure password change with current password verification
- **Account settings** — update display name, email, locale, and theme preferences
- **Security page** — view active sessions, revoke sessions, view security events
- **Theme system** — light/dark/system theme with persistent preferences
- **i18n support** — English, 简体中文, 繁體中文 with automatic language detection
- **Route protection** — all API endpoints protected except public auth routes
- **Security events** — audit trail for login, logout, password changes, session revocation
- **Doctor auth checks** — system health checks for auth-related issues

### Getting Started

#### First-Time Setup

When you first access the application, you'll see a setup page to create your admin account:

1. Visit `http://localhost:8080`
2. Fill in username, email, and password (minimum 10 characters)
3. Click "Create Admin Account"
4. You'll be automatically logged in and redirected to the main application

#### Configuration

Auth settings in `config.toml`:

```toml
[auth]
enabled = true
session_ttl_hours = 168              # 7 days
cookie_name = "markdown_vault_session"
secure_cookie = false                # Set to true in production with HTTPS
max_login_failures = 5
login_failure_window_minutes = 15
lockout_minutes = 30
password_min_length = 10
bootstrap_admin_enabled = false      # Set to true for automated setup

[auth.bootstrap_admin]
username = "admin"
email = "admin@vault.birdor.com"
password = "Change-Me-Strong-Password-123"
```

#### Environment Variable Override

All auth settings can be overridden via environment variables:

```bash
AUTH_ENABLED=true
AUTH_SESSION_TTL_HOURS=168
AUTH_SECURE_COOKIE=true
AUTH_MAX_LOGIN_FAILURES=5
AUTH_LOGIN_FAILURE_WINDOW_MINUTES=15
AUTH_LOCKOUT_MINUTES=30
AUTH_PASSWORD_MIN_LENGTH=10
AUTH_COOKIE_NAME=markdown_vault_session
```

Bootstrap admin settings:

```bash
AUTH_BOOTSTRAP_ADMIN_ENABLED=true
AUTH_BOOTSTRAP_ADMIN_USERNAME=admin
AUTH_BOOTSTRAP_ADMIN_EMAIL=admin@vault.birdor.com
AUTH_BOOTSTRAP_ADMIN_PASSWORD=Change-Me-Strong-Password-123
```

### Authentication API

All auth endpoints are under `/api/v1/auth`.

#### Public Endpoints (no authentication required)

```http
POST /api/v1/auth/setup
```
Create initial admin account (only works when no users exist).

Request:
```json
{
  "username": "admin",
  "email": "admin@vault.birdor.com",
  "password": "StrongPassword123"
}
```

```http
GET /api/v1/auth/status
```
Check if system is initialized.

Response:
```json
{
  "data": {
    "initialized": true
  }
}
```

```http
POST /api/v1/auth/login
```
Login with username/email and password.

Request:
```json
{
  "username": "admin",
  "password": "StrongPassword123"
}
```

Response sets `Set-Cookie` header with session token.

```http
GET /api/v1/health
```
System health check (always public).

#### Protected Endpoints (require authentication)

```http
POST /api/v1/auth/logout
```
Logout and revoke current session.

```http
GET /api/v1/auth/me
```
Get current user profile.

```http
PUT /api/v1/auth/me
```
Update user profile.

Request:
```json
{
  "display_name": "Admin User",
  "email": "admin@vault.birdor.com",
  "locale": "en-US",
  "theme": "dark"
}
```

```http
POST /api/v1/auth/change-password
```
Change password (requires current password).

Request:
```json
{
  "current_password": "OldPassword123",
  "new_password": "NewStrongPassword456"
}
```

```http
GET /api/v1/auth/sessions
```
List all active sessions for current user.

Response:
```json
{
  "data": {
    "sessions": [
      {
        "id": "01K...",
        "user_agent": "Mozilla/5.0...",
        "ip_address": "192.168.1.100",
        "created_at": "2026-05-29T10:00:00Z",
        "expires_at": "2026-06-05T10:00:00Z"
      }
    ]
  }
}
```

```http
POST /api/v1/auth/sessions/revoke-all
```
Revoke all sessions except current.

All other API endpoints (`/api/v1/documents`, `/api/v1/search`, etc.) now require authentication and return `401 Unauthorized` if not logged in.

### Security Features

#### Password Requirements

- Minimum 10 characters (configurable via `password_min_length`)
- Must contain at least one uppercase letter
- Must contain at least one lowercase letter
- Must contain at least one digit
- Passwords are hashed using PBKDF2-SHA512 (210,000 iterations)

#### Session Security

- Session tokens stored in HttpOnly cookies (not accessible via JavaScript)
- Session tokens hashed using SHA-256 before storage
- Configurable session TTL (default 7 days)
- Secure cookie flag for HTTPS deployments
- SameSite=Lax to prevent CSRF

#### Rate Limiting

- Failed login attempts tracked per username
- After 5 failed attempts within 15 minutes, account locked for 30 minutes
- Rate limit counter resets after successful login
- All rate limit violations logged as security events

#### Security Events

All authentication actions are logged to `security_events` table:

- `login_success` / `login_failed`
- `logout`
- `password_changed`
- `session_revoked`
- `account_locked`
- `admin_setup` / `admin_bootstrap`

View security events in the Security page or via API:

```http
GET /api/v1/auth/security-events?limit=50&offset=0
```

### Theme System

Three theme modes available:

- **Light** — light background with dark text
- **Dark** — dark background with light text
- **System** — follow OS preference (via `prefers-color-scheme`)

Theme preference stored in user profile and synced to localStorage for instant application on page load.

Change theme via:
1. Account settings page → Theme dropdown
2. API: `PUT /api/v1/auth/me` with `{"theme": "dark"}`

### Internationalization (i18n)

Supported locales:

- `en-US` — English (US)
- `zh-CN` — 简体中文 (Simplified Chinese)
- `zh-TW` — 繁體中文 (Traditional Chinese)

Locale preference stored in user profile and synced to localStorage.

Change locale via:
1. Account settings page → Language dropdown
2. API: `PUT /api/v1/auth/me` with `{"locale": "zh-CN"}`

i18n covers:
- Navigation tabs (Vault, Search, Import, etc.)
- Common actions (Save, Cancel, Delete, etc.)
- Auth pages (Login, Setup, Account, Security)
- Form labels and error messages
- Theme names (Light, Dark, System)
- Locale names (English, 简体中文, 繁體中文)

### Frontend Pages

#### Setup Page
- Shown on first visit when no users exist
- Creates initial admin account
- Auto-login after successful setup

#### Login Page
- Username/email + password form
- Error messages for invalid credentials
- Rate limit warnings when account locked

#### Account Settings Page
- Update display name, email
- Change locale (language)
- Change theme (light/dark/system)
- Save button to persist changes

#### Security Page
- Change password (requires current password)
- View active sessions with device info
- Revoke individual sessions
- Revoke all other sessions button
- View security event audit log

### Database Schema

New tables in migration 006:

| Table | Purpose |
|---|---|
| `users` | User accounts with profile preferences |
| `user_sessions` | Active session tokens with metadata |
| `login_attempts` | Failed login tracking for rate limiting |
| `security_events` | Audit log for authentication actions |

### System Doctor Auth Checks

Run auth-related health checks:

```bash
curl -X POST http://localhost:8080/api/v1/system/doctor \
  -H 'Content-Type: application/json' \
  -d '{"checks":["auth"]}'
```

Checks:
- `users_table_exists` — verify users table exists
- `has_admin_user` — at least one admin user exists
- `expired_sessions` — count of expired but not revoked sessions
- `orphaned_sessions` — sessions referencing non-existent users
- `orphaned_security_events` — security events referencing non-existent users

### Security Best Practices

1. **Enable HTTPS in production**
   - Set `secure_cookie = true` in config
   - Configure reverse proxy (nginx/caddy) with SSL certificate
   - Redirect all HTTP traffic to HTTPS

2. **Use strong passwords**
   - Minimum 10 characters enforced
   - Consider requiring special characters for admin accounts

3. **Regular password rotation**
   - Change admin password periodically
   - All sessions automatically revoked after password change

4. **Monitor security events**
   - Check Security page regularly
   - Look for unusual login patterns or locations
   - Revoke suspicious sessions immediately

5. **Disable bootstrap after setup**
   - Set `bootstrap_admin_enabled = false` after initial setup
   - Prevents accidental admin account creation

6. **Backup database regularly**
   - `users` table contains hashed passwords
   - `user_sessions` contains active session tokens
   - `security_events` contains audit trail

### V0.7 Current Limitations

- Single-user or small-scale private deployment only
- No OAuth/SSO integration
- No role-based access control (RBAC)
- No multi-tenancy
- No email verification
- No password reset via email
- No two-factor authentication (2FA)
- Rate limiting is per-username only (no IP-based limiting)
- Session cleanup requires manual triggering or scheduled job
- i18n covers core UI but not all error messages
- Bootstrap admin requires server restart after config changes

### Migration from V0.6

If upgrading from V0.6:

1. Backup your database: `cp data/app.db data/app.db.backup`
2. Stop the running application
3. Run migration: `go run ./cmd/server` (migration 006 runs automatically)
4. Access the application — you'll see the setup page
5. Create your admin account
6. All existing documents and data remain intact

If you want to skip the setup page, enable bootstrap admin in `config.toml` before starting:

```toml
[auth]
enabled = true
bootstrap_admin_enabled = true

[auth.bootstrap_admin]
username = "admin"
email = "admin@vault.birdor.com"
password = "Change-Me-Strong-Password-123"
```

The bootstrap admin will be created automatically on startup, and you can login immediately.

## V0.8: Productization, Backup & Deployment

V0.8 focuses on production readiness: configuration validation, backup/restore, deployment packaging, deployment-aware doctor checks, and comprehensive regression testing.

### Configuration System

#### Loading Priority

Configuration is loaded in this order (later wins):

1. **Defaults** — built-in defaults
2. **config.toml** — TOML file (default: `./config.toml`, or `--config <path>`)
3. **.env file** — KEY=VALUE pairs (default: `./.env`)
4. **Environment variables** — `APP_*`, `AUTH_*`, `QINIU_*`, etc.
5. **CLI flags** — `--addr`, `--db-path`, etc.

```bash
# Copy example configs
cp config.example.toml config.toml
cp env.example .env

# Edit .env with your secrets
nano .env

# Run the server
./cloud-vault --config config.toml
```

#### Configuration Validation

On startup, Cloud Vault validates all configuration and refuses to start with invalid settings:

| Section | Required Fields |
|---|---|
| `[server]` | `addr` non-empty, port 1–65535 |
| `[database]` | `path` non-empty, parent dir writable |
| `[storage.local]` | `root` non-empty, writable |
| `[storage.qiniu]` | `access_key`, `secret_key`, `bucket`, `domain`, `region` all non-empty |
| `[auth]` (enabled) | `cookie_name`, `session_ttl_hours > 0`, `password_min_length >= 8` |
| `[ai]` (enabled + openai) | `base_url`, `api_key`, `model` non-empty |

Validation errors are structured with field paths — secrets are never logged.

#### Sensitive Field Masking

All log output automatically masks sensitive fields:

- `Qiniu.SecretKey` → `[REDACTED]`
- `AI.APIKey` → `[REDACTED]`
- `Auth.BootstrapAdmin.Password` → `[REDACTED]`

### Backup & Restore

#### Creating Backups

**Via UI:** System tab → Backup Management → Create Backup

**Via API:**
```bash
curl -X POST http://localhost:8080/api/v1/system/backup \
  -H "Cookie: session=<your-session-cookie>"
```

**Via CLI:**
```bash
make backup
```

Backups are stored as `backups/cloud-vault-backup-YYYYMMDD-HHMMSS.zip` containing:
- SQLite database (consistent snapshot via `VACUUM INTO`)
- Local storage objects (if provider=local)
- `manifest.json` with metadata

#### Listing & Downloading Backups

```bash
# List all backups
curl http://localhost:8080/api/v1/system/backups \
  -H "Cookie: session=<your-session-cookie>"

# Download a specific backup
curl -O http://localhost:8080/api/v1/system/backups/cloud-vault-backup-20260530-120000.zip/download \
  -H "Cookie: session=<your-session-cookie>"

# Delete a backup
curl -X DELETE http://localhost:8080/api/v1/system/backups/cloud-vault-backup-20260530-120000.zip \
  -H "Cookie: session=<your-session-cookie>"
```

#### Restoring from Backup

**Via CLI (recommended):**
```bash
# Stop the server first!
./cloud-vault restore --file backups/cloud-vault-backup-20260530-120000.zip --data-dir ./data
```

Steps:
1. Validates the backup zip exists and contains a valid manifest
2. Extracts `app.db` to `<data-dir>/app.db.new`
3. Extracts `objects/` to `<data-dir>/objects.new/`
4. Atomically renames to replace existing files
5. Prints "run doctor to verify"

**Via API (requires confirmation):**
```bash
curl -X POST http://localhost:8080/api/v1/system/restore \
  -H "Content-Type: application/json" \
  -H "Cookie: session=<your-session-cookie>" \
  -d '{"backup_name":"cloud-vault-backup-20260530-120000.zip","confirm":"RESTORE"}'
```

The API refuses if any active sessions exist (returns 409). Stop the server or revoke all sessions first.

> **Note:** Qiniu mode backups include only the manifest — bucket contents must be restored using Qiniu's own backup tools.

### Doctor Deployment Checks

V0.8 adds 8 deployment-aware checks to the doctor system:

| Check | What it verifies |
|---|---|
| `config_check` | All config sections pass `ValidateConfig()` |
| `data_dir_check` | Data directory exists and is writable |
| `storage_writable_check` | Can write/read/delete a test object |
| `auth_security_check` | Auth enabled, admin exists, no disabled users with sessions |
| `cookie_security_check` | `secure_cookie=true` (warns if false) |
| `qiniu_config_check` | All Qiniu fields set, bucket accessible |
| `backup_check` | Backups dir writable, last backup < 7 days |
| `migration_check` | Schema version matches expected latest |

Run deployment checks:

```bash
curl -X POST http://localhost:8080/api/v1/system/doctor \
  -H "Content-Type: application/json" \
  -H "Cookie: session=<your-session-cookie>" \
  -d '{"checks":["config_check","data_dir_check","storage_writable_check","auth_security_check","cookie_security_check","backup_check","migration_check"]}'
```

### Migration Upgrade Path

V0.8 includes comprehensive migration upgrade tests covering:

- `empty → latest` (fresh install)
- `v01 → latest` (documents + versions only)
- `v03 → latest` (+ FTS, search history)
- `v05 → latest` (+ AI tasks, prompts, sources)
- `v07 → latest` (+ users, sessions, security events)
- `latest → latest` (idempotency)

All migrations are idempotent — running `db.Migrate()` multiple times is safe.

### Release Build

Build a production release package:

```bash
make release VERSION=1.0.0
# or
./scripts/release.sh 1.0.0
```

Produces `dist/cloud-vault-1.0.0-linux-amd64.tar.gz` containing:
- `cloud-vault` binary (linux/amd64, statically linked)
- `config.example.toml`
- `.env.example`
- `README.md`

### Docker Deployment

**Build the image:**
```bash
docker build -t cloud-vault:latest .
```

**Run with docker-compose:**
```bash
cp docker-compose.yml docker-compose.override.yml
# Edit to set volumes, ports, environment
docker-compose up -d
```

**Run standalone:**
```bash
docker run -d \
  --name cloud-vault \
  -p 8080:8080 \
  -v ./data:/app/data \
  -v ./backups:/app/backups \
  -e AUTH_ENABLED=true \
  -e AUTH_SECURE_COOKIE=true \
  cloud-vault:latest
```

The Docker image uses a multi-stage build:
1. **frontend-builder** — Node 20 Alpine, builds React app
2. **backend-builder** — Go 1.26 Alpine, compiles Go binary
3. **runtime** — Alpine Linux, minimal footprint

### systemd Installation

Install as a systemd service on Linux:

```bash
# Build release binary
make release

# Create system user
sudo useradd --system --shell /sbin/nologin cloud-vault

# Install
sudo mkdir -p /opt/cloud-vault
sudo cp dist/cloud-vault/cloud-vault /opt/cloud-vault/
sudo cp config.example.toml /opt/cloud-vault/config.toml
sudo mkdir -p /opt/cloud-vault/data /opt/cloud-vault/backups
sudo chown -R cloud-vault:cloud-vault /opt/cloud-vault

# Edit config
sudo nano /opt/cloud-vault/config.toml

# Install service
sudo cp deploy/systemd/cloud-vault.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now cloud-vault
```

The systemd unit includes security hardening:
- `ProtectSystem=strict` — prevents writes outside designated paths
- `PrivateTmp=true` — isolates temporary files
- `NoNewPrivileges=true` — prevents privilege escalation
- `ReadWritePaths=/opt/cloud-vault/data /opt/cloud-vault/backups` — explicit write paths

See `deploy/systemd/INSTALL.md` for full documentation.

### Settings Page

The Settings page (`Settings` tab) displays non-sensitive configuration:

- App version
- Storage provider (Local / Qiniu)
- Auth enabled status
- Search enabled status
- AI enabled status
- Database path
- Storage root (local mode only)

Secrets (API keys, passwords) are never shown.

### Security Recommendations

1. **Production: enable authentication**
   ```toml
   [auth]
   enabled = true
   secure_cookie = true
   ```

2. **Use HTTPS**
   - Terminate TLS at a reverse proxy (nginx, Caddy, Traefik)
   - Set `AUTH_SECURE_COOKIE=true`
   - Redirect all HTTP to HTTPS

3. **Never commit secrets to Git**
   - Use `.env` file (add to `.gitignore`)
   - Inject Qiniu/AI keys via environment variables
   - Use a secrets manager in production (Vault, AWS Secrets Manager)

4. **Never log secrets**
   - Cloud Vault automatically masks sensitive fields in logs
   - Do not override this behavior

5. **Back up regularly**
   - SQLite: `cloud-vault backup` or API
   - Local storage: included in backup zip
   - Qiniu: use Qiniu's own backup tools

6. **Rotate passwords periodically**
   - Change admin password via Security page
   - All sessions automatically revoked after password change

7. **Monitor security events**
   - Check Security page for unusual login patterns
   - Revoke suspicious sessions immediately

### V0.8 Current Limitations

- **Restore recommended offline** — concurrent writes during restore are undefined
- **Qiniu mode doesn't auto-backup bucket contents** — use Qiniu's own backup tools
- **Docker/systemd are examples** — production hardening (secrets management, health checks, TLS termination) is user responsibility
- **No incremental backups** — every backup is a full snapshot
- **No scheduled backup** — use external cron/systemd timer
- **No backup encryption at rest** — zip is stored plain; encrypt if storing offsite
- **No automatic rotation of old backups** — manual cleanup required
- **Multi-user permissions and team collaboration not implemented**

### Migration from V0.7

If upgrading from V0.7:

1. Backup your database: `cp data/app.db data/app.db.backup`
2. Stop the running application
3. Run migration: `./cloud-vault` (migration 008 runs automatically)
4. Access the application — no setup required, all users and sessions preserved
5. Review new deployment checks: System tab → Doctor → select deployment checks

All existing documents, users, sessions, and data remain intact.

## V0.9: Desktop Application (Wails)

V0.9 wraps Cloud Vault in a Wails v2 desktop application with system tray, native directory picker, and desktop-optimized import experience.

### Architecture

```
cmd/desktop/main.go          → Wails entry point
internal/desktop/
├── config.go                → DesktopConfig + platform data dirs
├── server.go                → Local HTTP server (127.0.0.1:0 + token)
├── service.go               → DesktopService: runtime info, scan preview
├── bindings.go              → WailsBindings: methods exposed to frontend
├── tray.go                  → TrayManager: systray icon + menu
└── data_dir.go              → Open data directory in file manager
internal/app/app.go          → HTTPHandler() + StartBackgroundTasks()
```

The desktop app:
- Starts a local HTTP server on `127.0.0.1:0` (never `0.0.0.0`)
- Generates a random 32-byte token for authentication
- Embeds the existing React frontend via Wails AssetServer
- Reuses all existing backend routes and services unchanged
- Provides desktop-specific APIs via Wails bindings

### Data Directory

Desktop mode stores all data in the system App Data directory:

| Platform | Default Path |
|---|---|
| macOS | `~/Library/Application Support/CloudVault` |
| Windows | `%APPDATA%\CloudVault` |
| Linux | `$XDG_DATA_HOME/cloudvault` or `~/.local/share/cloudvault` |

The directory contains:
```
CloudVault/
├── app.db          ← SQLite database
├── objects/        ← Local object storage
├── backups/        ← Backup archives
├── logs/           ← Application logs
└── cache/          ← Temporary cache
```

### Desktop Configuration

Desktop-specific config in `config.toml`:

```toml
[desktop]
app_name = "Cloud Vault"
data_dir = ""                              # auto-detect by platform
close_to_tray = true
native_notifications = true
launch_at_login = false
```

Environment variables:

| Variable | Description |
|---|---|
| `DESKTOP_APP_NAME` | Application name shown in tray |
| `DESKTOP_DATA_DIR` | Override data directory path |
| `DESKTOP_CLOSE_TO_TRAY` | Close to tray instead of quit (`true`/`false`) |
| `DESKTOP_NATIVE_NOTIFICATIONS` | Enable native desktop notifications |
| `DESKTOP_LAUNCH_AT_LOGIN` | Start at login (not yet implemented) |

### System Tray

The tray icon provides quick access to common actions:

| Menu Item | Action |
|---|---|
| Show Window | Restore the main window |
| Open Data Directory | Open data folder in file manager |
| Scan Import Directory | Preview files in import directory |
| Runtime Info | Display runtime information |
| Close to Tray | Toggle close-to-tray behavior |
| Quit | Exit the application |

When close-to-tray is enabled, closing the window hides it instead of quitting.

### Desktop API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/system/runtime` | Runtime info (mode, port, data dir, platform) |
| POST | `/api/v1/import-jobs/scan-preview` | Preview scan of import directory |
| GET | `/api/v1/desktop/data-dir` | Data directory info |
| POST | `/api/v1/desktop/open-data-dir` | Open data directory in file manager |

All endpoints require the token passed via `?token=...` query parameter or `X-Auth-Token` header.

### Wails Bindings

The frontend can call these methods via Wails runtime:

| Method | Description |
|---|---|
| `GetRuntimeInfo()` | Get runtime information |
| `SelectDirectory(title, defaultPath)` | Open native directory picker |
| `ScanPreview(req)` | Preview scan of directory |
| `OpenDataDirectory()` | Open data folder in file manager |
| `ShowNotification(title, message)` | Show desktop notification |
| `StartImport(jobID, req)` | Start import with progress tracking |
| `GetConfig()` | Get desktop configuration |
| `UpdateConfig(config)` | Update desktop configuration |
| `Quit()` | Quit the application |
| `MinimizeToTray()` | Hide window to tray |
| `RestoreFromTray()` | Show window from tray |

### Building the Desktop App

```bash
# Development mode with hot reload
make desktop-dev

# Production build
make desktop-build

# Package for distribution
make desktop-package

# Clean build artifacts
make desktop-clean
```

### Running the Desktop App

```bash
# Build and run
go build -o bin/cloud-vault-desktop ./cmd/desktop
./bin/cloud-vault-desktop

# Or use go run
go run ./cmd/desktop
```

### Security

Desktop mode enforces strict security:

1. **Local-only binding** — HTTP server binds to `127.0.0.1` only, never `0.0.0.0`
2. **Random port** — Port is selected randomly from available ports
3. **Token authentication** — 32-byte random token required for all API calls
4. **Timing-safe comparison** — Token validation uses constant-time comparison
5. **No secrets in logs** — Token and sensitive config never logged

### Server Mode vs Desktop Mode

| Feature | Server Mode (`cmd/server`) | Desktop Mode (`cmd/desktop`) |
|---|---|---|
| Entry point | `cmd/server/main.go` | `cmd/desktop/main.go` |
| Bind address | Configurable (`APP_ADDR`) | Always `127.0.0.1:0` |
| Authentication | Cookie-based sessions | Token-based (random) |
| Data directory | Configurable (`DB_PATH`) | System App Data |
| UI | Browser-based | Wails native window |
| System tray | No | Yes |
| TLS | Configurable | Disabled |

### V0.9 Current Limitations

- **Launch at login** — not yet implemented
- **Auto-update** — not yet implemented
- **Multi-window** — single window only
- **Offline sync** — not yet implemented
- **Native notifications on Linux** — falls back to logging

### Migration from V0.8

Desktop mode uses a separate data directory from server mode. To migrate data:

1. Export from server mode: `cp data/app.db ~/Library/Application\ Support/CloudVault/app.db`
2. Copy objects: `cp -r data/objects ~/Library/Application\ Support/CloudVault/objects`
3. Start desktop mode — data is available immediately

### V0.9 Verification

See `docs/testing/desktop-v0.9-checklist.md` for the complete verification checklist covering:
- Build verification
- Configuration validation
- HTTP server security
- Desktop service APIs
- System tray functionality
- Wails bindings
- Platform-specific behavior
- Security requirements
- Performance requirements

## Current Limitations

- No multi-user collaboration (single admin account only)
- No semantic or vector search
- No OAuth/SSO integration
- No role-based access control (RBAC)
- No multi-tenancy support
- No two-factor authentication (2FA)
- No email verification or password reset
- No incremental or scheduled backups
- No backup encryption at rest
