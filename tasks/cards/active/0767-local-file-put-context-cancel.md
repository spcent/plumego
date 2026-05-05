# Card 0767

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/file
Owned Files: x/data/file/local.go, x/data/file/local_test.go
Depends On:

Goal:

Make LocalStorage.Put honor context cancellation during upload reads and hash/write work.

Scope:

- Check ctx before starting local write work.
- Wrap the upload reader so cancellation interrupts long copies.
- Return errors with the existing file operation context.
- Add focused cancellation coverage.

Non-goals:

- Adding resumable uploads.
- Changing local file path layout.
- Changing S3 cancellation behavior.

Files:

- x/data/file/local.go
- x/data/file/local_test.go

Tests:

- go test -race -timeout 60s ./x/data/file/...
- go test -timeout 20s ./x/data/file/...
- go vet ./x/data/file/...

Docs Sync:

- Not required; context-aware behavior follows the existing method signature.

Done Definition:

- A canceled context stops Local Put before completing disk write.
- Returned errors remain classifiable through the file error wrapper.
- Module tests and vet pass.

Outcome:

