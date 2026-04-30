# Card 0703

Milestone: M-001
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: cmd/plumego
Owned Files: cmd/plumego/commands/new.go, cmd/plumego/commands/cli_e2e_test.go, cmd/plumego/internal/scaffold/scaffold.go
Depends On: 0702-cli-template-truth-audit

Goal:
Align `plumego new --template` validation, help text, and dry-run behavior with the template truth matrix from Card 0702.

Scope:
Update the `new` command template allowlist and user-facing help/error text so every supported template is accepted and every unsupported template fails clearly.

Non-goals:
Do not change generated file contents unless template dispatch currently prevents an implemented scaffold from running.
Do not update external docs in this card.
Do not add dependencies.

Files:
cmd/plumego/commands/new.go
cmd/plumego/commands/cli_e2e_test.go
cmd/plumego/internal/scaffold/scaffold.go

Tests:
cd cmd/plumego && go test -timeout 20s ./...
cd cmd/plumego && go run . new --template rest-api --dry-run trust-check
cd cmd/plumego && go run . new --template invalid-template-name --dry-run trust-check

Docs Sync:
None in this card. Documentation follows in Card 0704 after behavior is fixed.

Done Definition:
All templates selected by Card 0702 are accepted by `plumego new`.
Invalid templates return non-zero with a clear error.
CLI tests cover at least one accepted scenario template and one invalid template.

Outcome:
Completed.

Changes:

- Added scaffold-backed scenario templates to `plumego new --template` validation.
- Updated `new` help/error template text to use the same supported-template list.
- Added CLI e2e coverage for all scenario template dry-run paths and invalid-template failure.

Validation:

- `cd cmd/plumego && go test -timeout 20s ./...` passed.
- `cd cmd/plumego && go run . new --template rest-api --dry-run trust-check` passed.
- `cd cmd/plumego && go run . new --template invalid-template-name --dry-run trust-check` failed as expected with exit status 3 and the complete valid-template list.

