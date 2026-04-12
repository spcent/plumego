# 0945 · x/websocket — sentinel errors in security.go and validation.go

## Status
active

## Problem
`x/websocket/errors.go` centralises most connection/hub/auth errors, but **eight more**
sentinel vars are spread across two other files:

```
security.go:76  ErrWeakJWTSecret
security.go:79  ErrWeakRoomPassword
security.go:82  ErrInvalidWebSocketKey
security.go:85  ErrInvalidConfig

validation.go:13  ErrInvalidUTF8
validation.go:16  ErrControlCharacters
validation.go:19  ErrMessageTooLong
validation.go:22  ErrEmptyMessage
```

Additionally, `server.go` returns inline `errors.New("websocket server misconfigured: …")`
strings (five occurrences) that are not testable via `errors.Is`.

A developer searching for all websocket sentinel errors must open three files. The
pattern breaks the expectation established by `errors.go`.

## Fix
1. Move all 8 sentinel vars from `security.go` and `validation.go` into `errors.go`.
   Group them with comments (`// Security`, `// Validation`).
2. Promote the five inline `errors.New("websocket server misconfigured: …")` strings in
   `server.go` to named sentinels in `errors.go`:
   ```go
   ErrNilHub            = errors.New("websocket: hub is nil")
   ErrNilAuthenticator  = errors.New("websocket: authenticator is nil")
   ErrNegativeQueueSize = errors.New("websocket: queue size cannot be negative")
   ErrInvalidSendBehavior = errors.New("websocket: invalid send behavior")
   ErrNegativeReadLimit = errors.New("websocket: read limit cannot be negative")
   ```
3. Also move `ErrWeakJWTSecret`, `ErrInvalidWebSocketKey` from `security.go` —
   after moving, those files become pure logic, no sentinel declarations.

## Verification
```bash
go build ./x/websocket/...
go test ./x/websocket/...
grep -n "^var Err" x/websocket/security.go x/websocket/validation.go  # should be empty
grep -n "errors\.New" x/websocket/server.go                             # should be empty
```

## Affected files
- `x/websocket/errors.go` (add 13 sentinels)
- `x/websocket/security.go` (remove 4 var declarations)
- `x/websocket/validation.go` (remove 4 var declarations)
- `x/websocket/server.go` (replace inline errors.New with sentinel refs)

## Severity
Low-Medium — developer experience; the inline validation errors in server.go are
untestable via errors.Is, making config-error handling fragile
