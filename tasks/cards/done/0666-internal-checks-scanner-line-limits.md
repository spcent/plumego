# Card 0666

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/checks
Owned Files: internal/checks/checkutil/checkutil.go, internal/checks/checkutil/controlplane.go, internal/checks/checkutil/checkutil_test.go, internal/checks/extension-api-snapshot/main.go, internal/checks/extension-beta-evidence/main.go, internal/checks/extension-maturity/main.go
Depends On:

Goal:
Prevent internal repository checks from failing on valid long single-line control-plane data.

Scope:
- Add one shared line scanner helper in `internal/checks/checkutil` with an explicit larger token limit.
- Replace direct `bufio.NewScanner` use in internal checks that read repository metadata, manifests, evidence docs, or snapshots.
- Add focused regression coverage for at least one long-line path.

Non-goals:
- Do not replace the existing lightweight YAML/Markdown parsing approach.
- Do not change check rules or reported violation text.

Files:
- internal/checks/checkutil/checkutil.go
- internal/checks/checkutil/controlplane.go
- internal/checks/checkutil/checkutil_test.go
- internal/checks/extension-api-snapshot/main.go
- internal/checks/extension-beta-evidence/main.go
- internal/checks/extension-maturity/main.go

Tests:
- go test -timeout 20s ./internal/checks/...
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; check robustness only.

Done Definition:
- Internal checks no longer rely on the scanner default 64 KiB token limit.
- Regression coverage proves a long valid line is accepted.

Outcome:
Done. Added `checkutil.NewLineScanner` with a 1 MiB token limit and routed
internal repository check scanners through it. Added regression coverage for a
long baseline line.

Validation:
- go test -timeout 20s ./internal/checks/...
- go test -timeout 20s ./internal/...
- go vet ./internal/...
