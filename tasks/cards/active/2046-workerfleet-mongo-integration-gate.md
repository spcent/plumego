# Card 2046

Milestone:
Recipe: specs/change-recipes/add-acceptance-tests.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- Makefile
- reference/workerfleet/docs/storage.md
- reference/workerfleet/README.md
- tasks/cards/active/2046-workerfleet-mongo-integration-gate.md
Depends On: 2045

Goal:
Expose a clear optional Mongo integration test gate for workerfleet production storage behavior.

Scope:
Add a repo-native make target that runs workerfleet Mongo integration tests when `WORKERFLEET_MONGO_TEST_URI` is set and document the command.

Non-goals:
- Do not make the default `make gates` require a local MongoDB.
- Do not add Docker, testcontainers, or new third-party test dependencies.
- Do not change Mongo runtime behavior.

Files:
- Makefile
- reference/workerfleet/README.md
- reference/workerfleet/docs/storage.md
- tasks/cards/active/2046-workerfleet-mongo-integration-gate.md

Acceptance Tests:
- Make target exists and has a clear missing-URI failure message.

Tests:
- `make workerfleet-mongo-test` without URI exits with an actionable message.

Docs Sync:
- reference/workerfleet/README.md
- reference/workerfleet/docs/storage.md

Validation:
- make workerfleet-mongo-test
- cd reference/workerfleet && go test -timeout 30s ./internal/platform/store/mongo
- git diff --check

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
