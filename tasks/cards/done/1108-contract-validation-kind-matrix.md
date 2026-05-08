# Card 1108

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/active_cards_regression_test.go
- docs/modules/contract/README.md
Depends On:
- 0749

Goal:
Freeze `ValidateStruct` rule/type semantics for unsupported `min` and `max` kinds.

Scope:
- Add tests showing unsupported `min`/`max` target kinds are no-ops, not configuration errors.
- Document the rule/type support matrix for `required`, `email`, `min`, and `max`.
- Preserve current compatibility behavior.

Non-goals:
- Do not add new validation rules.
- Do not make unsupported kinds fail in this card.
- Do not add localization or cross-field validation.

Files:
- contract/active_cards_regression_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update validation support matrix.

Done Definition:
- Unsupported `min`/`max` behavior is executable and documented.
- Contract tests and vet pass.

Outcome:
- Added regression coverage proving `min` and `max` on unsupported kinds remain compatibility no-ops.
- Documented the retained `ValidateStruct` rule/type matrix and unsupported target behavior.
- Reconfirmed malformed rule arguments and unknown rules remain configuration errors.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
