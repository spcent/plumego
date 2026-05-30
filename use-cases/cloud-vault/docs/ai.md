# AI Features Guide

This guide covers AI-powered features in Markdown Cloud Vault.

## Overview

Cloud Vault integrates AI to enhance your document management:
- **Summarization**: Generate concise summaries
- **Q&A**: Ask questions about your documents
- **Prompt Extraction**: Extract prompts from documents
- **Auto-Tagging**: Suggest tags based on content
- **Topic Detection**: Identify document topics

## Configuration

### Enable AI Features

```toml
[ai]
enabled = true
provider = "openai"
model = "gpt-4"
temperature = 0.7
max_tokens = 4096
```

### AI Providers

**OpenAI**:
```toml
[ai]
provider = "openai"

[ai.openai]
api_key = "${OPENAI_API_KEY}"
base_url = "https://api.openai.com/v1"
```

**Azure OpenAI**:
```toml
[ai]
provider = "azure"

[ai.azure]
api_key = "${AZURE_API_KEY}"
endpoint = "https://your-resource.openai.azure.com"
deployment = "gpt-4"
```

**Local LLM (Ollama)**:
```toml
[ai]
provider = "local"

[ai.local]
endpoint = "http://localhost:11434"
model = "llama2"
```

## Summarization

### Summarize Document

**Via Web Interface**:
1. Open document
2. Click **AI** → **Summarize**
3. View generated summary
4. Click **Save** to add as metadata

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/ai/summarize \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_id": "abc123",
    "max_length": 200
  }'
```

### Batch Summarization

Summarize multiple documents:

**Via CLI**:
```bash
./cloud-vault ai summarize \
  --collection "Project Alpha" \
  --max-length 200
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/ai/summarize/batch \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_ids": ["abc123", "def456"],
    "max_length": 200
  }'
```

## Q&A (Question Answering)

### Ask Questions

**Via Web Interface**:
1. Open document or collection
2. Click **AI** → **Ask Question**
3. Type your question
4. View AI-generated answer

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/ai/qa \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_id": "abc123",
    "question": "What are the main points?"
  }'
```

### Multi-Document Q&A

Ask questions across multiple documents:

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/ai/qa/multi \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_ids": ["abc123", "def456", "ghi789"],
    "question": "Compare the approaches in these documents"
  }'
```

## Prompt Extraction

### Extract Prompts

Extract prompts and questions from documents:

**Via Web Interface**:
1. Open document
2. Click **AI** → **Extract Prompts**
3. Review extracted prompts
4. Save to prompt library

**Via CLI**:
```bash
./cloud-vault ai extract-prompts \
  --document abc123
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/ai/extract-prompts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "document_id": "abc123"
  }'
```

### Prompt Library

Manage extracted prompts:

**Via Web Interface**:
1. Navigate to **Prompts** page
2. View all prompts
3. Organize by category
4. Use in AI tasks

**Via API**:
```bash
# List prompts
curl http://localhost:8080/api/v1/prompts \
  -H "Authorization: Bearer $TOKEN"

# Create prompt
curl -X POST http://localhost:8080/api/v1/prompts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Summarize Meeting Notes",
    "content": "Summarize these meeting notes in 3 bullet points",
    "category": "meetings"
  }'
```

## Auto-Tagging

### Enable Auto-Tagging

```toml
[ai]
enabled = true
auto_tag = true
auto_tag_threshold = 0.7
```

### Manual Tag Suggestion

**Via Web Interface**:
1. Open document
2. Click **Tags** → **Suggest Tags**
3. Review suggestions
4. Click to add

**Via API**:
```bash
curl http://localhost:8080/api/v1/documents/abc123/tag-suggestions \
  -H "Authorization: Bearer $TOKEN"
```

## Topic Detection

### Detect Topics

**Via Web Interface**:
1. Navigate to **Organize** page
2. Click **Detect Topics**
3. Review detected topics
4. Apply to documents

**Via CLI**:
```bash
./cloud-vault ai detect-topics \
  --min-documents 5
```

## AI Task Queue

### View AI Tasks

**Via Web Interface**:
1. Navigate to **AI Tasks** page
2. View active and completed tasks
3. Check progress and results

**Via API**:
```bash
curl http://localhost:8080/api/v1/ai/tasks \
  -H "Authorization: Bearer $TOKEN"
```

### Cancel AI Task

**Via Web Interface**:
1. Navigate to **AI Tasks** page
2. Find active task
3. Click **Cancel**

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/ai/tasks/123/cancel \
  -H "Authorization: Bearer $TOKEN"
```

## Privacy and Security

### Data Sent to AI

**What is sent**:
- Document content (for summarization, Q&A)
- Document metadata (for tagging)
- User prompts

**What is NOT sent**:
- API keys
- Authentication tokens
- Database content
- File paths

### Local Processing

For maximum privacy, use local LLM:

```toml
[ai]
provider = "local"

[ai.local]
endpoint = "http://localhost:11434"
model = "llama2"
```

**Benefits**:
- No data leaves your machine
- No API costs
- Full control over model

**Limitations**:
- Requires GPU for good performance
- Smaller models may be less accurate

## Cost Management

### Monitor API Usage

```bash
curl http://localhost:8080/api/v1/ai/stats \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
{
  "requests_today": 45,
  "tokens_today": 125000,
  "estimated_cost_today": 1.25,
  "requests_this_month": 1234,
  "tokens_this_month": 3456000,
  "estimated_cost_this_month": 34.56
}
```

### Rate Limiting

```toml
[ai]
rate_limit = 60  # Requests per minute
daily_limit = 1000  # Max requests per day
```

### Cost Optimization

**Use smaller models for simple tasks**:
```toml
[ai]
model = "gpt-4"  # For complex tasks

[ai.tasks]
summarize_model = "gpt-3.5-turbo"  # For summarization
tag_model = "gpt-3.5-turbo"  # For tagging
```

**Cache results**:
```toml
[ai]
cache_enabled = true
cache_ttl = 86400  # 24 hours
```

## AI Best Practices

### 1. Review AI Output

Always review AI-generated content:
- Summaries may miss key points
- Q&A answers may be inaccurate
- Tags may not match your taxonomy

### 2. Provide Context

Better context = better results:
```
Question: "What are the main points?"
Better: "What are the 3 main technical decisions in this architecture document?"
```

### 3. Use Prompts Library

Save effective prompts for reuse:
- Summarization templates
- Q&A patterns
- Extraction rules

### 4. Batch Processing

Process multiple documents efficiently:
```bash
./cloud-vault ai summarize \
  --collection "Meeting Notes" \
  --batch-size 10
```

### 5. Monitor Costs

Set up alerts:
```toml
[ai]
cost_alert_threshold = 50.00  # Alert at $50/month
cost_alert_email = "admin@example.com"
```

## Troubleshooting AI

### API Errors

**Error**: `401 Unauthorized`
- Check API key is valid
- Verify API key has required permissions

**Error**: `429 Too Many Requests`
- Reduce request rate
- Increase rate limit
- Use caching

**Error**: `500 Internal Server Error`
- Check AI provider status
- Retry request
- Switch to backup provider

### Slow Responses

**Solutions**:
```toml
[ai]
timeout = 120  # Increase timeout
max_tokens = 2048  # Reduce response length
```

### Poor Quality Results

**Solutions**:
1. Use better model (gpt-4 instead of gpt-3.5)
2. Provide more context
3. Adjust temperature (lower = more focused)
4. Break into smaller tasks

## Next Steps

- [Organization Guide](./organize.md) - AI-powered organization
- [Search Guide](./search.md) - Enhanced search with AI
- [Security](./security.md) - AI security considerations

---

**Last updated**: 2026-05-30
