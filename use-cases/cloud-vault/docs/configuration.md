# Configuration Guide

This guide covers all configuration options for Markdown Cloud Vault.

## Configuration Files

Cloud Vault uses a layered configuration system:

1. **Default values** - Built-in defaults
2. **config.toml** - Main configuration file
3. **Environment variables** - Override config file values
4. **Command-line flags** - Highest priority

## Configuration File Location

**Desktop App**:
- Windows: `%APPDATA%\Cloud Vault\config.toml`
- macOS: `~/Library/Application Support/Cloud Vault/config.toml`
- Linux: `~/.config/cloud-vault/config.toml`

**Server**:
- Default: `./config.toml` (current directory)
- Custom: `./cloud-vault --config /path/to/config.toml`

## Configuration Sections

### Server Configuration

```toml
[server]
addr = ":8080"                    # Listen address
host = "0.0.0.0"                  # Bind host
port = 8080                       # Port number
read_timeout = 30                 # Read timeout (seconds)
write_timeout = 30                # Write timeout (seconds)
idle_timeout = 120                # Idle timeout (seconds)
max_header_bytes = 1048576        # Max header size (1 MB)
trusted_proxies = []              # Trusted proxy IPs
```

**Environment variables**:
```bash
SERVER_ADDR=":8080"
SERVER_HOST="0.0.0.0"
SERVER_PORT=8080
```

### Database Configuration

```toml
[database]
path = "./data/app.db"           # Database file path
max_open_conns = 25              # Max open connections
max_idle_conns = 10              # Max idle connections
conn_max_lifetime = 300          # Connection max lifetime (seconds)
```

**Environment variables**:
```bash
DATABASE_PATH="./data/app.db"
```

### Storage Configuration

See [Storage Configuration Guide](./storage.md) for detailed storage setup.

```toml
[storage]
provider = "local"               # "local" or "qiniu"
base_url = "/files"              # Base URL for file access

[storage.local]
root = "./data/objects"          # Local storage root

[storage.qiniu]
access_key = ""                  # Qiniu access key
secret_key = ""                  # Qiniu secret key
bucket = ""                      # Bucket name
domain = ""                      # CDN domain
region = "z0"                    # Region code
```

### Authentication Configuration

```toml
[auth]
enabled = true                   # Enable authentication
jwt_secret = ""                  # JWT signing secret (auto-generated if empty)
token_ttl = 86400                # Token TTL (seconds, default: 24 hours)
refresh_ttl = 604800             # Refresh token TTL (seconds, default: 7 days)
cookie_name = "cv_session"       # Session cookie name
cookie_secure = true             # HTTPS-only cookies
cookie_http_only = true          # HTTP-only cookies
cookie_same_site = "strict"      # SameSite attribute

[auth.rate_limit]
enabled = true                   # Enable rate limiting
requests_per_minute = 60         # Max requests per minute
burst = 10                       # Burst size
```

**Environment variables**:
```bash
AUTH_ENABLED=true
AUTH_JWT_SECRET="your-secret-key"
AUTH_TOKEN_TTL=86400
```

### AI Features Configuration

```toml
[ai]
enabled = true                   # Enable AI features
provider = "openai"              # AI provider: "openai", "azure", "local"
model = "gpt-4"                  # Model name
temperature = 0.7                # Temperature (0.0-2.0)
max_tokens = 4096                # Max tokens per request
timeout = 60                     # Request timeout (seconds)

[ai.openai]
api_key = ""                     # OpenAI API key
base_url = "https://api.openai.com/v1"  # API base URL

[ai.azure]
api_key = ""                     # Azure API key
endpoint = ""                    # Azure endpoint
deployment = ""                  # Deployment name

[ai.local]
endpoint = "http://localhost:11434"  # Local LLM endpoint (Ollama)
```

**Environment variables**:
```bash
AI_ENABLED=true
AI_PROVIDER="openai"
AI_MODEL="gpt-4"
AI_OPENAI_API_KEY="sk-..."
```

### Search Configuration

```toml
[search]
enabled = true                   # Enable full-text search
index_on_save = true             # Auto-index on document save
batch_size = 100                 # Indexing batch size
max_content_size = 1048576       # Max content size for indexing (1 MB)
snippet_length = 200             # Search snippet length
highlight_pre_tag = "<mark>"     # Highlight start tag
highlight_post_tag = "</mark>"   # Highlight end tag
```

**Environment variables**:
```bash
SEARCH_ENABLED=true
SEARCH_INDEX_ON_SAVE=true
```

### Import Configuration

```toml
[import]
max_file_size = 10485760         # Max file size (10 MB)
allowed_extensions = [".md", ".markdown", ".txt"]
duplicate_detection = true       # Detect duplicates on import
auto_tag = true                  # Auto-generate tags
extract_metadata = true          # Extract frontmatter metadata
```

**Environment variables**:
```bash
IMPORT_MAX_FILE_SIZE=10485760
IMPORT_DUPLICATE_DETECTION=true
```

### Backup Configuration

```toml
[backup]
enabled = true                   # Enable automatic backups
schedule = "0 2 * * *"           # Cron schedule (default: 2 AM daily)
retention_days = 30              # Keep backups for N days
max_backups = 10                 # Max number of backups
compression = true               # Compress backups
encryption = false               # Encrypt backups
encryption_key = ""              # Encryption key (if encryption enabled)
```

**Environment variables**:
```bash
BACKUP_ENABLED=true
BACKUP_SCHEDULE="0 2 * * *"
BACKUP_RETENTION_DAYS=30
```

### Logging Configuration

```toml
[logging]
level = "info"                   # Log level: debug, info, warn, error
format = "json"                  # Log format: json, text
output = "stdout"                # Log output: stdout, file
file_path = "./logs/app.log"     # Log file path (if output = file)
max_size = 104857600             # Max log file size (100 MB)
max_backups = 10                 # Max log file backups
max_age = 30                     # Max log file age (days)
compress = true                  # Compress old log files
redact_sensitive = true          # Redact sensitive data
```

**Environment variables**:
```bash
LOGGING_LEVEL="info"
LOGGING_FORMAT="json"
LOGGING_OUTPUT="stdout"
```

### Update Configuration

```toml
[update]
enabled = true                   # Enable update checking
check_on_startup = true          # Check on app startup
check_interval = 86400           # Check interval (seconds, default: 24 hours)
update_url = "https://releases.cloud-vault.com/latest.json"
auto_download = false            # Auto-download updates
auto_install = false             # Auto-install updates (desktop only)
```

**Environment variables**:
```bash
UPDATE_ENABLED=true
UPDATE_CHECK_ON_STARTUP=true
```

### Diagnostics Configuration

```toml
[diagnostics]
enabled = true                   # Enable diagnostics export
include_logs = true              # Include logs in bundle
include_config = true            # Include config (redacted)
include_database_stats = true    # Include database statistics
exclude_content = true           # Exclude document content
max_log_lines = 10000            # Max log lines to include
```

**Environment variables**:
```bash
DIAGNOSTICS_ENABLED=true
DIAGNOSTICS_INCLUDE_LOGS=true
```

### CORS Configuration

```toml
[cors]
enabled = true                   # Enable CORS
allowed_origins = ["*"]          # Allowed origins
allowed_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
allowed_headers = ["*"]          # Allowed headers
exposed_headers = []             # Exposed headers
allow_credentials = true         # Allow credentials
max_age = 3600                   # Preflight cache (seconds)
```

**Environment variables**:
```bash
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS="*"
```

### Rate Limiting Configuration

```toml
[rate_limit]
enabled = true                   # Enable rate limiting
requests_per_minute = 60         # Default rate limit
burst = 10                       # Burst size

[rate_limit.endpoints]
"/api/auth/login" = 5            # Login: 5 requests/minute
"/api/documents" = 100           # Documents: 100 requests/minute
"/api/search" = 30               # Search: 30 requests/minute
"/api/ai/*" = 20                 # AI endpoints: 20 requests/minute
```

**Environment variables**:
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
```

## Configuration Examples

### Minimal Configuration

```toml
# config.toml - Minimal setup

[server]
addr = ":8080"

[database]
path = "./data/app.db"

[storage]
provider = "local"

[storage.local]
root = "./data/objects"
```

### Production Configuration

```toml
# config.toml - Production setup

[server]
addr = ":443"
host = "0.0.0.0"
read_timeout = 30
write_timeout = 30

[database]
path = "/var/lib/cloud-vault/data/app.db"
max_open_conns = 50
max_idle_conns = 20

[storage]
provider = "qiniu"

[storage.qiniu]
access_key = "${QINIU_ACCESS_KEY}"
secret_key = "${QINIU_SECRET_KEY}"
bucket = "cloud-vault-prod"
domain = "https://cdn.yourdomain.com"
region = "z0"

[auth]
enabled = true
jwt_secret = "${JWT_SECRET}"
token_ttl = 86400
cookie_secure = true

[ai]
enabled = true
provider = "openai"
model = "gpt-4"

[ai.openai]
api_key = "${OPENAI_API_KEY}"

[search]
enabled = true
index_on_save = true

[backup]
enabled = true
schedule = "0 2 * * *"
retention_days = 30

[logging]
level = "info"
format = "json"
output = "file"
file_path = "/var/log/cloud-vault/app.log"
```

### Development Configuration

```toml
# config.toml - Development setup

[server]
addr = ":8080"

[database]
path = "./data/dev.db"

[storage]
provider = "local"

[storage.local]
root = "./data/objects"

[auth]
enabled = false  # Disable auth for development

[ai]
enabled = false  # Disable AI to save API calls

[search]
enabled = true

[logging]
level = "debug"
format = "text"
output = "stdout"
```

## Environment Variables

All configuration options can be set via environment variables:

```bash
# Server
export SERVER_ADDR=":8080"
export SERVER_HOST="0.0.0.0"

# Database
export DATABASE_PATH="./data/app.db"

# Storage
export STORAGE_PROVIDER="qiniu"
export STORAGE_QINIU_ACCESS_KEY="your-access-key"
export STORAGE_QINIU_SECRET_KEY="your-secret-key"

# Authentication
export AUTH_ENABLED=true
export AUTH_JWT_SECRET="your-secret"

# AI
export AI_ENABLED=true
export AI_PROVIDER="openai"
export AI_OPENAI_API_KEY="sk-..."

# Logging
export LOGGING_LEVEL="info"
export LOGGING_FORMAT="json"
```

## Configuration Validation

Validate your configuration:

```bash
./cloud-vault config validate
```

Output:
```
✓ Configuration is valid
✓ Database path is accessible
✓ Storage provider is configured
✓ Authentication is enabled
```

## Configuration Hot Reload

Some configuration changes can be applied without restart:

```bash
# Reload configuration
./cloud-vault config reload

# Or send SIGHUP signal
kill -HUP $(cat /var/run/cloud-vault.pid)
```

**Hot-reloadable settings**:
- Logging level
- Rate limits
- CORS settings
- AI model parameters

**Requires restart**:
- Server address/port
- Database path
- Storage provider
- Authentication settings

## Troubleshooting Configuration

### Configuration Not Loading

**Symptoms**: App uses default values instead of config file

**Solutions**:
```bash
# Check config file path
./cloud-vault --config /path/to/config.toml

# Verify file permissions
ls -l config.toml
# Should be readable by cloud-vault user

# Check for syntax errors
./cloud-vault config validate
```

### Environment Variables Not Working

**Symptoms**: Config file values used instead of env vars

**Solutions**:
```bash
# Verify env var is set
echo $AUTH_ENABLED

# Check env var name format
# Must be uppercase with underscores
export AUTH_ENABLED=true  # ✓ Correct
export auth.enabled=true  # ✗ Wrong

# Restart service to pick up env vars
sudo systemctl restart cloud-vault
```

### Invalid Configuration

**Error**: `Configuration validation failed`

**Solutions**:
1. Run validation: `./cloud-vault config validate`
2. Check error message for specific issues
3. Fix configuration file
4. Restart service

### Missing Required Fields

**Error**: `Required field X is missing`

**Solutions**:
1. Check configuration documentation
2. Add missing field to config.toml
3. Or set via environment variable

## Next Steps

- [Storage Configuration](./storage.md) - Detailed storage setup
- [Security](./security.md) - Security best practices
- [Server Deployment](./server-deploy.md) - Production deployment

---

**Last updated**: 2026-05-30
