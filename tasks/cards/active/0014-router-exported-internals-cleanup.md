# Card 0014

Priority: P2
State: active
Primary Module: router
Owned Files:
  - router/types.go
  - router/cache.go

Depends On: —

Goal:
`router` 包将多个仅内部使用的符号导出，污染了公共 API 面：

1. **`MatchResult`**（router/types.go:8）：`matchRoute` 返回 `*MatchResult`，
   但该方法是包私有方法（小写），`MatchResult` 本身从未被包外代码直接引用。
   导出一个调用方永远无法触达的类型无实际意义。

2. **`DefaultPatternCacheSize = 50`**（router/cache.go:13）：该常量在 `cache.go`
   构造函数中内部使用，没有任何包外调用，但被导出，让使用者误以为可配置。

3. 与此同时，router 包的 `RouterOption`（router.go:86）和 `RouteOption`
   （registration.go）共存，命名风格不统一——前者是中间件注入选项，
   后者是路由注册选项，用途和层次均不同，却命名相似，易混淆。

Scope:
- 将 `MatchResult` 改为 `matchResult`（unexported），更新包内所有引用
  （router/cache.go、router/dispatch.go、router/matcher.go）
- 将 `DefaultPatternCacheSize` 改为 `defaultPatternCacheSize`，更新包内引用
- 若 `DefaultPoolSliceCap`（pool.go）同样仅内部使用，一并改为 unexported
- 在 router/types.go 顶部注释中说明 RouterOption 与 RouteOption 的用途差异，
  消除命名混淆（若有重命名价值则作为 future work 单独卡片处理）
- 执行前：`grep -rn "router\.MatchResult\|router\.DefaultPatternCacheSize" . --include="*.go"`
  确认无包外引用

Non-goals:
- 不改变路由匹配算法或性能特性
- 不重命名 RouterOption / RouteOption（仅加注释说明）
- 不修改 router 包的公共 API（Handle、Group 等方法）

Files:
  - router/types.go（MatchResult → matchResult）
  - router/cache.go（DefaultPatternCacheSize → defaultPatternCacheSize，更新引用）
  - router/dispatch.go（更新 matchResult 引用）
  - router/matcher.go（更新 matchResult 引用）
  - router/pool.go（检查 DefaultPoolSliceCap）

Tests:
  - go build ./router/...
  - go test ./router/...

Docs Sync: —

Done Definition:
- `grep -rn "router\.MatchResult" . --include="*.go"` 结果为空
- `grep -rn "router\.DefaultPatternCacheSize" . --include="*.go"` 结果为空
- `go build ./...` 通过（确认无外部依赖被破坏）
- `go test ./router/...` 通过

Outcome:
