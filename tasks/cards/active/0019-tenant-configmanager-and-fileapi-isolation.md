# Card 0019

Priority: P1
State: active
Primary Module: x/tenant
Owned Files:
  - x/tenant/core/config.go
  - x/fileapi/handler.go

Depends On: —

Goal:
两个问题均涉及 tenant 隔离的正确性，归为一张卡：

**问题 1：x/tenant.ConfigManager 接口过宽（core/config.go:52-61）**

`ConfigManager` 将三个独立的 provider 接口嵌入为一个大接口：

```go
type ConfigManager interface {
    QuotaConfigProvider
    PolicyConfigProvider
    RateLimitConfigProvider
    GetTenantConfig(ctx, tenantID) (Config, error)
}
```

注释承认这是"便利性"设计（"so a single implementation can be passed directly"），
但代价是：
- 任何新增的 provider 接口都强制所有 ConfigManager 实现者同步实现
- 业务代码中持有 ConfigManager 的组件实际上只需要其中一个子接口，
  但被迫依赖整个 facade，违反最小权限接口原则
- InMemoryConfigManager 是唯一实现，使其成为隐式单体

修复：保留 ConfigManager 作为 app 层 facade，但让各子系统（quota、policy、ratelimit）
依赖自己的专项 provider 接口而非 ConfigManager，从而解耦。

**问题 2：x/fileapi 部分 handler 未校验 tenant 上下文（handler.go）**

`Upload()`（line 53-100）从 context 读取 tenantID 并在缺失时拒绝请求。
但以下 handler 完全没有 tenant 检查：

- `Download()`（line 121）：通过 fileID 直接访问，未验证调用方是否属于文件所在租户
- `GetInfo()`（line 163）：同上
- `Delete()`（line 187）：同上
- `List()`（line 210）：通过查询参数 `tenant_id` 过滤，非身份验证校验

攻击者只需知道 fileID 即可跨租户读取或删除文件。

另外，`Download()` 的 line 141：
```go
go h.metadata.UpdateAccessTime(context.Background(), fileID)
```
fire-and-forget goroutine 使用 `context.Background()`，无法参与请求链路追踪，
也无法在 server shutdown 时优雅退出。

Scope:
- **ConfigManager 解耦**：
  - 各子系统函数签名改为依赖最小接口（QuotaConfigProvider、PolicyConfigProvider、
    RateLimitConfigProvider），而非整个 ConfigManager
  - ConfigManager 接口本身保持不变（app 层仍可传入单一实现）
  - 在 core/config.go 顶部注释说明：ConfigManager 是 app 层 facade，
    子系统应依赖各自的专项 provider 接口
- **fileapi 租户隔离**：
  - `Download`、`GetInfo`、`Delete`：读取 tenantID（同 Upload），
    验证文件的 `FileMeta.TenantID` 是否与 context 中一致，不一致返回 403
  - `List`：忽略查询参数 `tenant_id`，改为强制从 context 读取（消除越权列举可能）
  - `GetURL`：同 Download 逻辑
  - 删除 fire-and-forget goroutine，改为在请求处理链路内同步调用
    （或接受丢失 access time 更新并在 handler 关闭时 flush），
    并在失败时记录警告日志而非静默丢弃

Non-goals:
- 不改变 InMemoryConfigManager 的实现逻辑
- 不为 ConfigManager 添加新方法
- 不改变文件存储后端（x/data/file）
- 不修改 fileapi 的路由注册或 URL 结构

Files:
  - x/tenant/core/config.go（注释说明 + 子系统接口依赖更新）
  - x/tenant/core/quota.go、policy.go、ratelimit.go（改依赖专项 provider 接口）
  - x/fileapi/handler.go（Download/GetInfo/Delete/List/GetURL 加 tenant 校验，
    移除 fire-and-forget goroutine）
  - x/fileapi/handler_test.go（补充跨租户访问被拒的测试用例）

Tests:
  - go test ./x/tenant/...
  - go test ./x/fileapi/...

Docs Sync: —

Done Definition:
- `grep -n "go h.metadata.UpdateAccessTime" x/fileapi/handler.go` 结果为空
- `Download`、`GetInfo`、`Delete` 均含 tenantID 校验，跨租户返回 403
- `List` 不再从 query param 读取 tenant_id 作为过滤来源
- `go test ./x/fileapi/...` 含跨租户 403 测试且通过

Outcome:
