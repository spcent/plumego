# Config Module

The **config** package provides a comprehensive, type-safe, and extensible configuration management system for Go applications. It supports multiple configuration sources, built-in validation, hot reloading, and full backward compatibility with existing APIs.

## Quick Start

Create a new configuration system with multiple sources:

```go
// Create configuration instance
cfg := config.New()

// Add environment variables source
cfg.AddSource(config.NewEnvSource(""))

// Add file source with JSON format
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, false))

// Load configuration
ctx := context.Background()
if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Type-safe access with validation
serverPort, err := cfg.Int("server_port", 8080, &config.Range{Min: 1, Max: 65535})
if err != nil {
    log.Fatalf("Invalid server port: %v", err)
}
```

## Configuration Sources

### Environment Variables

Environment variables are automatically loaded from the process environment:

```go
// Load all environment variables
envSource := config.NewEnvSource("")
cfg.AddSource(envSource)

// Load from specific prefix
prefixedEnv := config.NewEnvSource("APP_")
cfg.AddSource(prefixedEnv)
```

Supported formats:
- `KEY=value`
- `KEY="quoted value"`
- `KEY='single quoted'`

### File Sources

Support for JSON and environment file formats:

```go
// JSON configuration file
jsonSource := config.NewFileSource("config.json", config.FormatJSON, false)
cfg.AddSource(jsonSource)

// .env file format
envFileSource := config.NewFileSource(".env", config.FormatEnv, false)
cfg.AddSource(envFileSource)

// Enable file watching for hot reload
watchingSource := config.NewFileSource("config.json", config.FormatJSON, true)
cfg.AddSource(watchingSource)
```

#### JSON Configuration Format
```json
{
  "server": {
    "port": 8080,
    "host": "localhost"
  },
  "database": {
    "url": "postgres://localhost:5432/myapp",
    "max_connections": 100
  },
  "features": {
    "debug_mode": true,
    "api_timeout": 5000
  }
}
```

#### Environment File Format
```ini
# Server configuration
SERVER_PORT=8080
SERVER_HOST=localhost

# Database configuration
DATABASE_URL=postgres://localhost:5432/myapp
DATABASE_MAX_CONNECTIONS=100

# Feature flags
DEBUG_MODE=true
API_TIMEOUT=5000
```

## Validation System

Built-in validators ensure configuration integrity:

### Required Fields
```go
apiKey, err := cfg.String("api_key", "", &config.Required{})
if err != nil {
    log.Fatalf("API key is required: %v", err)
}
```

### Value Ranges
```go
// Numeric range validation
port, err := cfg.Int("port", 8080, &config.Range{Min: 1, Max: 65535})
if err != nil {
    log.Fatalf("Port must be between 1-65535: %v", err)
}

// Float range validation
timeout, err := cfg.Float("timeout", 30.0, &config.Range{Min: 0.1, Max: 300.0})
if err != nil {
    log.Fatalf("Timeout must be between 0.1-300.0 seconds: %v", err)
}
```

### Pattern Matching
```go
// Email validation
emailPattern := config.Pattern{
    Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
}
email, err := cfg.String("admin_email", "", &emailPattern)
if err != nil {
    log.Fatalf("Invalid email format: %v", err)
}

// Custom pattern validation
customPattern := config.Pattern{
    Pattern: `^[A-Z]{3}-\d{4}$`, // Format: ABC-1234
}
code, err := cfg.String("promo_code", "", &customPattern)
if err != nil {
    log.Fatalf("Invalid promo code format: %v", err)
}
```

### URL Validation
```go
apiURL, err := cfg.String("api_url", "", &config.URL{})
if err != nil {
    log.Fatalf("Invalid API URL: %v", err)
}
```

### Enumerated Values
```go
// Only allow specific values
logLevel, err := cfg.String("log_level", "info", &config.OneOf{
    Values: []string{"debug", "info", "warn", "error"},
})
if err != nil {
    log.Fatalf("Log level must be debug, info, warn, or error: %v", err)
}
```

### String Length Validation
```go
// Minimum length
username, err := cfg.String("username", "", &config.MinLength{Min: 3})
if err != nil {
    log.Fatalf("Username must be at least 3 characters: %v", err)
}

// Maximum length
description, err := cfg.String("description", "", &config.MaxLength{Max: 500})
if err != nil {
    log.Fatalf("Description must not exceed 500 characters: %v", err)
}
```

## Type-Safe Accessors

### String Configuration
```go
appName := cfg.GetString("app_name", "MyApp")
```

### Integer Configuration
```go
serverPort := cfg.GetInt("server_port", 8080)
maxConnections := cfg.GetInt("max_connections", 100)
```

### Boolean Configuration
```go
debugMode := cfg.GetBool("debug_mode", false)
enableFeature := cfg.GetBool("enable_feature", true)
```

### Float Configuration
```go
timeout := cfg.GetFloat("timeout", 30.0)
rate := cfg.GetFloat("rate_limit", 1.5)
```

### Duration Configuration
```go
// Get duration in milliseconds
requestTimeout := cfg.GetDurationMs("request_timeout", 5000)
// Returns time.Duration type
```

### Type-Safe with Validation
```go
// Combine multiple validators
validatedPort, err := cfg.Int("server_port", 8080,
    &config.Range{Min: 1, Max: 65535},
    &config.Required{},
)
if err != nil {
    log.Fatalf("Invalid server port: %v", err)
}

// String with pattern and length validation
validatedAPIKey, err := cfg.String("api_key", "",
    &config.Required{},
    &config.MinLength{Min: 32},
    &config.MaxLength{Max: 64},
    &config.Pattern{Pattern: `^[a-zA-Z0-9]+$`},
)
if err != nil {
    log.Fatalf("Invalid API key: %v", err)
}
```

## Hot Reloading

Enable automatic configuration reloading when files change:

```go
// Create configuration with file watching
cfg := config.New()
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, true))

ctx := context.Background()
if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Watch for configuration changes
updates, errs := cfg.Watch(ctx)
for {
    select {
    case <-ctx.Done():
        return
    case update := <-updates:
        log.Println("Configuration updated:", update)
    case err := <-errs:
        log.Printf("Configuration watch error: %v", err)
    }
}
```

## Schema Validation

Define validation schemas for complex configuration structures:

```go
// Create validation schema
schema := config.NewConfigSchema()

// Add field validation rules
schema.AddField("server_port", 
    &config.Required{},
    &config.Range{Min: 1, Max: 65535},
)

schema.AddField("database_url", 
    &config.Required{},
    &config.URL{},
)

schema.AddField("log_level",
    &config.OneOf{Values: []string{"debug", "info", "warn", "error"}},
)

// Validate configuration data
cfg := config.New()
cfg.AddSource(config.NewEnvSource(""))
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, false))

if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Validate against schema
configData := cfg.GetAll()
if err := schema.Validate(configData); err != nil {
    log.Fatalf("Configuration validation failed: %v", err)
}
```

## Configuration Unmarshaling

Unmarshal configuration into typed structs:

```go
// Define configuration struct
type AppConfig struct {
    AppName    string  `json:"app_name"`
    AppVersion string  `json:"app_version"`
    Server     struct {
        Port    int     `json:"port"`
        Host    string  `json:"host"`
        Timeout float64 `json:"timeout"`
    } `json:"server"`
    Database struct {
        URL            string `json:"url"`
        MaxConnections int    `json:"max_connections"`
    } `json:"database"`
    Features struct {
        DebugMode   bool `json:"debug_mode"`
        EnableCache bool `json:"enable_cache"`
    } `json:"features"`
}

// Load configuration
cfg := config.New()
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, false))

if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Unmarshal into struct
var appConfig AppConfig
if err := cfg.Unmarshal(&appConfig); err != nil {
    log.Fatalf("Failed to unmarshal config: %v", err)
}

log.Printf("Starting %s v%s on %s:%d", 
    appConfig.AppName, 
    appConfig.AppVersion,
    appConfig.Server.Host,
    appConfig.Server.Port,
)
```

## Backward Compatibility

The config package maintains full backward compatibility with existing global functions:

```go
// All existing functions still work
config.LoadEnv()
config.LoadEnvFile("config.env")

// New configuration system integrates seamlessly
cfg := config.New()
cfg.AddSource(config.NewEnvSource(""))
if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Use either approach or mix them
serverPort := config.GetInt("server_port", 8080)  // Old way
validatedPort := cfg.GetInt("server_port", 8080)  // New way
```

## Advanced Usage

### Custom Configuration Sources

Create custom configuration sources by implementing the Source interface:

```go
type DatabaseSource struct {
    db *sql.DB
}

func (d *DatabaseSource) Load(ctx context.Context) (map[string]any, error) {
    rows, err := d.db.QueryContext(ctx, "SELECT key, value FROM config")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    config := make(map[string]any)
    for rows.Next() {
        var key, value string
        if err := rows.Scan(&key, &value); err != nil {
            return nil, err
        }
        config[key] = value
    }
    
    return config, nil
}

func (d *DatabaseSource) Watch(ctx context.Context) (<-chan map[string]any, <-chan error) {
    // Implement database change notifications
    // ...
    return updates, errs
}

// Use custom source
dbSource := &DatabaseSource{db: yourDB}
cfg.AddSource(dbSource)
```

### Configuration Priority

Configuration sources are checked in the order they are added:

```go
cfg := config.New()

// Environment variables have highest priority
cfg.AddSource(config.NewEnvSource(""))

// File configuration provides defaults
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, false))

// Database configuration provides base values
cfg.AddSource(&DatabaseSource{db: yourDB})
```

## Best Practices

1. **Use Type-Safe Accessors**: Always use typed accessors (`GetInt`, `GetString`, etc.) instead of `Get` to avoid type conversion errors.

2. **Validate Early**: Apply validation when loading configuration to catch issues during application startup.

3. **Provide Sensible Defaults**: Use default values that make sense for your application.

4. **Use Hot Reloading for Development**: Enable file watching during development for faster iteration.

5. **Document Your Schema**: Keep configuration documentation up to date with your validation rules.

6. **Separate Development and Production**: Use different configuration sources for different environments.

7. **Monitor Configuration Changes**: Implement proper logging for configuration updates in production.

```go
// Example: Production-ready configuration setup
func setupConfig() (*config.Config, error) {
    cfg := config.New()
    
    // Add environment variables (highest priority)
    cfg.AddSource(config.NewEnvSource(""))
    
    // Add file configuration (development defaults)
    if env := os.Getenv("APP_ENV"); env != "production" {
        cfg.AddSource(config.NewFileSource("config.dev.json", config.FormatJSON, true))
    } else {
        cfg.AddSource(config.NewFileSource("config.prod.json", config.FormatJSON, false))
    }
    
    // Validate configuration
    ctx := context.Background()
    if err := cfg.Load(ctx); err != nil {
        return nil, fmt.Errorf("failed to load configuration: %w", err)
    }
    
    // Validate critical settings
    schema := config.NewConfigSchema()
    schema.AddField("server_port", &config.Required{}, &config.Range{Min: 1, Max: 65535})
    schema.AddField("database_url", &config.Required{}, &config.URL{})
    
    if err := schema.Validate(cfg.GetAll()); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return cfg, nil
}
```

## Configuration Reference

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `APP_ENV` | Application environment | `development` | `production` |
| `SERVER_PORT` | Server listening port | `8080` | `3000` |
| `SERVER_HOST` | Server host address | `localhost` | `0.0.0.0` |
| `DATABASE_URL` | Database connection string | - | `postgres://user:pass@localhost:5432/db` |
| `DEBUG_MODE` | Enable debug features | `false` | `true` |
| `LOG_LEVEL` | Logging level | `info` | `debug` |

### File Formats

#### JSON Structure
```json
{
  "app": {
    "name": "MyApp",
    "version": "1.0.0"
  },
  "server": {
    "port": 8080,
    "host": "localhost",
    "timeout": 30.0
  },
  "database": {
    "url": "postgres://localhost:5432/myapp",
    "max_connections": 100,
    "ssl_mode": "require"
  },
  "features": {
    "debug_mode": false,
    "enable_metrics": true,
    "api_timeout": 5000
  }
}
```

#### Environment File Structure
```env
# Application
APP_NAME=MyApp
APP_VERSION=1.0.0

# Server
SERVER_PORT=8080
SERVER_HOST=localhost
SERVER_TIMEOUT=30.0

# Database
DATABASE_URL=postgres://localhost:5432/myapp
DATABASE_MAX_CONNECTIONS=100
DATABASE_SSL_MODE=require

# Features
DEBUG_MODE=false
ENABLE_METRICS=true
API_TIMEOUT=5000
```

This comprehensive configuration system provides enterprise-grade features while maintaining simplicity and ease of use for Go applications of any size.