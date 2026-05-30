# Organization Guide

This guide covers organizing your Markdown documents using collections, tags, and smart organization features.

## Overview

Cloud Vault provides multiple ways to organize documents:
- **Collections**: Named groups of related documents
- **Tags**: Flexible labels for categorization
- **Topics**: AI-powered topic clustering
- **Review Queue**: Process unorganized documents
- **Duplicate Detection**: Find and manage duplicates

## Collections

### Create Collection

**Via Web Interface**:
1. Navigate to **Collections** page
2. Click **New Collection**
3. Enter collection name and description
4. Click **Create**

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/collections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Project Alpha",
    "description": "All documents related to Project Alpha"
  }'
```

### Add Documents to Collection

**Via Web Interface**:
1. Open collection
2. Click **Add Documents**
3. Search and select documents
4. Click **Add**

Or drag-and-drop documents from Vault to Collection.

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/collections/123/documents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_ids": ["abc123", "def456"]
  }'
```

### Remove from Collection

**Via Web Interface**:
1. Open collection
2. Select document(s)
3. Click **Remove**

**Via API**:
```bash
curl -X DELETE http://localhost:8080/api/v1/collections/123/documents/abc123 \
  -H "Authorization: Bearer $TOKEN"
```

### Nested Collections

Organize collections hierarchically:

```
Projects/
├── Project Alpha/
│   ├── Specifications
│   └── Meeting Notes
└── Project Beta/
    ├── Research
    └── Development
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/collections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Specifications",
    "parent_id": "project-alpha-id"
  }'
```

## Tags

### Add Tags

**Via Web Interface**:
1. Open document
2. Click **Tags** section
3. Type tag name and press Enter
4. Or select from suggested tags

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/documents/abc123/tags \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["markdown", "guide", "tutorial"]
  }'
```

### Remove Tags

**Via Web Interface**:
1. Open document
2. Click **Tags** section
3. Click × next to tag

**Via API**:
```bash
curl -X DELETE http://localhost:8080/api/v1/documents/abc123/tags/markdown \
  -H "Authorization: Bearer $TOKEN"
```

### Tag Management

**View all tags**:
1. Navigate to **Tags** page
2. See tag cloud and usage statistics

**Rename tag**:
```bash
curl -X PUT http://localhost:8080/api/v1/tags/old-name \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "new-name"
  }'
```

**Merge tags**:
```bash
curl -X POST http://localhost:8080/api/v1/tags/merge \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "markdown-docs",
    "target": "markdown"
  }'
```

### Auto-Tagging

Enable automatic tag generation:

```toml
[organize]
auto_tagging = true
auto_tag_sources = ["filename", "path", "content"]
```

**From filename**:
```
project-alpha-spec.md → tags: ["project", "alpha", "spec"]
```

**From content** (requires AI):
```toml
[ai]
enabled = true
auto_tag_content = true
```

## Topics

### Topic Detection

AI-powered topic clustering:

**Via Web Interface**:
1. Navigate to **Organize** page
2. Click **Detect Topics**
3. Review detected topics
4. Apply or refine

**Via CLI**:
```bash
./cloud-vault organize detect-topics
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/organize/detect-topics \
  -H "Authorization: Bearer $TOKEN"
```

### Topic Management

**View topics**:
1. Navigate to **Topics** page
2. See topic clusters and documents

**Rename topic**:
```bash
curl -X PUT http://localhost:8080/api/v1/topics/123 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Machine Learning"
  }'
```

**Merge topics**:
```bash
curl -X POST http://localhost:8080/api/v1/topics/merge \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "ai-research",
    "target_id": "machine-learning"
  }'
```

## Duplicate Detection

### Find Duplicates

**Via Web Interface**:
1. Navigate to **Organize** page
2. Click **Detect Duplicates**
3. Review duplicate groups
4. Choose action: merge, delete, or ignore

**Via CLI**:
```bash
./cloud-vault organize detect-duplicates
```

Output:
```
Duplicate Detection Results
============================
Found 3 duplicate groups:

Group 1 (2 documents):
  - document1.md (ID: abc123)
  - document1-copy.md (ID: def456)
  Similarity: 100%

Group 2 (3 documents):
  - notes.md (ID: ghi789)
  - notes-backup.md (ID: jkl012)
  - notes-old.md (ID: mno345)
  Similarity: 95%

Actions: [m]erge, [d]elete, [i]gnore, [q]uit
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/organize/detect-duplicates \
  -H "Authorization: Bearer $TOKEN"
```

### Resolve Duplicates

**Merge duplicates**:
```bash
curl -X POST http://localhost:8080/api/v1/organize/merge-duplicates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "group_id": "group-1",
    "keep": "abc123",
    "delete": ["def456"]
  }'
```

**Configuration**:
```toml
[organize]
duplicate_threshold = 0.95  # 95% similarity
duplicate_action = "skip"   # "skip", "merge", "delete"
```

## Review Queue

### Add to Review Queue

Documents are automatically added to review queue when:
- Imported without tags
- No collection assigned
- Missing metadata

**Via Web Interface**:
1. Navigate to **Review** page
2. See documents needing review
3. Add tags, collections, or metadata
4. Mark as reviewed

**Via API**:
```bash
# Get review queue
curl http://localhost:8080/api/v1/organize/review-queue \
  -H "Authorization: Bearer $TOKEN"

# Mark as reviewed
curl -X POST http://localhost:8080/api/v1/organize/review/abc123/complete \
  -H "Authorization: Bearer $TOKEN"
```

### Batch Review

Process multiple documents:

**Via Web Interface**:
1. Navigate to **Review** page
2. Select multiple documents
3. Click **Batch Actions**
4. Choose: Add tags, Add to collection, Mark reviewed

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/organize/review/batch \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_ids": ["abc123", "def456"],
    "action": "add_tags",
    "tags": ["reviewed"]
  }'
```

## Smart Organization

### Auto-Categorization

Automatically categorize documents:

```toml
[organize]
auto_categorize = true
categorize_rules = [
  { pattern: "spec|specification", collection: "Specifications" },
  { pattern: "meeting|notes", collection: "Meeting Notes" },
  { pattern: "research|study", collection: "Research" }
]
```

### Tag Suggestions

Get AI-powered tag suggestions:

**Via Web Interface**:
1. Open document
2. Click **Tags** section
3. See suggested tags
4. Click to add

**Via API**:
```bash
curl http://localhost:8080/api/v1/documents/abc123/tag-suggestions \
  -H "Authorization: Bearer $TOKEN"
```

### Quality Scoring

Score document quality:

**Via Web Interface**:
1. Navigate to **Organize** page
2. Click **Score Quality**
3. View quality scores

**Via CLI**:
```bash
./cloud-vault organize score-quality
```

**Scoring criteria**:
- Has title
- Has tags
- In collection
- Has metadata
- Content length
- Readability

## Bulk Operations

### Bulk Tag

Add tags to multiple documents:

**Via Web Interface**:
1. Navigate to **Vault** page
2. Select multiple documents
3. Click **Bulk Actions** → **Add Tags**
4. Enter tags
5. Click **Apply**

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/organize/bulk-tags \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_ids": ["abc123", "def456", "ghi789"],
    "tags": ["project-alpha", "2026"]
  }'
```

### Bulk Move

Move documents to collection:

**Via Web Interface**:
1. Select multiple documents
2. Click **Bulk Actions** → **Move to Collection**
3. Choose collection
4. Click **Move**

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/organize/bulk-move \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_ids": ["abc123", "def456"],
    "collection_id": "project-alpha"
  }'
```

### Bulk Delete

Delete multiple documents:

**Via Web Interface**:
1. Select multiple documents
2. Click **Bulk Actions** → **Delete**
3. Confirm deletion

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/organize/bulk-delete \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_ids": ["abc123", "def456"]
  }'
```

## Organization Best Practices

### 1. Use Consistent Tagging

**Good**:
```
project-alpha, specification, 2026
```

**Bad**:
```
Project Alpha, spec, 2026, project_alpha
```

### 2. Hierarchical Collections

```
Projects/
├── Active/
│   ├── Project Alpha
│   └── Project Beta
└── Completed/
    ├── Project Gamma
    └── Project Delta
```

### 3. Tag Naming Conventions

**Use lowercase**:
```
markdown, guide, tutorial
```

**Use hyphens for multi-word**:
```
machine-learning, deep-learning
```

**Use prefixes for categories**:
```
status:active, status:completed
type:spec, type:notes
```

### 4. Regular Maintenance

**Weekly**:
- Process review queue
- Merge duplicates
- Add tags to new documents

**Monthly**:
- Review tag usage
- Merge unused tags
- Archive completed projects

**Quarterly**:
- Reorganize collections
- Update topic clusters
- Clean up old documents

## Organization Statistics

### View Statistics

**Via Web Interface**:
1. Navigate to **Organize** page
2. View statistics panel

**Via API**:
```bash
curl http://localhost:8080/api/v1/organize/stats \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
{
  "total_documents": 1234,
  "collections": 45,
  "tags": 234,
  "topics": 28,
  "duplicates": 12,
  "review_queue": 56,
  "untagged": 23,
  "uncollected": 45
}
```

## Next Steps

- [Search Guide](./search.md) - Search organized documents
- [AI Features](./ai.md) - AI-powered organization
- [Collections Guide](./collections.md) - Advanced collection management

---

**Last updated**: 2026-05-30
