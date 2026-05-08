# Card 0835

Milestone: v1
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/error_codes.go
- x/messaging/api.go
- x/messaging/api_test.go
- x/resilience/circuitbreaker/middleware.go
- x/resilience/circuitbreaker/*_test.go
Depends On:
- 0726

Goal:
Remove extension-owned messaging, pub/sub, and resilience error codes from stable `contract`.

Scope:
- Enumerate all uses of each extension-owned `contract.Code*` symbol before editing.
- Move messaging error codes to the messaging transport owner.
- Move circuit breaker error codes to the resilience transport owner.
- Update affected tests and call sites.
- Re-run symbol searches to confirm the old `contract.Code*` names are gone.

Non-goals:
- Do not change response envelope shape.
- Do not redesign extension error taxonomies beyond moving ownership.
- Do not add deprecated wrappers in `contract`.

Files:
- contract/error_codes.go
- x/messaging/api.go
- x/messaging/api_test.go
- x/resilience/circuitbreaker/middleware.go
- x/resilience/circuitbreaker/*_test.go

Tests:
- go test -timeout 20s ./contract/... ./x/messaging/... ./x/resilience/circuitbreaker/...
- go build ./...
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/contract/README.md` only if the public surface list changes.

Done Definition:
- No Go call site references the removed extension-owned `contract.Code*` symbols.
- Extension packages own their local machine-readable codes.
- Targeted tests, build, and dependency checks pass.

Outcome:
- Enumerated extension-owned code symbols before editing.
- Moved messaging API error codes into `x/messaging`.
- Moved circuit breaker open-state error code into `x/resilience/circuitbreaker`.
- Removed messaging, pub/sub, and resilience code constants from stable `contract`.

Validation:
- rg -n --glob '*.go' 'contract\.CodeProviderError|contract\.CodeQuotaExceeded|contract\.CodeDuplicateMessage|contract\.CodeTaskExpired|contract\.CodeSendError|contract\.CodeEmptyBatch|contract\.CodeStatsError|contract\.CodeInvalidPayload|contract\.CodeInvalidMessage|contract\.CodeCircuitOpen' .
- go test -timeout 20s ./contract/... ./x/messaging/... ./x/resilience/circuitbreaker/...
- go build ./...
- go run ./internal/checks/dependency-rules
