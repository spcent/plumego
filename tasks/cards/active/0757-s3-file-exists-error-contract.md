# Card 0757

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P1
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/s3.go
- x/data/file/s3_test.go
Depends On:

Goal:
Make S3 Exists report non-404 S3 errors instead of treating them as missing objects.

Scope:
- Keep 200 as exists and 404 as not found.
- Return a *file.Error for other response statuses and request/client failures.
- Add tests for 403 or 5xx responses.

Non-goals:
- Do not redesign S3 signing or URL construction.
- Do not add external S3 client dependencies.

Files:
- x/data/file/s3.go
- x/data/file/s3_test.go

Tests:
- go test ./x/data/file

Docs Sync:
- Not required unless public docs mention Exists error mapping.

Done Definition:
- Exists no longer returns false,nil for authorization or server failures.
- S3 file tests pass.

Outcome:

