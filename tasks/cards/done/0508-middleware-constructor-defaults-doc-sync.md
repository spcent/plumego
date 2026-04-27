# Card 0508

Milestone:
Recipe: specs/change-recipes/add-middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files: middleware/module.yaml; docs/modules/middleware/README.md; middleware/conformance/static_checks_test.go; tasks/cards/done/0508-middleware-constructor-defaults-doc-sync.md
Depends On: 2217

Goal:
Record and enforce the middleware constructor/defaults conventions that the audit found were inconsistent across packages.

Scope:
- Sync the middleware module primer with the current constructor split: legacy explicit names remain stable, new configurable packages prefer `Middleware(Config)` or documented package-specific canonical names.
- Add or tighten static conformance checks for direct `http.Error`, business DTO binding, and config types lacking a default path where the rule is practical.
- Document the remaining intentional exceptions so future cleanup does not fragment into one-off constructor families.

Non-goals:
- Do not rename exported middleware constructors in this card.
- Do not add compatibility wrappers or deprecations.
- Do not edit unrelated module docs.

Files:
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
- `middleware/conformance/static_checks_test.go`

Tests:
- `go test -timeout 20s ./middleware/conformance`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- Required: `docs/modules/middleware/README.md` and `middleware/module.yaml`.

Done Definition:
- The primer explains the intentional constructor/default exceptions and preferred new-code rule.
- Conformance tests protect the highest-signal middleware style rules without flagging intentional stable API names.
- Targeted middleware tests and vet pass.

Outcome:
- Documented preferred new configurable middleware constructor/default paths and current stable exceptions.
- Updated `middleware/module.yaml` agent hints to separate new-code convention from existing stable constructor names.
- Added a conformance check for exported middleware `Config`/`Options` types that lack `WithDefaults`, `Default`, or `Default<Type>()`, with explicit allowlisted stable exceptions.
- Validation run: `go test -timeout 20s ./middleware/conformance`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
