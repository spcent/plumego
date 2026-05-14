# Card 1400

Milestone: M-004
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: log
Owned Files:
- log/glog.go
- log/glog_writer.go
- log/glog_rotation.go
- log/glog_lifecycle.go
- log/glog_test.go
Depends On:
- 1394

Goal:
Mechanically split the stable `log` glog backend into smaller implementation files without changing behavior.

Scope:
- Keep public logging interfaces, constructor behavior, and field semantics unchanged.
- Move writer/cache helpers into a focused file.
- Move rotation and cleanup helpers into a focused file.
- Move close/flush lifecycle helpers into a focused file.
- Preserve current tests and add only narrow regression coverage if a move exposes a gap.

Non-goals:
- Do not add package-global singleton registration to the stable path.
- Do not add export backends, aggregation, or observability pipeline wiring.
- Do not change text/JSON field formatting.

Files:
- log/glog.go
- log/glog_writer.go
- log/glog_rotation.go
- log/glog_lifecycle.go
- log/glog_test.go

Tests:
- go test -race -timeout 60s ./log
- go test -timeout 20s ./log
- go vet ./log

Docs Sync:
- None expected unless implementation comments become misleading.

Done Definition:
- `log` public behavior and tests remain unchanged.
- `glog.go` review radius is reduced through mechanical file split.
- Target checks pass.

Outcome:
