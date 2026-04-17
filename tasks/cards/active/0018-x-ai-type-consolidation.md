# Card 0018

Priority: P1
State: active
Primary Module: x/ai
Owned Files:
  - x/ai/provider/provider.go
  - x/ai/multimodal/provider.go
  - x/ai/multimodal/types.go
  - x/ai/tokenizer/tokenizer.go

Depends On: —

Goal:
`x/ai` 子域内核心数据类型被重复定义，三个子包各自独立维护相似类型，
调用方不得不在包边界手动转换：

**CompletionResponse — 定义两次，字段不对称：**

| 字段 | provider.CompletionResponse | multimodal.CompletionResponse |
|------|-----------------------------|-------------------------------|
| ID, Model, StopReason, Usage, Metadata | ✓ | ✓ |
| Content | `[]provider.ContentBlock` | `[]ContentBlock`（multimodal 自定义） |
| Role | ✗ 无 | ✓ 有 |
| JSON tag | ✗ 无 | ✓ 有 |
| StopReason 类型 | `provider.StopReason`（类型别名）| `string` |

**Message — 定义三次，语义重叠：**

| 字段 | provider.Message | multimodal.Message | tokenizer.Message |
|------|------------------|--------------------|-------------------|
| Role | `Role`（typed） | `MessageRole`（另一别名） | `string` |
| Content | `any`（string 或 ContentBlock） | `[]ContentBlock` | `string` |
| Name, ToolCallID | ✓ | ✓ | ✓ |
| Timestamp | ✗ | ✓ | ✗ |
| JSON tags | ✗ | ✓ | ✗ |

三套不同的 Role 类型（`Role`、`MessageRole`、`string`）让包间互操作充满转换代码。

**StreamReader — 双层包装：**
`multimodal.StreamReader` 内嵌 `*provider.StreamReader`，`Next()` 把
`provider.CompletionResponse` 转成 `multimodal.CompletionResponse`——
仅为添加 `Role` 字段就重新包装整个类型，是设计的信号问题。

Scope:
- **合并 CompletionResponse**：将 `Role` 字段加入 `provider.CompletionResponse`，
  补充 JSON tag；`multimodal.CompletionResponse` 改为 `= provider.CompletionResponse`
  的类型别名，或直接删除并更新 multimodal 调用方
- **统一 Role 类型**：将 `multimodal.MessageRole` 改为 `= provider.Role` 的类型别名，
  消除两套常量（RoleSystem/RoleUser/RoleAssistant/RoleTool 在两个包各定义一套）
- **合并 Message**：将 `Timestamp`（multimodal 独有）和 JSON tag 加入
  `provider.Message`；`multimodal.Message` 改为 `= provider.Message`；
  `tokenizer.Message` 因用途不同（仅 token 计数）保持独立，但改用 `string` Role
  明确标注"仅供 tokenizer 内部使用"
- **简化 StreamReader**：`multimodal.StreamReader` 的 `Next()` 直接返回
  `*provider.CompletionResponse`，删除额外的转换层
- 执行前 grep 全部调用方，分批迁移

Non-goals:
- 不合并 provider 和 multimodal 为同一个包
- 不改变 AI provider 实现（claude.go、openai.go）的内部逻辑
- 不影响 semanticcache 和 marketplace 子包（如有独立类型则单独跟进）

Files:
  - x/ai/provider/provider.go（CompletionResponse 加 Role，Message 加 Timestamp/JSON tag）
  - x/ai/multimodal/provider.go（CompletionResponse → 类型别名，StreamReader 简化）
  - x/ai/multimodal/types.go（Message → 类型别名，MessageRole → 类型别名）
  - x/ai/tokenizer/tokenizer.go（Message 加注释说明 tokenizer-only 用途）
  - x/ai/provider/claude.go、openai.go（更新构建 CompletionResponse 的代码）

Tests:
  - go build ./x/ai/...
  - go test ./x/ai/...

Docs Sync: —

Done Definition:
- `grep -rn "type CompletionResponse struct" ./x/ai/` 仅剩一处（provider 包）
- `grep -rn "type MessageRole\|type Role\b" ./x/ai/` 类型统一，multimodal 中为别名
- `go build ./x/ai/...` 通过
- `go test ./x/ai/...` 通过

Outcome:
