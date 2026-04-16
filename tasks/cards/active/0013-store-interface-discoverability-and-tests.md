# Card 0013

Priority: P2
State: active
Primary Module: store
Owned Files:
  - store/file/file.go
  - store/idempotency/store.go

Depends On: —

Goal:
`store/file` 和 `store/idempotency` 是稳定根中的纯接口包（仅定义接口和类型），
具体实现分布在 `x/data/` 下，但包内无任何注释或文档说明这一设计决策，
也无指向实现的引导，导致：

1. 开发者找到 `store/file.Storage` 接口后不知道去哪里找实现
2. `store/idempotency/store_test.go` 仅 14 行，只用 `noopStore` 验证接口编译，
   没有覆盖 `ErrNotFound`、`ErrInvalidKey`、`ErrExpired` 等错误常量的语义正确性
3. `store/file` 的测试文件（`coverage_test.go`）是空占位符

同时，`store/cache/cache.go` 包含完整实现（MemoryCache），与上述两个
纯接口包风格截然不同，但无文档说明差异原因。

Scope:
- 在 `store/file/doc.go`（或文件顶部注释）中明确说明：
  - 本包定义稳定接口，实现在 `x/data/file`（local、S3 等）
  - 引用 `x/fileapi` 作为 HTTP handler 层
- 在 `store/idempotency/doc.go` 或文件注释中类似说明：
  - 实现在 `x/data/idempotency`（sql、kv）
- 补充 `store/idempotency/store_test.go`，测试以下行为：
  - `StatusInProgress`、`StatusComplete` 常量值合理（非零值）
  - `ErrNotFound`、`ErrExpired`、`ErrInvalidKey` 三个 sentinel error 均不为 nil 且互不相等
  - `Record` 结构体字段完整性（Key、Status、Response 字段可赋值）
- 删除 `store/file/coverage_test.go` 占位符（或补充有意义的编译测试）
- 在 `store/cache/doc.go` 或文件顶部注释说明：本包包含内存缓存实现
  （与 store/file、store/idempotency 的纯接口定位不同）

Non-goals:
- 不为 store/file 或 store/idempotency 编写新的具体实现
- 不改动 x/data/file 或 x/data/idempotency 的任何逻辑
- 不修改接口签名

Files:
  - store/file/file.go（顶部 package doc 补充说明）
  - store/file/coverage_test.go（删除或替换）
  - store/idempotency/store.go（顶部 package doc 补充说明）
  - store/idempotency/store_test.go（补充有意义的测试）
  - store/cache/cache.go（顶部注释说明定位差异）

Tests:
  - go test ./store/...

Docs Sync: —

Done Definition:
- `store/file` 和 `store/idempotency` 的 package doc 中含有指向实现所在包的明确说明
- `store/idempotency/store_test.go` 验证所有 sentinel error 非 nil 且互不相等
- `store/file` 无空占位符测试
- `go test ./store/...` 通过

Outcome:
