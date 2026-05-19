# Card 0018

Priority: P1
State: done
Primary Module: x/ai
Owned Files:
  - x/ai/provider/provider.go
  - x/ai/multimodal/provider.go
  - x/ai/multimodal/types.go
  - x/ai/tokenizer/tokenizer.go

Depends On: —

Goal:
Core data types within the `x/ai` sub-domain were defined redundantly; three sub-packages each
maintained similar types independently, forcing callers to convert manually at package boundaries:

**CompletionResponse — defined twice with asymmetric fields:**

| Field                                       | provider.CompletionResponse | multimodal.CompletionResponse         |
|---------------------------------------------|-----------------------------|---------------------------------------|
| ID, Model, StopReason, Usage, Metadata      | yes                         | yes                                   |
| Content                                     | `[]provider.ContentBlock`   | `[]ContentBlock` (multimodal-defined) |
| Role                                        | no                          | yes                                   |
| JSON tags                                   | no                          | yes                                   |
| StopReason type                             | `provider.StopReason` (alias) | `string`                            |

**Message — defined three times with overlapping semantics:**

| Field                     | provider.Message  | multimodal.Message       | tokenizer.Message |
|---------------------------|-------------------|--------------------------|-------------------|
| Role                      | `Role` (typed)    | `MessageRole` (alias)    | `string`          |
| Content                   | `any`             | `[]ContentBlock`         | `string`          |
| Name, ToolCallID          | yes               | yes                      | yes               |
| Timestamp                 | no                | yes                      | no                |
| JSON tags                 | no                | yes                      | no                |

Three different Role types (`Role`, `MessageRole`, `string`) filled inter-package code with
conversion boilerplate.

**StreamReader — double-wrapped:**
`multimodal.StreamReader` embedded `*provider.StreamReader` and its `Next()` converted
`provider.CompletionResponse` to `multimodal.CompletionResponse` — rewrapping the entire type
just to add a `Role` field was a design smell.

Scope:
- **Merge CompletionResponse**: Add `Role` field and JSON tags to `provider.CompletionResponse`;
  make `multimodal.CompletionResponse` a type alias (`= provider.CompletionResponse`), or
  delete it and update multimodal callers directly.
- **Unify Role type**: Make `multimodal.MessageRole` a type alias (`= provider.Role`),
  eliminating the duplicate constant sets (RoleSystem/RoleUser/RoleAssistant/RoleTool in
  both packages).
- **Merge Message**: Add `Timestamp` (multimodal-only) and JSON tags to `provider.Message`;
  make `multimodal.Message` a type alias (`= provider.Message`).
  `tokenizer.Message` stays independent (token-counting only) but use `string` Role and
  annotate it as "tokenizer-internal only".
- **Simplify StreamReader**: `multimodal.StreamReader.Next()` returns
  `*provider.CompletionResponse` directly; remove the conversion layer.
- Grep all callers before starting; migrate in batches.

Non-goals:
- Do not merge the provider and multimodal packages into one
- Do not change the internal logic of AI provider implementations (claude.go, openai.go)
- Do not affect semanticcache and marketplace sub-packages (track separately if they have
  independent types)

Files:
  - x/ai/provider/provider.go (CompletionResponse gains Role field + JSON tags; Message gains
    Timestamp + JSON tags)
  - x/ai/multimodal/provider.go (CompletionResponse → type alias; StreamReader simplified)
  - x/ai/multimodal/types.go (Message → type alias; MessageRole → type alias)
  - x/ai/tokenizer/tokenizer.go (Message annotated as tokenizer-only)
  - x/ai/provider/claude.go, openai.go (code building CompletionResponse updated)

Tests:
  - go build ./x/ai/...
  - go test ./x/ai/...

Docs Sync: —

Done Definition:
- `grep -rn "type CompletionResponse struct" ./x/ai/` returns exactly one result (provider package)
- `grep -rn "type MessageRole\|type Role\b" ./x/ai/` shows types unified; multimodal entry is a type alias
- `go build ./x/ai/...` passes
- `go test ./x/ai/...` passes

Outcome:
- Made `multimodal.CompletionResponse` a type alias: `type CompletionResponse = provider.CompletionResponse`
- Made `multimodal.MessageRole` a type alias: `type MessageRole = provider.Role`
- Changed `RoleSystem`/`RoleUser`/`RoleAssistant`/`RoleTool` constants to reference their
  `provider.*` equivalents
- Added JSON struct tags to `provider.CompletionResponse` fields and `provider.Message` fields
- Added `Role Role` field to `provider.CompletionResponse`
