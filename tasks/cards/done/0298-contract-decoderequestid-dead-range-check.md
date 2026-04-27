# Card 0298

Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/request_id_generation.go`
Depends On:

Goal:
- Remove the unreachable range check in `DecodeRequestID` and the sentinel error it was meant to protect.

Problem:
- `DecodeRequestID` (`request_id_generation.go:185`) extracts the random component `r` via:
  ```go
  r = int((v >> shiftRand) & uint64(randMask))
  ```
  where `randMask = (1 << randBits) - 1 = 16383`.
- Immediately after, it performs:
  ```go
  if r < 0 || r > randMask {
      return 0, 0, 0, errOutOfRange
  }
  ```
- This check can never be true:
  - `r < 0` is impossible: the AND with `uint64(randMask)` produces a non-negative value; casting to `int` on any supported platform keeps it non-negative for a 14-bit value.
  - `r > randMask` is impossible: the AND operation guarantees the result is in `[0, randMask]` by definition.
- As a result, `errOutOfRange` (`request_id_generation.go:37`) is unreachable dead code — it can never be returned from any execution path.
- The check creates a false impression that `DecodeRequestID` validates the random field range, adding complexity and a misleading error path.

Scope:
- Delete the `if r < 0 || r > randMask` block from `DecodeRequestID`.
- Delete the `errOutOfRange` sentinel error declaration (`request_id_generation.go:37`).
- Verify that no other code references `errOutOfRange`.

Non-goals:
- Do not change the decoding logic or other validation checks in `DecodeRequestID`.
- Do not remove `errInvalidBase62` (it is reachable and serves a real purpose).

Files:
- `contract/request_id_generation.go`
- `contract/request_id_generation_test.go` (remove any test of the impossible path)

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None required.

Done Definition:
- The `if r < 0 || r > randMask` block is deleted.
- `errOutOfRange` is deleted.
- No remaining reference to `errOutOfRange` in the package.
- All tests pass.

Outcome:
- Pending.
