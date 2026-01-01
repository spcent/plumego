# Config 模块

**config** 包为 Go 应用程序提供全面、类型安全且可扩展的配置管理系统。它支持多种配置源、内置验证、热重载，并与现有 API 完全向后兼容。

## 快速开始

创建包含多个源的新配置系统：

```go
// 创建配置实例
cfg := config.New()

// 添加环境变量源
cfg.AddSource(config.NewEnvSource(""))

// 添加 JSON 格式的文件源
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, false))

// 加载配置
ctx := context.Background()
if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 类型安全访问并验证
serverPort, err := cfg.Int("server_port", 8080, &config.Range{Min: 1, Max: 65535})
if err != nil {
    log.Fatalf("Invalid server port: %v", err)
}
```

## 配置源

### 环境变量

自动从进程环境加载环境变量：

```go
// 加载所有环境变量
envSource := config.NewEnvSource("")
cfg.AddSource(envSource)

// 从特定前缀加载
prefixedEnv := config.NewEnvSource("APP_")
cfg.AddSource(prefixedEnv)
```

支持的格式：
- `KEY=value`
- `KEY="quoted value"`
- `KEY='single quoted'`

### 文件源

支持 JSON 和环境文件格式：

```go
// JSON 配置文件
jsonSource := config.NewFileSource("config.json", config.FormatJSON, false)
cfg.AddSource(jsonSource)

// .env 文件格式
envFileSource := config.NewFileSource(".env", config.FormatEnv, false)
cfg.AddSource(envFileSource)

// 启用文件监控以实现热重载
watchingSource := config.NewFileSource("config.json", config.FormatJSON, true)
cfg.AddSource(watchingSource)
```

#### JSON 配置格式
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

#### 环境文件格式
```ini 
# 服务器配置
SERVER_PORT=8080
SERVER_HOST=localhost

# 数据库配置
DATABASE_URL=postgres://localhost:5432/myapp
DATABASE_MAX_CONNECTIONS=100

# 功能标志
DEBUG_MODE=true
API_TIMEOUT=5000
```

## 验证系统

内置验证器确保配置完整性：

### 必填字段
```go
apiKey, err := cfg.String("api_key", "", &config.Required{})
if err != nil {
    log.Fatalf("API key is required: %v", err)
}
```

### 值范围
```go
// 数值范围验证
port, err := cfg.Int("port", 8080, &config.Range{Min: 1, Max: 65535})
if err != nil {
    log.Fatalf("Port must be between 1-65535: %v", err)
}

// 浮点范围验证
timeout, err := cfg.Float("timeout", 30.0, &config.Range{Min: 0.1, Max: 300.0})
if err != nil {
    log.Fatalf("Timeout must be between 0.1-300.0 seconds: %v", err)
}
```

### 模式匹配
```go
// 邮箱验证
emailPattern := config.Pattern{
    Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
}
email, err := cfg.String("admin_email", "", &emailPattern)
if err != nil {
    log.Fatalf("Invalid email format: %v", err)
}

// 自定义模式验证
customPattern := config.Pattern{
    Pattern: `^[A-Z]{3}-\d{4}$`, // 格式：ABC-1234
}
code, err := cfg.String("promo_code", "", &customPattern)
if err != nil {
    log.Fatalf("Invalid promo code format: %v", err)
}
```

### URL 验证
```go
apiURL, err := cfg.String("api_url", "", &config.URL{})
if err != nil {
    log.Fatalf("Invalid API URL: %v", err)
}
```

### 枚举值
```go
// 仅允许特定值
logLevel, err := cfg.String("log_level", "info", &config.OneOf{
    Values: []string{"debug", "info", "warn", "error"},
})
if err != nil {
    log.Fatalf("Log level must be debug, info, warn, or error: %v", err)
}
```

### 字符串长度验证
```go
// 最小长度
username, err := cfg.String("username", "", &config.MinLength{Min: 3})
if err != nil {
    log.Fatalf("Username must be at least 3 characters: %v", err)
}

// 最大长度
description, err := cfg.String("description", "", &config.MaxLength{Max: 500})
if err != nil {
    log.Fatalf("Description must not exceed 500 characters: %v", err)
}
```

## 类型安全访问器

### 字符串配置
```go
appName := cfg.GetString("app_name", "MyApp")
```

### 整数配置
```go
serverPort := cfg.GetInt("server_port", 8080)
maxConnections := cfg.GetInt("max_connections", 100)
```

### 布尔配置
```go
debugMode := cfg.GetBool("debug_mode", false)
enableFeature := cfg.GetBool("enable_feature", true)
```

### 浮点配置
```go
timeout := cfg.GetFloat("timeout", 30.0)
rate := cfg.GetFloat("rate_limit", 1.5)
```

### 持续时间配置
```go
// 获取毫秒单位的持续时间
requestTimeout := cfg.GetDurationMs("request_timeout", 5000)
// 返回 time.Duration 类型
```

### 类型安全与验证
```go
// 组合多个验证器
validatedPort, err := cfg.Int("server_port", 8080,
    &config.Range{Min: 1, Max: 65535},
    &config.Required{},
)
if err != nil {
    log.Fatalf("Invalid server port: %v", err)
}

// 字符串模式验证和长度验证
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

## 热重载

当文件发生变化时启用自动配置重载：

```go
// 创建带文件监控的配置
cfg := config.New()
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, true))

ctx := context.Background()
if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 监控配置变化
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

## 模式验证

为复杂配置结构定义验证模式：

```go
// 创建验证模式
schema := config.NewConfigSchema()

// 添加字段验证规则
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

// 验证配置数据
cfg := config.New()
cfg.AddSource(config.NewEnvSource(""))
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, false))

if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 根据模式验证
configData := cfg.GetAll()
if err := schema.Validate(configData); err != nil {
    log.Fatalf("Configuration validation failed: %v", err)
}
```

## 配置解组

将配置解组为类型化结构体：

```go
// 定义配置结构体
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

// 加载配置
cfg := config.New()
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, false))

if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 解组为结构体
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

## 向后兼容性

config 包与现有的全局函数保持完全向后兼容：

```go
// 所有现有函数仍然有效
config.LoadEnv()
config.LoadEnvFile("config.env")

// 新配置系统无缝集成
cfg := config.New()
cfg.AddSource(config.NewEnvSource(""))
if err := cfg.Load(ctx); err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// 可以使用任一方法或混合使用
serverPort := config.GetInt("server_port", 8080)  // 旧方式
validatedPort := cfg.GetInt("server_port", 8080)  // 新方式
```

## 高级用法

### 自定义配置源

通过实现 Source 接口创建自定义配置源：

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
    // 实现数据库变更通知
    // ...
    return updates, errs
}

// 使用自定义源
dbSource := &DatabaseSource{db: yourDB}
cfg.AddSource(dbSource)
```

### 配置优先级

按添加顺序检查配置源：

```go
cfg := config.New()

// 环境变量具有最高优先级
cfg.AddSource(config.NewEnvSource(""))

// 文件配置提供默认值
cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, false))

// 数据库配置提供基础值
cfg.AddSource(&DatabaseSource{db: yourDB})
```

## 最佳实践

1. **使用类型安全访问器**：始终使用类型化访问器（`GetInt`、`GetString` 等）而不是 `Get`，以避免类型转换错误。

2. **早期验证**：在加载配置时应用验证，以在应用程序启动期间发现问题。

3. **提供合理默认值**：使用对您的应用程序有意义的默认值。

4. **在开发中使用热重载**：在开发期间启用文件监控以实现更快迭代。

5. **记录模式**：保持配置文档与验证规则同步更新。

6. **分离开发和生产**：为不同环境使用不同的配置源。

7. **监控配置变化**：在生产环境中为配置更新实现适当的日志记录。

```go
// 示例：生产就绪的配置设置
func setupConfig() (*config.Config, error) {
    cfg := config.New()
    
    // 添加环境变量（最高优先级）
    cfg.AddSource(config.NewEnvSource(""))
    
    // 添加文件配置（开发默认值）
    if env := os.Getenv("APP_ENV"); env != "production" {
        cfg.AddSource(config.NewFileSource("config.dev.json", config.FormatJSON, true))
    } else {
        cfg.AddSource(config.NewFileSource("config.prod.json", config.FormatJSON, false))
    }
    
    // 验证配置
    ctx := context.Background()
    if err := cfg.Load(ctx); err != nil {
        return nil, fmt.Errorf("failed to load configuration: %w", err)
    }
    
    // 验证关键设置
    schema := config.NewConfigSchema()
    schema.AddField("server_port", &config.Required{}, &config.Range{Min: 1, Max: 65535})
    schema.AddField("database_url", &config.Required{}, &config.URL{})
    
    if err := schema.Validate(cfg.GetAll()); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return cfg, nil
}
```

## 配置参考

### 环境变量

| 变量 | 描述 | 默认值 | 示例 |
|------|------|--------|------|
| `APP_ENV` | 应用程序环境 | `development` | `production` |
| `SERVER_PORT` | 服务器监听端口 | `8080` | `3000` |
| `SERVER_HOST` | 服务器主机地址 | `localhost` | `0.0.0.0` |
| `DATABASE_URL` | 数据库连接字符串 | - | `postgres://user:pass@localhost:5432/db` |
| `DEBUG_MODE` | 启用调试功能 | `false` | `true` |
| `LOG_LEVEL` | 日志级别 | `info` | `debug` |

### 文件格式

#### JSON 结构
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

#### 环境文件结构
```env
# 应用程序
APP_NAME=MyApp
APP_VERSION=1.0.0

# 服务器
SERVER_PORT=8080
SERVER_HOST=localhost
SERVER_TIMEOUT=30.0

# 数据库
DATABASE_URL=postgres://localhost:5432/myapp
DATABASE_MAX_CONNECTIONS=100
DATABASE_SSL_MODE=require

# 功能
DEBUG_MODE=false
ENABLE_METRICS=true
API_TIMEOUT=5000
```

这个全面的配置系统为企业级功能提供支持，同时为任何规模的 Go 应用程序保持简洁和易用性。