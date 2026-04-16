# Card 0006

Priority: P1
State: active
Primary Module: x/rest
Owned Files:
  - x/rest/resource.go

Depends On: —

Goal:
`BaseResourceController` 和 `BaseContextResourceController` 两个结构体的 10 个 handler 方法
（Index, Show, Create, Update, Delete, Patch, Options, Head, BatchCreate, BatchDelete）
代码完全相同，属于纯复制粘贴。应通过嵌入消除重复。

Scope:
- 让 `BaseContextResourceController` 嵌入 `BaseResourceController`，删除重复的 10 个方法
- 调整 `BaseContextResourceController.resourceName()` 调用，改为读取嵌入字段 `ResourceName`（已在 base 中公开）
- 保持两个类型的公共 API 不变，不影响任何已有调用者

Non-goals:
- 不合并两个结构体为一个
- 不改动 ResourceController 接口定义
- 不修改 QueryParams / QueryBuilder / ResourceOptions

Files:
  - x/rest/resource.go（行 58–95 BaseResourceController 方法，行 487–524 BaseContextResourceController 重复方法）

Tests:
  - go test ./x/rest/...
  - 确认两个 controller 类型均可被接口断言（ResourceController）

Docs Sync: —

Done Definition:
- `BaseContextResourceController` 不再声明任何与 `BaseResourceController` 重复的方法
- 所有 10 个方法通过嵌入自动满足 ResourceController 接口
- `go build ./x/rest/...` 与 `go test ./x/rest/...` 均通过
- `grep -n "func (c \*BaseContextResourceController)" x/rest/resource.go` 仅剩与 base 有差异的方法

Outcome:
