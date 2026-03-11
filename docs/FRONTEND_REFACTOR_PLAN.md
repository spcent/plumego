# frontend 包重构计划

> 基准：`docs/CANONICAL_STYLE_GUIDE.md`
> 原则：不考虑兼容，按最佳实践规划，清除所有偏离。

---

## 一、现状分析

`frontend` 包由 3 个文件组成：

| 文件 | 职责 |
|------|------|
| `frontend.go` | `Config`、`Option`、`Register*` 入口、`handler` 实现 |
| `embedded_fs.go` | `go:embed` 支持、`RegisterEmbedded` |
| `frontend_test.go` | 单元测试 |
| `example_test.go` | 示例文档 |

---

## 二、违反风格指南的具体问题

### 2.1 `Config` 不应导出

**文件**：`frontend.go:21-58`

**问题**：`Config` 是公开类型，但用户的唯一交互路径是 `Option` 函数。导出 `Config`：
- 允许用户绕过 `Option` 直接构造，产生两条入口（违反"one obvious way"）
- 将内部字段名作为 API 契约固化，提高维护成本
- `Option func(*Config)` 依赖导出类型，强行将 `Config` 泄漏给用户

**目标**：`Config` → `config`（未导出），`Option` 的签名改为 `func(*config)`。

---

### 2.2 `serveError` 回退使用 `http.Error`，与 `contract.WriteError` 不一致

**文件**：`frontend.go:562-570`、`frontend.go:331-338`

**问题**：同一个 `handler` 在不同路径返回不同格式的错误响应：

```go
// ServeHTTP 中：返回 JSON（contract.WriteError）
contract.WriteError(w, r, contract.APIError{...})

// serveError 中：回退返回纯文本
http.Error(w, message, code)       // ← 纯文本，不是 JSON
```

根据风格指南第 10 节："Identical error class → identical shape across modules"，同一个包必须使用统一的 error shape。

**目标**：`serveError` 的回退路径也使用 `contract.WriteError`，确保所有错误响应均为 JSON。

---

### 2.3 `embedded_fs.go` 使用可变包级变量作为测试接缝

**文件**：`embedded_fs.go:16-22`

```go
var embeddedFS fs.FS = embedded   // 测试中被替换
var embeddedRoot = "embedded"     // 测试中被替换
```

**问题**：
- 风格指南第 13 节明确要求："No hidden global state between tests"
- 测试通过 `embeddedFS = os.DirFS(dir)` 修改包级变量，不是通过 DI，而是通过隐式全局突变
- `defer` 恢复是脆弱的，并发测试下会产生数据竞争
- `HasEmbedded()` 是一个有状态的全局检查，鼓励调用方写 `if HasEmbedded() { RegisterEmbedded() }` 这种条件注册，与"explicit over implicit"相悖

**目标**：
- 删除 `embeddedFS`、`embeddedRoot` 两个可变包级变量
- `embedded` 保留为编译期常量（`//go:embed embedded/*` 生成，不可变）
- 删除 `HasEmbedded()`，让 `RegisterEmbedded` 在资产为空时直接返回 error
- 测试改为使用 `RegisterFS(r, http.FS(...), opts...)` + 临时目录，不再需要替换全局变量

---

### 2.4 `ServeHTTP` 路径提取逻辑过于复杂

**文件**：`frontend.go:329-426`

**问题**：近 100 行的 `ServeHTTP` 混合了三件事：
1. Method 校验
2. 路径标准化与 prefix 剥离
3. 文件查找与回退逻辑

路径剥离有 5 个分支（root 路径、prefix 精确匹配、prefix 非根、relative 为空、相对路径清洗），每个分支都内嵌 `serveFile` 调用，阅读困难。

**目标**：提取 `stripPrefix(requestPath, prefix string) (relativePath string, valid bool)` 纯函数，让 `ServeHTTP` 只做三件事：

```go
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet && r.Method != http.MethodHead {
        // 返回 405
        return
    }
    rel, ok := h.stripPrefix(r.URL.Path)
    if !ok {
        h.serveNotFound(w, r)
        return
    }
    h.serve(w, r, rel)   // serve 负责文件查找 + fallback
}
```

---

### 2.5 `WithMIMETypes` Option 函数含业务逻辑

**文件**：`frontend.go:132-150`

**问题**：Option 函数本应是纯粹的配置赋值（setter），但 `WithMIMETypes` 内部执行了：
- 空 map 判断
- 键值 trim
- 自动补全 `.` 前缀
- 空键过滤

风格指南要求 Option 函数只应修改 `Config`，不应有副作用逻辑。标准化/校验应当在 `RegisterFS` 统一处理（即 option 解析后的统一 validate 阶段）。

**目标**：`WithMIMETypes` 只做 `cfg.MIMETypes = mimeTypes`；在 `RegisterFS` 的参数校验阶段调用 `normalizeMIMETypes(raw map[string]string) map[string]string`。

---

### 2.6 `serveFile` 中响应头设置逻辑重复

**文件**：`frontend.go:432-503`

**问题**：`applyHeaders`、Cache-Control、Content-Type 的设置代码在 precompressed 分支和普通分支中**各写了一遍**，共 18 行重复代码。

**目标**：提取 `h.applyResponseHeaders(w http.ResponseWriter, filePath string)` 统一处理所有响应头，在两个分支之前调用一次。

---

### 2.7 `RegisterFromDir` 中多余的可读性检查

**文件**：`frontend.go:155-172`

```go
// Verify directory is readable
if f, err := os.Open(dir); err != nil {
    return fmt.Errorf("frontend directory %q not readable: %w", dir, err)
} else {
    f.Close()
}
```

**问题**：`os.Stat` 已经验证目录存在，`http.Dir` 的后续 `Open` 操作会捕获权限错误。多余的 `os.Open` 检查只增加了一个系统调用，没有新信息。

**目标**：删除该检查；若需要早期失败，直接在 `os.Stat` 的 `os.IsPermission` 中处理。

---

### 2.8 `acceptsEncoding` 使用 `strings.Contains`，存在误判风险

**文件**：`frontend.go:572-580`

```go
return strings.Contains(acceptEncoding, encoding)
```

**问题**：`strings.Contains(acceptEncoding, "br")` 会匹配含有 "br" 子串的任意 token，例如 `brotli`（假设将来有此 token）或 `Accept-Encoding: *;q=0.1, brrr`。HTTP Accept-Encoding 是逗号分隔的 token 列表，必须按 token 匹配。

**目标**：实现 `acceptsToken(header, token string) bool`，逐 token 比较（忽略权重参数 `;q=…`）：

```go
func acceptsToken(headerVal, token string) bool {
    for _, part := range strings.Split(headerVal, ",") {
        t, _, _ := strings.Cut(strings.TrimSpace(part), ";")
        if strings.EqualFold(strings.TrimSpace(t), token) {
            return true
        }
    }
    return false
}
```

---

### 2.9 `handler` 结构体完整复制 `Config` 字段

**文件**：`frontend.go:285-297`

**问题**：`handler` 有 9 个字段，与 `config` 字段一一对应，且在 `RegisterFS` 中逐一赋值（约 12 行代码）。这是不必要的双重表示，增加了维护成本，每次给 `config` 新增字段时都要同步修改 `handler` 和赋值代码。

**目标**：`handler` 直接嵌入 `config`：

```go
type handler struct {
    cfg config
    fs  http.FileSystem
}
```

访问时改为 `h.cfg.fallback`，构造时只需 `handler{cfg: *cfg, fs: fsys}`。

---

### 2.10 `example_test.go` 中的 `FullConfigurationExample` 不是合法 Example 函数

**文件**：`example_test.go:143-186`

**问题**：该函数名以 `FullConfigurationExample` 开头，不是 `Example` 前缀，Go 测试工具不会将其识别为示例函数，也不会在 godoc 中展示为 example。注释说"不可运行"，但这让它成为一段实际上不被任何工具处理的死代码。

**目标**：改名为 `ExampleRegisterFromDir_fullConfig` 或拆入文档注释，确保 godoc 可渲染。

---

### 2.11 测试中 `write` 闭包重复定义

**文件**：`frontend_test.go`（至少 6 个测试函数中定义相同的 `write` 闭包）

**问题**：相同逻辑被复制 6 次，任何改动需修改多处。

**目标**：提取为包级测试辅助函数：

```go
func writeTestFile(t *testing.T, dir, relPath, content string) {
    t.Helper()
    full := filepath.Join(dir, relPath)
    if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
        t.Fatalf("mkdir %s: %v", relPath, err)
    }
    if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
        t.Fatalf("write %s: %v", relPath, err)
    }
}
```

---

## 三、重构任务清单（按优先级排序）

### P0 — 正确性 / 一致性

| # | 任务 | 影响文件 |
|---|------|---------|
| T1 | 统一错误响应：`serveError` 回退改用 `contract.WriteError`，消除 plain-text/JSON 不一致 | `frontend.go` |
| T2 | 修复 `acceptsEncoding`：改为 token 级别匹配 `acceptsToken` | `frontend.go` |

### P1 — 全局状态 / 测试隔离

| # | 任务 | 影响文件 |
|---|------|---------|
| T3 | 删除 `embeddedFS`、`embeddedRoot` 可变包级变量，消除测试全局状态 | `embedded_fs.go`, `frontend_test.go` |
| T4 | 删除 `HasEmbedded()`，让 `RegisterEmbedded` 在空资产时返回 error | `embedded_fs.go` |
| T5 | 重写 `TestRegisterEmbedded`/`TestRegisterEmbeddedMissing`/`TestEmptyEmbeddedDirectory`/`TestNestedDirectoriesInEmbedded` 以不依赖包级变量替换 | `frontend_test.go` |

### P2 — API 清晰度

| # | 任务 | 影响文件 |
|---|------|---------|
| T6 | 将 `Config` 改为未导出的 `config`；`Option` 改为 `func(*config)` | `frontend.go` |
| T7 | 将 `WithMIMETypes` 还原为纯 setter；提取 `normalizeMIMETypes` 在 `RegisterFS` 验证阶段调用 | `frontend.go` |

### P3 — 可读性 / 可维护性

| # | 任务 | 影响文件 |
|---|------|---------|
| T8 | 提取 `stripPrefix(requestPath, prefix string) (string, bool)` 纯函数，简化 `ServeHTTP` | `frontend.go` |
| T9 | `handler` 嵌入 `config`，删除构造时的字段逐一复制 | `frontend.go` |
| T10 | 提取 `applyResponseHeaders`，消除 precompressed/普通分支中的重复头设置代码 | `frontend.go` |
| T11 | 删除 `RegisterFromDir` 中多余的 `os.Open` 可读性检查 | `frontend.go` |

### P4 — 测试 / 文档质量

| # | 任务 | 影响文件 |
|---|------|---------|
| T12 | 提取 `writeTestFile` 测试辅助函数，替换 6 个重复的 `write` 闭包 | `frontend_test.go` |
| T13 | 将 `FullConfigurationExample` 改名为合法 `Example*` 函数 | `example_test.go` |

---

## 四、重构后的目标文件结构

```
frontend/
├── frontend.go          # config(未导出), Option, handler, Register*, 工具函数
├── embedded_fs.go       # go:embed 声明 + RegisterEmbedded（无可变全局变量）
├── frontend_test.go     # 所有测试（不再替换包级变量）
└── example_test.go      # 全部合法 Example* 函数
```

**frontend.go 内部结构（重构后）**：

```
config（未导出）
Option func(*config)
With* 选项函数（纯 setter）

RegisterFromDir(r *router.Router, dir string, opts ...Option) error
RegisterFS(r *router.Router, fsys http.FileSystem, opts ...Option) error

handler struct { cfg config; fs http.FileSystem }
handler.ServeHTTP          → method check → stripPrefix → serve
handler.stripPrefix        → 纯函数，提取相对路径
handler.serve              → 文件查找 + fallback 决策
handler.serveFile          → 打开文件 + 调用 applyResponseHeaders + http.ServeContent
handler.applyResponseHeaders → headers + cache-control + content-type（统一）
handler.tryPrecompressed   → 尝试 .br / .gz
handler.tryOpenFile
handler.serveNotFound
handler.serveError         → 全部使用 contract.WriteError

acceptsToken               → 正确的 token 级别解析
normalizeMIMETypes         → 独立的标准化函数
normalizePrefix
isIndexFile
copyHeaders
statusCodeWriter
```

---

## 五、关键设计决策说明

### 5.1 为什么删除 `HasEmbedded()`？

`HasEmbedded()` 诱导调用方写：

```go
if frontend.HasEmbedded() {
    frontend.RegisterEmbedded(app)
}
```

这是隐式的条件注册。风格指南要求"一个明显的构建入口"。`RegisterEmbedded` 在无资产时返回 error，调用方根据 error 决策，更显式。

### 5.2 为什么 `Config` 要未导出？

唯一的公开入口是 `Register*` + `Option` 函数。导出 `Config` 给用户两条路径：

```go
// 路径 A（Option 风格）
RegisterFS(r, fs, WithFallback(true), WithCacheControl("..."))

// 路径 B（直接构造，如果 Config 是导出的）
cfg := frontend.Config{Fallback: true, CacheControl: "..."}
// 但如何使用 cfg？没有 RegisterWithConfig 函数，cfg 就成了死路
```

保持 `Config` 未导出，消除歧义。

### 5.3 为什么 `handler` 要嵌入 `config` 而不是用指针？

`config` 在构造后不再修改，值语义更安全，无需担心并发修改。`handler` 本身也是只读的（构造后不变），值嵌入是正确选择。

### 5.4 错误响应的统一

`frontend` 是 HTTP 服务器的一部分，客户端（包括 JavaScript SPA）可能会解析 4xx/5xx 错误响应。如果同一服务器在 API 路由返回 `{"error":{"code":"..."}}` 而静态文件路由返回 `internal server error\n`（plain text），前端代码必须处理两种格式，这是不可接受的。

---

## 六、重构验证标准

每个 PR/commit 完成后需通过：

```bash
go test -timeout 20s ./frontend/...
go vet ./frontend/...
gofmt -w ./frontend/
go test -race ./frontend/...
```

逐步验证：
- T1+T2 完成后：所有现有测试通过，新增 `TestAcceptsToken` 单元测试
- T3+T4+T5 完成后：`go test -race ./frontend/...` 无竞争报告
- T6+T7 完成后：无破坏性 API 变更（所有 `With*` 函数签名不变）
- T8~T11 完成后：`ServeHTTP` 不超过 30 行
- T12+T13 完成后：`go doc ./frontend/` 可渲染所有 Example 函数

---

## 七、不在本次重构范围内的事项

- **不迁移到 `io/fs.FS`**：`RegisterFS` 目前接受 `http.FileSystem`，与 `http.Dir`、`http.FS(embed.FS)` 兼容，无需修改。
- **不引入新的中间件**：`frontend` 仅是静态文件服务，不涉及业务中间件。
- **不调整路由注册策略**：`router.ANY` + `/*filepath` 的模式已经是框架惯例，保持不变。
- **不引入压缩运行时**：`WithPrecompressed` 只处理预先生成的 `.gz`/`.br` 文件，不在运行时压缩，保持不变。
