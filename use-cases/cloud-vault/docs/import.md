# Import Guide

This guide covers importing Markdown documents into Cloud Vault from various sources.

## Overview

Cloud Vault supports importing:
- Markdown files (`.md`, `.markdown`)
- Text files (`.txt`)
- Directory structures (recursive import)
- ZIP archives (bulk import)
- Single files (drag-and-drop)

## Import Methods

### 1. Directory Import (Recommended)

Import an entire server-configured directory of Markdown files. The web and API
flows do not accept arbitrary filesystem paths; they list directories under
`IMPORTER_SAFE_ROOT` and submit the selected opaque `source_id`.

**Via Web Interface**:
1. Navigate to **Import** page
2. Select one of the available source directories
3. Enter an optional job name
4. Click **Start Import**

**Via CLI**:
```bash
./cloud-vault import \
  --source /path/to/markdown/files \
  --recursive \
  --auto-tag \
  --extract-metadata
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/imports/sources \
  -H "Authorization: Bearer $TOKEN"

curl -X POST http://localhost:8080/api/v1/imports \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Knowledge base import",
    "source_id": "root"
  }'
```

### 2. Single File Import

Import individual Markdown files:

**Via Web Interface**:
1. Navigate to **Import** page
2. Drag-and-drop file(s) onto the import area
3. Or click **Select Files** and choose files
4. Review import preview
5. Click **Import**

**Via CLI**:
```bash
./cloud-vault import \
  --source /path/to/document.md \
  --auto-tag
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/import/file \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@document.md"
```

### 3. Archive Import

Import from ZIP archives:

**Via Web Interface**:
1. Navigate to **Import** page
2. Click **Import Archive**
3. Select ZIP file
4. Configure options:
   - **Preserve structure**: Keep directory hierarchy
   - **Auto-tag**: Generate tags from paths
5. Click **Start Import**

**Via CLI**:
```bash
./cloud-vault import \
  --source /path/to/archive.zip \
  --preserve-structure \
  --auto-tag
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/import/archive \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@archive.zip" \
  -F "preserve_structure=true"
```

## Import Options

### Recursive Import

Import subdirectories:

```toml
[import]
recursive = true
```

**Example**:
```
source/
├── docs/
│   ├── guide1.md
│   └── guide2.md
└── notes/
    ├── note1.md
    └── note2.md
```

Result: All 4 files imported with path-based tags.

### Auto-Tagging

Generate tags automatically:

**From filename**:
```
my-document.md → tags: ["my", "document"]
project-alpha-notes.md → tags: ["project", "alpha", "notes"]
```

**From directory path**:
```
/projects/alpha/docs/spec.md → tags: ["projects", "alpha", "docs"]
```

**From frontmatter**:
```markdown
---
tags: [markdown, guide, tutorial]
---
```

Configuration:
```toml
[import]
auto_tag = true
tag_separator = "-"  # Split filenames on this character
tag_from_path = true  # Include path components as tags
```

### Metadata Extraction

Extract frontmatter metadata:

```markdown
---
title: My Document
author: John Doe
date: 2026-01-15
tags: [markdown, guide]
custom_field: custom_value
---

# Document Content
```

Extracted fields:
- `title` → Document title
- `author` → Author metadata
- `date` → Created date
- `tags` → Document tags
- `custom_field` → Custom metadata

Configuration:
```toml
[import]
extract_metadata = true
metadata_fields = ["title", "author", "date", "tags"]
```

### Duplicate Detection

Detect and handle duplicate files:

**Detection methods**:
1. **Content hash**: SHA-256 hash of file content
2. **Filename**: Exact filename match
3. **Title**: Document title match

**Actions**:
- `skip`: Skip duplicate (default)
- `overwrite`: Replace existing document
- `version`: Create new version
- `rename`: Import with new name

Configuration:
```toml
[import]
duplicate_detection = true
duplicate_method = "content_hash"  # "content_hash", "filename", "title"
duplicate_action = "skip"  # "skip", "overwrite", "version", "rename"
```

**Example**:
```bash
./cloud-vault import \
  --source /path/to/files \
  --duplicate-detection \
  --duplicate-action skip
```

Result:
```
Importing files...
✓ Imported: document1.md
⊘ Skipped (duplicate): document2.md
✓ Imported: document3.md
```

## Import Preview

Preview import before executing:

**Via Web Interface**:
1. Select files/directory
2. Review preview showing:
   - Total files
   - Total size
   - Estimated tags
   - Duplicate count
3. Click **Import** or **Cancel**

**Via CLI**:
```bash
./cloud-vault import \
  --source /path/to/files \
  --dry-run
```

Output:
```
Import Preview (dry run)
========================
Source: /path/to/files
Files: 42
Total size: 15.3 MB
Duplicates: 3
Estimated tags: 28

Files to import:
  ✓ document1.md (2.1 KB)
  ✓ document2.md (3.5 KB)
  ⊘ document3.md (duplicate)
  ...

Run without --dry-run to import.
```

## Batch Import

Import large numbers of files efficiently:

**Via CLI**:
```bash
./cloud-vault import \
  --source /path/to/files \
  --batch-size 100 \
  --workers 4
```

**Configuration**:
```toml
[import]
batch_size = 100
workers = 4
progress_interval = 5  # Progress update interval (seconds)
```

**Progress tracking**:
```
Importing files... [=====>    ] 45% (450/1000)
  Imported: 447
  Skipped: 3
  Failed: 0
  ETA: 2m 30s
```

## Import Monitoring

### View Import Jobs

**Via Web Interface**:
1. Navigate to **Import** page
2. View active and completed jobs
3. Click job to see details

**Via API**:
```bash
# List all import jobs
curl http://localhost:8080/api/v1/imports \
  -H "Authorization: Bearer $TOKEN"

# Get specific job
curl http://localhost:8080/api/v1/imports/job_123 \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
{
  "id": "job_123",
  "status": "running",
  "source_path": ".",
  "total_count": 1000,
  "processed_count": 453,
  "success_count": 450,
  "skipped_count": 3,
  "failed_count": 0,
  "started_at": "2026-01-15T10:00:00Z",
  "created_at": "2026-01-15T09:59:50Z",
  "updated_at": "2026-01-15T10:00:10Z"
}
```

### Cancel Import

**Via Web Interface**:
1. Navigate to **Import** page
2. Find active job
3. Click **Cancel**

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/import/jobs/123/cancel \
  -H "Authorization: Bearer $TOKEN"
```

### View Import Logs

**Via Web Interface**:
1. Navigate to **Import** page
2. Click job to expand
3. View detailed logs

**Via CLI**:
```bash
./cloud-vault import logs --job-id 123
```

Output:
```
[2026-01-15 10:00:00] Starting import from /path/to/files
[2026-01-15 10:00:01] Found 1000 files to import
[2026-01-15 10:00:02] Importing document1.md...
[2026-01-15 10:00:02] ✓ Imported document1.md (ID: abc123)
[2026-01-15 10:00:03] Importing document2.md...
[2026-01-15 10:00:03] ⊘ Skipped document2.md (duplicate)
...
```

## Post-Import Processing

### Automatic Indexing

Documents are automatically indexed after import:

```toml
[search]
index_on_save = true
```

**Manual reindex**:
```bash
./cloud-vault search reindex
```

### Automatic Tagging

Tags are generated based on:
- Filename
- Directory path
- Frontmatter
- Content analysis (if AI enabled)

### Collection Assignment

Assign documents to collections:

**Via Web Interface**:
1. Navigate to **Import** page
2. Select completed import job
3. Click **Add to Collection**
4. Choose or create collection

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/collections/456/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_ids": ["abc123", "def456"]
  }'
```

## Import from External Sources

### Obsidian Vault

Import from Obsidian:

```bash
./cloud-vault import \
  --source /path/to/obsidian/vault \
  --recursive \
  --auto-tag \
  --preserve-wikilinks
```

**Notes**:
- Wiki-links (`[[link]]`) are preserved
- Obsidian metadata is extracted
- Attachments are imported

### Notion Export

Import from Notion:

1. Export from Notion as Markdown/CSV
2. Extract ZIP file
3. Import directory:

```bash
./cloud-vault import \
  --source /path/to/notion-export \
  --recursive \
  --auto-tag
```

### Joplin Export

Import from Joplin:

1. Export from Joplin as JEX (Joplin Export)
2. Extract JEX file (it's a ZIP)
3. Import Markdown files:

```bash
./cloud-vault import \
  --source /path/to/joplin-export \
  --recursive \
  --extract-metadata
```

### Roam Research

Import from Roam:

1. Export from Roam as JSON
2. Convert to Markdown (use external tool)
3. Import converted files:

```bash
./cloud-vault import \
  --source /path/to/roam-export \
  --recursive
```

## Advanced Import Features

### Content Transformation

Transform content during import:

**Configuration**:
```toml
[import.transform]
strip_html = true
convert_links = true
normalize_whitespace = true
fix_encoding = true
```

**Custom transformations** (via plugin):
```javascript
// transform.js
module.exports = function(content, metadata) {
  // Custom transformation logic
  content = content.replace(/TODO:/g, '☐');
  return content;
};
```

### Selective Import

Import only specific files:

**By extension**:
```bash
./cloud-vault import \
  --source /path/to/files \
  --include "*.md" \
  --exclude "draft-*.md"
```

**By size**:
```bash
./cloud-vault import \
  --source /path/to/files \
  --max-size 1MB
```

**By date**:
```bash
./cloud-vault import \
  --source /path/to/files \
  --newer-than 2026-01-01
```

### Incremental Import

Import only new/changed files:

```bash
./cloud-vault import \
  --source /path/to/files \
  --incremental \
  --since 2026-01-01
```

**How it works**:
1. Checks last import timestamp
2. Only imports files modified since then
3. Updates existing documents if changed

## Import Best Practices

### 1. Organize Before Import

**Recommended structure**:
```
documents/
├── projects/
│   ├── project-alpha/
│   │   ├── spec.md
│   │   └── notes.md
│   └── project-beta/
├── references/
│   ├── books/
│   └── articles/
└── personal/
    ├── journal/
    └── ideas/
```

### 2. Use Frontmatter

Add metadata to documents:

```markdown
---
title: Project Alpha Specification
author: John Doe
date: 2026-01-15
tags: [project, alpha, specification]
status: draft
---

# Project Alpha

Content here...
```

### 3. Preview Before Import

Always preview large imports:

```bash
./cloud-vault import \
  --source /path/to/files \
  --dry-run
```

### 4. Enable Duplicate Detection

Avoid importing duplicates:

```toml
[import]
duplicate_detection = true
duplicate_action = "skip"
```

### 5. Monitor Import Progress

Watch import logs:

```bash
./cloud-vault import logs --follow
```

### 6. Backup Before Bulk Import

Create backup before large imports:

```bash
curl -X POST http://localhost:8080/api/v1/system/backup \
  -H "Cookie: session=<your-session-cookie>"
```

## Troubleshooting Import

### Import Fails

**Error**: `Import failed: permission denied`

**Solution**:
```bash
# Check file permissions
ls -l /path/to/files

# Fix permissions
chmod -R 644 /path/to/files
chown -R cloud-vault:cloud-vault /path/to/files
```

### Slow Import

**Symptoms**: Import takes too long

**Solutions**:
```toml
[import]
batch_size = 200  # Increase batch size
workers = 8       # Increase parallel workers
```

```bash
# Use faster import mode
./cloud-vault import \
  --source /path/to/files \
  --batch-size 200 \
  --workers 8 \
  --skip-indexing  # Index later
```

### Out of Memory

**Error**: `Out of memory during import`

**Solutions**:
```toml
[import]
batch_size = 50  # Reduce batch size
max_file_size = 5242880  # Limit file size (5 MB)
```

```bash
# Import in smaller chunks
./cloud-vault import \
  --source /path/to/files \
  --batch-size 50
```

### Encoding Issues

**Error**: `Invalid UTF-8 encoding`

**Solution**:
```toml
[import.transform]
fix_encoding = true
fallback_encoding = "latin1"
```

### Missing Files

**Symptoms**: Some files not imported

**Solutions**:
```bash
# Check file extensions
ls /path/to/files/*.md

# Verify file size
find /path/to/files -name "*.md" -size +10M

# Check import logs
./cloud-vault import logs --job-id 123 | grep "skipped"
```

## Next Steps

- [Search Guide](./search.md) - Search imported documents
- [Organization Guide](./organize.md) - Organize your vault
- [Backup & Restore](./backup-restore.md) - Backup your data

---

**Last updated**: 2026-05-30
