# Card 0992

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P1
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/helpers.go
- x/data/file/s3.go
- x/data/file/s3_test.go
Depends On:
- tasks/cards/done/0978-file-metadata-tenant-scoped-dedupe.md

Goal:
Apply the same tenant/path safety rules to S3 object keys that Local storage already enforces.

Scope:
- Validate S3 tenant id before composing object keys.
- Validate S3 path and prefix inputs for Get/Delete/Exists/Stat/List/GetURL/Copy.
- Return store/file invalid-path errors consistently for unsafe inputs.

Non-goals:
- Do not add a new S3 client dependency.
- Do not change Local storage behavior except shared helper reuse if needed.

Files:
- x/data/file/helpers.go
- x/data/file/s3.go
- x/data/file/s3_test.go

Tests:
- go test -timeout 20s ./x/data/file

Docs Sync:
- docs/modules/store/README.md if backend path semantics change from the documented contract.

Done Definition:
- S3 rejects traversal and unsafe tenant/path inputs.
- Unsafe paths cannot be normalized into another namespace.
- Targeted tests pass.

Outcome:
S3 Put now validates tenant id before composing object keys, and S3 path-bearing operations reject unsafe object paths or prefixes before building requests. Added regression coverage for traversal, absolute paths, backslashes, and empty tenant ids.

Validation:
- go test -timeout 20s ./x/data/file
