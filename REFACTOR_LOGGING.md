# 日志模块重构方案

## 问题分析

### 当前状态

#### `store/db/sharding/logging.go` (215 行)
- 自定义 `LogLevel` 类型 (Debug, Info, Warn, Error)
- 完整的 `Logger` 结构体实现
- 支持 `context.Context`
- JSON 格式输出
- `WithFields/WithField` 模式用于结构化字段
- `LoggingRouter` 用于特定分片操作的日志记录
- **问题**: 与 `log/` 包功能重复,维护成本高

#### `log/` 包
1. **glog.go** (747 行) - Google 风格的日志实现
   - 功能完整:文件轮转、verbosity、符号链接
   - Level 系统 (INFO, WARNING, ERROR, FATAL)
   - 高性能:buffer pooling、caller 位置追踪
   - 配置灵活:按大小/时间/数量轮转

2. **logger.go** (136 行) - 接口定义
   - `StructuredLogger` 接口
   - `Fields map[string]any` 用于结构化字段
   - `gLogger` 适配器包装 glog
   - Lifecycle 支持 (Start/Stop)
   - **缺陷**: 没有 `context.Context` 支持

3. **traceid.go** (252 行) - Trace ID 生成器
   - Base62 编码
   - 时间戳 + 随机数 + 序列号

### 重叠部分
1. **日志级别**: 都有类似的级别概念 (Debug, Info, Warn, Error)
2. **结构化字段**: 都支持键值对字段
3. **输出控制**: 都支持可配置的输出
4. **并发安全**: 都使用互斥锁保护

### 关键差异
| 特性 | sharding/logging.go | log/glog.go | log/logger.go |
|------|---------------------|-------------|---------------|
| Context 支持 | ✅ | ❌ | ❌ |
| JSON 格式 | ✅ | ❌ | ❌ |
| 文件轮转 | ❌ | ✅ | ❌ |
| Verbosity | ❌ | ✅ | ❌ |
| 接口抽象 | ❌ | ❌ | ✅ |
| 性能优化 | 基础 | 高级 | 基础 |

---

## 重构方案

### 目标
1. **统一接口**: 扩展 `StructuredLogger` 接口,添加 `context.Context` 支持
2. **多种实现**: 提供 glog 和 JSON 两种实现
3. **易用性**: sharding 包可以直接使用,无需重复实现
4. **向后兼容**: 不破坏现有代码

### 阶段一: 增强 `log/` 包核心接口

#### 1.1 扩展 `StructuredLogger` 接口
**文件**: `log/logger.go`

```go
// StructuredLogger defines the minimal logging interface used by the application.
type StructuredLogger interface {
    // 基础方法 (保持向后兼容)
    WithFields(fields Fields) StructuredLogger
    Debug(msg string, fields Fields)
    Info(msg string, fields Fields)
    Warn(msg string, fields Fields)
    Error(msg string, fields Fields)

    // 新增: Context-aware 方法
    DebugCtx(ctx context.Context, msg string, fields Fields)
    InfoCtx(ctx context.Context, msg string, fields Fields)
    WarnCtx(ctx context.Context, msg string, fields Fields)
    ErrorCtx(ctx context.Context, msg string, fields Fields)

    // 新增: 格式化方法
    Debugf(format string, args ...any)
    Infof(format string, args ...any)
    Warnf(format string, args ...any)
    Errorf(format string, args ...any)

    // 新增: Context + 格式化
    DebugCtxf(ctx context.Context, format string, args ...any)
    InfoCtxf(ctx context.Context, format string, args ...any)
    WarnCtxf(ctx context.Context, format string, args ...any)
    ErrorCtxf(ctx context.Context, format string, args ...any)
}
```

#### 1.2 创建 JSON Logger 实现
**新文件**: `log/json.go`

```go
package glog

import (
    "context"
    "encoding/json"
    "io"
    "os"
    "sync"
    "time"
)

// JSONLogger implements StructuredLogger with JSON output format
type JSONLogger struct {
    mu     sync.Mutex
    output io.Writer
    level  Level
    fields Fields
}

type JSONLoggerConfig struct {
    Output io.Writer
    Level  Level
    Fields Fields
}

func NewJSONLogger(config JSONLoggerConfig) *JSONLogger {
    if config.Output == nil {
        config.Output = os.Stdout
    }
    if config.Fields == nil {
        config.Fields = make(Fields)
    }
    return &JSONLogger{
        output: config.Output,
        level:  config.Level,
        fields: config.Fields,
    }
}

func (l *JSONLogger) WithFields(fields Fields) StructuredLogger {
    merged := make(Fields, len(l.fields)+len(fields))
    for k, v := range l.fields {
        merged[k] = v
    }
    for k, v := range fields {
        merged[k] = v
    }
    return &JSONLogger{
        output: l.output,
        level:  l.level,
        fields: merged,
    }
}

func (l *JSONLogger) InfoCtx(ctx context.Context, msg string, fields Fields) {
    l.logCtx(ctx, INFO, msg, fields)
}

func (l *JSONLogger) logCtx(ctx context.Context, level Level, msg string, fields Fields) {
    if level < l.level {
        return
    }

    l.mu.Lock()
    defer l.mu.Unlock()

    // Build log entry
    entry := make(map[string]any)
    entry["time"] = time.Now().UTC().Format(time.RFC3339Nano)
    entry["level"] = levelNames[level]
    entry["msg"] = msg

    // Extract trace ID from context
    if traceID := TraceIDFromContext(ctx); traceID != "" {
        entry["trace_id"] = traceID
    }

    // Add default fields
    for k, v := range l.fields {
        entry[k] = v
    }

    // Add message-specific fields
    for k, v := range fields {
        entry[k] = v
    }

    // Marshal to JSON
    data, _ := json.Marshal(entry)
    l.output.Write(data)
    l.output.Write([]byte("\n"))
}

// Implement all other methods...
```

#### 1.3 增强 Context 支持
**新文件**: `log/context.go`

```go
package glog

import "context"

type contextKey int

const (
    traceIDKey contextKey = iota
    loggerKey
)

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from context
func TraceIDFromContext(ctx context.Context) string {
    if traceID, ok := ctx.Value(traceIDKey).(string); ok {
        return traceID
    }
    return ""
}

// WithLogger attaches a logger to the context
func WithLogger(ctx context.Context, logger StructuredLogger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext extracts the logger from context
func LoggerFromContext(ctx context.Context) StructuredLogger {
    if logger, ok := ctx.Value(loggerKey).(StructuredLogger); ok {
        return logger
    }
    return NewGLogger() // Default logger
}

// Auto-inject trace ID on first access
func LoggerFromContextOrNew(ctx context.Context) (StructuredLogger, context.Context) {
    if logger, ok := ctx.Value(loggerKey).(StructuredLogger); ok {
        return logger, ctx
    }

    // Generate trace ID if not present
    if TraceIDFromContext(ctx) == "" {
        ctx = WithTraceID(ctx, NewTraceID())
    }

    logger := NewGLogger()
    return logger, WithLogger(ctx, logger)
}
```

#### 1.4 更新 gLogger 适配器
**文件**: `log/logger.go`

为现有的 `gLogger` 添加 Context 方法:

```go
func (l *gLogger) DebugCtx(ctx context.Context, msg string, fields Fields) {
    if !V(1) {
        return
    }
    l.logWithFieldsAndContext(ctx, "DEBUG", msg, fields)
}

func (l *gLogger) InfoCtx(ctx context.Context, msg string, fields Fields) {
    l.logWithFieldsAndContext(ctx, "INFO", msg, fields)
}

// ... 其他方法

func (l *gLogger) logWithFieldsAndContext(ctx context.Context, level string, msg string, fields Fields) {
    combined := l.mergeFields(fields)

    // Add trace ID from context if present
    if traceID := TraceIDFromContext(ctx); traceID != "" {
        combined["trace_id"] = traceID
    }

    formatted := l.formatFields(combined)
    if formatted != "" {
        msg = msg + " " + formatted
    }

    // 调用底层 glog
    switch level {
    case "DEBUG":
        Info(msg)
    case "INFO":
        Info(msg)
    case "WARN":
        Warning(msg)
    case "ERROR":
        Error(msg)
    }
}
```

### 阶段二: 迁移 `store/db/sharding/logging.go`

#### 2.1 移除重复实现
删除 `store/db/sharding/logging.go` 中的以下部分:
- `LogLevel` 类型 (使用 `glog.Level`)
- `Logger` 结构体 (使用 `glog.JSONLogger`)
- 所有 `Debug/Info/Warn/Error` 方法

#### 2.2 重写 `LoggingRouter`
**文件**: `store/db/sharding/logging.go` (简化版)

```go
package sharding

import (
    "context"
    "fmt"
    "time"

    "github.com/spcent/plumego/log/glog"
)

// LoggingRouter wraps a router with structured logging
type LoggingRouter struct {
    router *Router
    logger glog.StructuredLogger
}

// NewLoggingRouter creates a logging router with optional logger
func NewLoggingRouter(router *Router, logger glog.StructuredLogger) *LoggingRouter {
    if logger == nil {
        // Default to JSON logger for sharding operations
        logger = glog.NewJSONLogger(glog.JSONLoggerConfig{
            Level: glog.INFO,
            Fields: glog.Fields{
                "component": "sharding",
            },
        })
    }

    return &LoggingRouter{
        router: router,
        logger: logger,
    }
}

func (lr *LoggingRouter) Logger() glog.StructuredLogger {
    return lr.logger
}

func (lr *LoggingRouter) Router() *Router {
    return lr.router
}

// LogQuery logs query execution with context
func (lr *LoggingRouter) LogQuery(ctx context.Context, query string, shardIndex int, latency time.Duration, err error) {
    fields := glog.Fields{
        "query":       query,
        "shard_index": shardIndex,
        "latency_ms":  latency.Milliseconds(),
    }

    if err != nil {
        fields["error"] = err.Error()
        lr.logger.ErrorCtx(ctx, "query failed", fields)
    } else {
        lr.logger.DebugCtx(ctx, "query executed", fields)
    }
}

// LogShardResolution logs shard resolution
func (lr *LoggingRouter) LogShardResolution(ctx context.Context, tableName string, shardKey any, shardIndex int) {
    lr.logger.DebugCtx(ctx, "shard resolved", glog.Fields{
        "table":       tableName,
        "shard_key":   fmt.Sprintf("%v", shardKey),
        "shard_index": shardIndex,
    })
}

// LogCrossShardQuery logs cross-shard queries
func (lr *LoggingRouter) LogCrossShardQuery(ctx context.Context, query string, policy string) {
    lr.logger.WarnCtx(ctx, "cross-shard query", glog.Fields{
        "query":  query,
        "policy": policy,
    })
}

// LogRewrite logs SQL rewriting
func (lr *LoggingRouter) LogRewrite(ctx context.Context, original, rewritten string, cached bool) {
    lr.logger.DebugCtx(ctx, "SQL rewritten", glog.Fields{
        "original":  original,
        "rewritten": rewritten,
        "cached":    cached,
    })
}
```

### 阶段三: 增强功能 (可选)

#### 3.1 日志采样 (高频日志场景)
**新文件**: `log/sampling.go`

```go
package glog

import (
    "sync"
    "time"
)

// SamplingLogger wraps a logger with rate limiting
type SamplingLogger struct {
    logger    StructuredLogger
    mu        sync.Mutex
    lastLog   map[string]time.Time
    threshold time.Duration
}

func NewSamplingLogger(logger StructuredLogger, threshold time.Duration) *SamplingLogger {
    return &SamplingLogger{
        logger:    logger,
        lastLog:   make(map[string]time.Time),
        threshold: threshold,
    }
}

func (s *SamplingLogger) InfoCtx(ctx context.Context, msg string, fields Fields) {
    if s.shouldLog(msg) {
        s.logger.InfoCtx(ctx, msg, fields)
    }
}

func (s *SamplingLogger) shouldLog(msg string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    now := time.Now()
    if last, ok := s.lastLog[msg]; ok {
        if now.Sub(last) < s.threshold {
            return false
        }
    }

    s.lastLog[msg] = now
    return true
}
```

#### 3.2 日志中间件 (HTTP 请求日志)
**新文件**: `log/middleware.go`

```go
package glog

import (
    "context"
    "net/http"
    "time"
)

// LoggingMiddleware creates HTTP middleware with request logging
func LoggingMiddleware(logger StructuredLogger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // Generate trace ID
            traceID := NewTraceID()
            ctx := WithTraceID(r.Context(), traceID)
            ctx = WithLogger(ctx, logger)

            // Wrap response writer to capture status
            rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

            // Log request start
            logger.InfoCtx(ctx, "request started", Fields{
                "method": r.Method,
                "path":   r.URL.Path,
                "remote": r.RemoteAddr,
            })

            // Process request
            next.ServeHTTP(rw, r.WithContext(ctx))

            // Log request completion
            logger.InfoCtx(ctx, "request completed", Fields{
                "method":   r.Method,
                "path":     r.URL.Path,
                "status":   rw.status,
                "duration": time.Since(start).Milliseconds(),
            })
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(status int) {
    rw.status = status
    rw.ResponseWriter.WriteHeader(status)
}
```

---

## 实施步骤

### Step 1: 增强 log/ 包 (1-2天)
- [ ] 扩展 `StructuredLogger` 接口添加 Context 方法
- [ ] 实现 `log/json.go` - JSON 格式 logger
- [ ] 实现 `log/context.go` - Context 工具函数
- [ ] 更新 `log/logger.go` 中的 gLogger 适配器
- [ ] 添加完整的单元测试

### Step 2: 迁移 sharding 包 (半天)
- [ ] 简化 `store/db/sharding/logging.go`
- [ ] 移除重复的 Logger 实现
- [ ] 更新 `LoggingRouter` 使用 `glog.StructuredLogger`
- [ ] 更新所有调用点传递 `context.Context`
- [ ] 运行测试确保功能正常

### Step 3: 文档和示例 (半天)
- [ ] 更新 `CLAUDE.md` 中的日志使用说明
- [ ] 添加 `examples/logging/` 示例代码
- [ ] 创建迁移指南

### Step 4: 可选增强 (1天)
- [ ] 实现日志采样 (`log/sampling.go`)
- [ ] 实现 HTTP 中间件 (`log/middleware.go`)
- [ ] 性能基准测试

---

## 收益

### 代码质量
- ✅ **减少重复**: 删除 ~200 行重复代码
- ✅ **统一接口**: 整个项目使用同一套日志接口
- ✅ **易于维护**: 日志逻辑集中在 `log/` 包

### 功能增强
- ✅ **Context 支持**: 自动传播 trace ID
- ✅ **灵活格式**: 可选择 glog 或 JSON 格式
- ✅ **性能优化**: 复用 glog 的 buffer pooling
- ✅ **向后兼容**: 现有代码无需大改

### 可扩展性
- ✅ **插件化**: 易于添加新的 logger 实现 (如 slog, zap)
- ✅ **中间件**: 统一的 HTTP 日志中间件
- ✅ **采样**: 高频场景的日志采样

---

## 风险和缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 接口变更破坏现有代码 | 中 | 保持向后兼容,添加新方法而非修改 |
| 性能下降 | 低 | 复用 glog 的优化,添加基准测试 |
| Context 传播复杂 | 低 | 提供工具函数简化使用 |
| 迁移工作量大 | 低 | 分阶段迁移,先 sharding 后其他模块 |

---

## 测试策略

### 单元测试
- `log/json_test.go` - JSON logger 各级别输出
- `log/context_test.go` - Context trace ID 传播
- `log/logger_test.go` - gLogger 适配器更新

### 集成测试
- `store/db/sharding/logging_test.go` - LoggingRouter 功能
- 跨包测试:验证 sharding 使用 log 包的正确性

### 性能测试
```go
func BenchmarkJSONLogger(b *testing.B) {
    logger := NewJSONLogger(JSONLoggerConfig{
        Output: io.Discard,
        Level:  INFO,
    })

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("benchmark", Fields{"iteration": i})
    }
}
```

---

## 后续优化方向

1. **结构化日志查询**: 添加日志解析和查询工具
2. **分布式追踪**: 集成 OpenTelemetry
3. **日志聚合**: 支持输出到 ELK/Loki
4. **动态配置**: 运行时调整日志级别
5. **指标集成**: 自动记录日志量和错误率指标

---

## 参考资料

- [Go 标准库 slog](https://pkg.go.dev/log/slog)
- [Uber zap](https://github.com/uber-go/zap)
- [Structured Logging Best Practices](https://www.dataset.com/blog/the-10-commandments-of-logging/)
