# Search Guide

This guide covers full-text search capabilities in Markdown Cloud Vault.

## Overview

Cloud Vault provides powerful full-text search powered by SQLite FTS5, supporting:
- Full-text search across document content
- Search by title, tags, and metadata
- Advanced search operators
- Search result highlighting
- Search history

## Basic Search

### Via Web Interface

1. Navigate to **Search** page (or press `Ctrl/Cmd + F`)
2. Enter search query in the search box
3. Results appear instantly as you type
4. Click result to open document

### Via API

```bash
curl -X GET "http://localhost:8080/api/v1/search?q=markdown+guide" \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
{
  "results": [
    {
      "id": "abc123",
      "title": "Markdown Guide",
      "snippet": "...comprehensive <mark>markdown</mark> <mark>guide</mark> for beginners...",
      "score": 0.95,
      "tags": ["markdown", "guide"],
      "updated_at": "2026-01-15T10:00:00Z"
    }
  ],
  "total": 42,
  "query_time_ms": 15
}
```

## Search Operators

### Phrase Search

Search for exact phrases:

```
"markdown guide"
```

Matches: "This is a markdown guide"
Does not match: "guide to markdown"

### Wildcard Search

Use `*` for wildcard matching:

```
mark*
```

Matches: "markdown", "marketing", "marker"

### Boolean Operators

**AND** (default):
```
markdown guide
```

Matches documents containing both "markdown" and "guide".

**OR**:
```
markdown OR guide
```

Matches documents containing either "markdown" or "guide".

**NOT**:
```
markdown -guide
```

Matches documents containing "markdown" but not "guide".

### Field-Specific Search

**Search in title**:
```
title:markdown
```

**Search in tags**:
```
tag:guide
```

**Search in content**:
```
content:markdown
```

**Search in author**:
```
author:"John Doe"
```

### Combined Search

Combine multiple operators:

```
title:guide tag:markdown -draft
```

Matches: Documents with "guide" in title, tagged "markdown", not tagged "draft".

## Advanced Search

### Date Range Search

Search documents by date:

```
date:2026-01-01..2026-01-31
```

Matches: Documents created in January 2026.

**Relative dates**:
```
date:>now-7d    # Last 7 days
date:>now-1M    # Last month
date:>now-1y    # Last year
```

### Size Range Search

Search by document size:

```
size:>1KB       # Larger than 1 KB
size:<10KB      # Smaller than 10 KB
size:1KB..10KB  # Between 1 KB and 10 KB
```

### Tag Count Search

Search by number of tags:

```
tags:>5         # More than 5 tags
tags:<3         # Fewer than 3 tags
```

### Collection Search

Search within specific collection:

```
collection:project-alpha
```

Matches: Documents in the "project-alpha" collection.

### Status Search

Search by document status:

```
status:published
status:draft
status:archived
```

## Search Result Features

### Snippet Highlighting

Search results show context snippets with highlighted matches:

```
...comprehensive <mark>markdown</mark> <mark>guide</mark> for beginners...
```

**Configuration**:
```toml
[search]
snippet_length = 200
highlight_pre_tag = "<mark>"
highlight_post_tag = "</mark>"
```

### Result Ranking

Results are ranked by relevance:
1. **Title matches** - Highest priority
2. **Tag matches** - High priority
3. **Content matches** - Normal priority
4. **Recency** - Newer documents rank higher

**Boost specific fields**:
```
title:markdown^3 guide
```

Boosts title matches by 3x.

### Faceted Search

Filter results by facets:

**Via Web Interface**:
1. Perform search
2. Use sidebar filters:
   - Tags
   - Collections
   - Date range
   - Status

**Via API**:
```bash
curl -X GET "http://localhost:8080/api/v1/search?q=markdown&tags=guide,tutorial&status=published" \
  -H "Authorization: Bearer $TOKEN"
```

## Search Indexing

### Automatic Indexing

Documents are indexed automatically:

```toml
[search]
index_on_save = true
```

**When indexing occurs**:
- Document created
- Document updated
- Document imported
- Tags changed

### Manual Reindexing

Rebuild entire search index:

**Via Web Interface**:
1. Navigate to **Search** page
2. Click **Index** tab
3. Click **Rebuild Index**

**Via CLI**:
```bash
./cloud-vault search reindex
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/search/reindex \
  -H "Authorization: Bearer $TOKEN"
```

### Index Status

Check index status:

**Via Web Interface**:
1. Navigate to **Search** page
2. Click **Index** tab
3. View statistics

**Via CLI**:
```bash
./cloud-vault search status
```

Output:
```
Search Index Status
===================
Indexed documents: 1234
Pending indexing: 0
Failed indexing: 0
Index size: 45.2 MB
Last indexed: 2026-01-15 10:00:00
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/search/status \
  -H "Authorization: Bearer $TOKEN"
```

### Batch Indexing

Index documents in batches:

```toml
[search]
batch_size = 100
index_interval = 5  # seconds
```

## Search History

### View History

**Via Web Interface**:
1. Navigate to **Search** page
2. Click search box
3. Recent searches appear in dropdown

**Via API**:
```bash
curl http://localhost:8080/api/v1/search/history \
  -H "Authorization: Bearer $TOKEN"
```

### Clear History

**Via Web Interface**:
1. Navigate to **Search** page
2. Click search box
3. Click **Clear History**

**Via API**:
```bash
curl -X DELETE http://localhost:8080/api/v1/search/history \
  -H "Authorization: Bearer $TOKEN"
```

### History Configuration

```toml
[search]
history_enabled = true
history_limit = 100
```

## Saved Searches

Save frequently used searches:

**Via Web Interface**:
1. Perform search
2. Click **Save Search**
3. Enter name
4. Click **Save**

**Access saved searches**:
1. Navigate to **Search** page
2. Click **Saved Searches**
3. Click saved search to run

**Via API**:
```bash
# Save search
curl -X POST http://localhost:8080/api/v1/search/saved \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Recent Markdown Guides",
    "query": "tag:markdown date:>now-30d"
  }'

# List saved searches
curl http://localhost:8080/api/v1/search/saved \
  -H "Authorization: Bearer $TOKEN"
```

## Search Shortcuts

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + F` | Focus search box |
| `Enter` | Open first result |
| `↑` / `↓` | Navigate results |
| `Esc` | Clear search |

### Quick Search

Search from anywhere:
1. Press `Ctrl/Cmd + F`
2. Type query
3. Press `Enter` to open result

## Search Performance

### Index Optimization

Optimize search index:

```bash
./cloud-vault search optimize
```

**Automatic optimization**:
```toml
[search]
auto_optimize = true
optimize_interval = 86400  # Daily
```

### Query Performance

Monitor query performance:

```bash
curl http://localhost:8080/api/v1/search/status \
  -H "Authorization: Bearer $TOKEN"
```

Response includes:
```json
{
  "avg_query_time_ms": 12,
  "slow_queries": 3,
  "cache_hit_rate": 0.85
}
```

### Caching

Enable search result caching:

```toml
[search]
cache_enabled = true
cache_ttl = 300  # 5 minutes
cache_max_size = 1000  # Max cached queries
```

## Advanced Search Examples

### Find Duplicate Content

```bash
./cloud-vault search duplicates
```

Or via search:
```
content:"specific phrase"
```

### Find Orphaned Documents

Documents not in any collection:
```
-collection:*
```

### Find Untagged Documents

```
-tag:*
```

### Find Large Documents

```
size:>1MB
```

### Find Recent Changes

```
date:>now-7d
```

### Find Draft Documents

```
status:draft
```

### Find Documents by Author

```
author:"John Doe"
```

### Complex Query

```
title:guide tag:markdown -draft date:>now-30d size:>1KB
```

## Search API

### Search Endpoint

```bash
GET /api/v1/search?q=query&page=1&limit=20
```

**Parameters**:
- `q` - Search query (required)
- `page` - Page number (default: 1)
- `limit` - Results per page (default: 20, max: 100)
- `sort` - Sort by: relevance, date, title (default: relevance)
- `order` - Sort order: asc, desc (default: desc)

**Response**:
```json
{
  "results": [...],
  "total": 42,
  "page": 1,
  "limit": 20,
  "query_time_ms": 15
}
```

### Suggest Endpoint

Get search suggestions:

```bash
GET /api/v1/search/suggest?q=mark
```

**Response**:
```json
{
  "suggestions": [
    "markdown",
    "marketing",
    "marker"
  ]
}
```

## Troubleshooting Search

### No Results Found

**Symptoms**: Search returns no results

**Solutions**:
1. Check spelling
2. Try broader search terms
3. Rebuild index: `./cloud-vault search reindex`
4. Verify document exists

### Slow Search

**Symptoms**: Search takes too long

**Solutions**:
```toml
[search]
cache_enabled = true
index_on_save = true
```

```bash
# Optimize index
./cloud-vault search optimize
```

### Index Out of Sync

**Symptoms**: Search results don't match documents

**Solution**:
```bash
# Rebuild index
./cloud-vault search reindex

# Verify index
./cloud-vault search status
```

### Highlighting Not Working

**Symptoms**: Search results don't show highlights

**Solutions**:
```toml
[search]
highlight_enabled = true
highlight_pre_tag = "<mark>"
highlight_post_tag = "</mark>"
```

## Next Steps

- [Organization Guide](./organize.md) - Organize search results
- [Collections Guide](./collections.md) - Group related documents
- [AI Features](./ai.md) - AI-powered search enhancements

---

**Last updated**: 2026-05-30
