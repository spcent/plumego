# Card 0071

Priority: P1

Goal:
- Remove dead code, consolidate overlapping abstractions, and split the
  overloaded `errors.go` into focused files so the contract package has a
  clear, navigable surface.

Problems identified (second-pass analysis):

### P1 — Structural issues

**P1-A: `errors.go` violates single responsibility (957 lines)**
Six orthogonal concerns live in one file: error type constants, HTTP error
construction, builder pattern, error wrapping, retry/panic utilities, and
metrics. These must be split into focused files.

**P1-B: Two parallel error-wrapping systems**
`ErrorChain` (lines 339-441) and `WrappedErrorWithContext` (lines 643-724)
both represent "error with context". Both have separate branches in
`IsRetryable()` and `GetErrorDetails()`. `WrapError()` only creates
`WrappedErrorWithContext`; `ErrorChain` is never created outside tests.
Consolidate: keep `WrappedErrorWithContext` / `WrapError` (externally used),
remove `ErrorChain`.

**P1-C: `ErrorHandler` is a zero-value wrapper class (error_utils.go)**
All 9 methods are one-liners that delegate to module-level functions. No
external callers. Remove the struct; elevate `SafeExecute` to a package-level
function matching `SafeExecuteWithResult`.

**P1-D: `CategoryForStatus` ↔ `HTTPStatusFromCategory` are non-invertible**
`CategoryForStatus(404)` → `CategoryClient`, but
`HTTPStatusFromCategory(CategoryClient)` → 400.
Fix `HTTPStatusFromCategory` to use `errorTypeLookup` (added in card 0505) so
round-trips are consistent.

**P1-E: `principalContextKeyVar` residual package-level variable (auth.go:148)**
`auth.go` still uses `var principalContextKeyVar principalContextKey` — the
old pattern that was eliminated from `context_core.go` and `trace.go` in
card 0509. Inline to `principalContextKey{}`.

### P2 — Dead code

**P3-A: `ErrorMetrics` is dead code (errors.go:625)**
The struct and `NewErrorMetrics()` are defined but never instantiated in
production code. Remove struct, constructor, and test coverage.

**P3-C: `Must`/`Must1`/`Must2` don't belong here (errors.go:911)**
Generic panic helpers have no relation to HTTP transport contracts. No
external callers. Remove.

### P3 — Style/correctness

**P3-D: `//Deprecated:` comment format not Go-1.19 compliant**
`context_stream.go` uses `// Deprecated:` mid-comment rather than as a doc
comment first line. Fix so gopls and staticcheck surface the deprecation.

Scope:
- Remove `ErrorMetrics` (struct + constructor + test)
- Remove `Must`/`Must1`/`Must2` (functions + test)
- Remove `ErrorHandler` struct; convert `SafeExecute` to package-level function
- Inline `principalContextKeyVar` → `principalContextKey{}`
- Fix `HTTPStatusFromCategory` to delegate to `errorTypeLookup`
- Fix `//Deprecated:` format in `context_stream.go`
- Extract `error_wrap.go`: move `WrappedErrorWithContext`, `ErrorContext`,
  `WrapError`, `WrapErrorf`, `NewWrappedError`, `IsRetryable`,
  `GetErrorDetails`, `FormatError`, `PanicToError`, `SafeExecute`,
  `SafeExecuteWithResult` out of `errors.go`
- Remove `ErrorChain` (no external callers); simplify `IsRetryable` and
  `GetErrorDetails` to not branch on it

Non-goals:
- Do not change the `APIError` wire format
- Do not remove `WrapError` / `WrappedErrorWithContext` (externally used)
- Do not change `ErrorCategory` / `ErrorType` constants

Files:
- `contract/errors.go`
- `contract/error_utils.go` → delete after evacuation
- `contract/error_wrap.go` → new file
- `contract/auth.go`
- `contract/context_stream.go`
- `contract/errors_test.go`
- `contract/error_handling_test.go`

Tests:
- `go test ./contract/...`
- `go test ./...` (verify no external breakage)
- `go vet ./...`

Done Definition:
- `errors.go` is ≤ 550 lines and covers only API error types + construction
- `error_wrap.go` covers all error wrapping/inspection utilities
- `error_utils.go` is deleted
- `ErrorHandler`, `ErrorMetrics`, `Must`/`Must1`/`Must2`, `ErrorChain` are gone
- `principalContextKeyVar` is inlined
- `HTTPStatusFromCategory` uses `errorTypeLookup`
- All tests pass; no external breakage
