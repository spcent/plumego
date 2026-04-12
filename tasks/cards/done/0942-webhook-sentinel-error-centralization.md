# 0942 · x/webhook — sentinel error scatter and ErrInvalidHex triplication

## Status
active

## Problem
Sentinel error variables are spread across six files in the `x/webhook` package, with
three distinct vars that share the **identical message text**:

| var | file | text |
|-----|------|------|
| `ErrGitHubInvalidEncoding` | `inbound_github.go` | `"invalid hex encoding"` |
| `ErrStripeInvalidEncoding` | `inbound_stripe.go` | `"invalid hex encoding"` |
| `ErrInvalidHex` | `outbound_signer.go` | `"invalid hex encoding"` |

Because they are separate `errors.New` allocations, `errors.Is(ErrGitHubInvalidEncoding,
ErrStripeInvalidEncoding)` returns `false`. Callers trying to handle hex-encoding failures
must enumerate all three, and a new provider would add a fourth.

Additional scattered sentinels:
- `ErrQueueFull` — `outbound_queue.go`
- `ErrNotFound` — `outbound_store_mem.go` (dangerously generic name, will shadow on import)

There is also no `errors.go` file in the package; `inbound_errors.go` only covers the
structured inbound `VerifyError` type/codes.

## Fix
1. Create `x/webhook/errors.go` with **all** exported sentinel errors:
   ```go
   var (
       // outbound / delivery
       ErrQueueFull    = errors.New("webhook: delivery queue full")
       ErrWebhookDisabled = errors.New("webhook: endpoint is disabled")

       // inbound verification
       ErrInvalidHexEncoding = errors.New("webhook: invalid hex encoding")
       ErrGitHubSignature    = errors.New("webhook: invalid github signature")
       ErrGitHubMissingHeader = errors.New("webhook: missing github signature header")
       ErrStripeSignature    = errors.New("webhook: invalid stripe signature")
       ErrStripeMissingHeader = errors.New("webhook: missing stripe signature header")
       ErrStripeInvalidTimestamp = errors.New("webhook: invalid timestamp in signature")
       ErrStripeExpired      = errors.New("webhook: signature expired")

       // store
       ErrTargetNotFound = errors.New("webhook: target not found")
   )
   ```
2. Delete `ErrGitHubInvalidEncoding`, `ErrStripeInvalidEncoding`, `ErrInvalidHex` from their
   current locations; replace with the single `ErrInvalidHexEncoding`.
3. Rename `outbound_store_mem.go:ErrNotFound` → `ErrTargetNotFound` (avoid generic-name
   shadowing; matches the `ErrTargetNotFound` sentinel in errors.go).
4. Move remaining sentinel vars (`ErrGitHubSignature`, `ErrStripeSignature`, etc.) to
   `errors.go`; remove them from inbound_github.go / inbound_stripe.go / outbound_signer.go.
5. Update all `case CodeInvalidEncoding: return nil, ErrGitHub/StripeInvalidEncoding` sites
   to use `ErrInvalidHexEncoding`.
6. Update tests accordingly.

## Verification
```bash
go build ./x/webhook/...
go test ./x/webhook/...
grep -rn "ErrGitHubInvalidEncoding\|ErrStripeInvalidEncoding\|ErrInvalidHex\b" ./x/webhook/ # should be empty
```

## Affected files
- `x/webhook/errors.go` (new)
- `x/webhook/inbound_github.go`
- `x/webhook/inbound_stripe.go`
- `x/webhook/outbound_signer.go`
- `x/webhook/outbound_queue.go`
- `x/webhook/outbound_store_mem.go`
- `x/webhook/outbound_service.go` (ErrWebhookDisabled, if promoted)

## Severity
Medium — three identical-message sentinels break `errors.Is` across inbound paths;
`ErrNotFound` generic name risks import-dot shadowing
