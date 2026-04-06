# Card 0806

Milestone: contract cleanup
Priority: P3
State: done
Primary Module: contract
Owned Files:
- `contract/error_handling_test.go`
Depends On: —

Goal:
- Replace the hand-rolled `contains`/`findSubstring` helpers in `error_handling_test.go`
  with the stdlib `strings.Contains`.

Problem:
`error_handling_test.go` defines two unexported helpers for substring checking:

```go
func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr ||
        len(s) > len(substr) && (s[:len(substr)] == substr ||
        s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr { return true }
    }
    return false
}
```

This is a 20-line reimplementation of `strings.Contains`. The standard library
function is simpler, well-tested, and familiar to every Go reader. The file
currently does not import `"strings"` even though it already uses types from that
package in other test files in the same directory.

The `contains` helper is called exactly once (line 170). `findSubstring` is called
only from `contains`. Both helpers are dead code beyond that one call.

Scope:
- Delete `contains` and `findSubstring` from `error_handling_test.go`.
- Replace the single call site with `strings.Contains`.
- Add `"strings"` to the import block.

Non-goals:
- No changes to test assertions or coverage.
- Do not touch the `findSubstring` helpers in `x/data/sharding` or `x/ai/provider`
  — those are in different packages and out of scope.

Files:
- `contract/error_handling_test.go`

Tests:
- `go test -timeout 20s ./contract/...`

Docs Sync: —

Done Definition:
- `contains` and `findSubstring` do not exist in `contract/error_handling_test.go`.
- The `TestFormatError` assertion uses `strings.Contains`.
- Tests pass.

Outcome:
