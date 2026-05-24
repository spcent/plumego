# Card 2055

Milestone:
Recipe: specs/change-recipes/add-acceptance-tests.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- .github/workflows/quality-gates.yml
- Makefile
- reference/workerfleet/README.md
- reference/workerfleet/docs/storage.md
Depends On: 2054

Goal:
Make the optional workerfleet Mongo integration gate visible and runnable in CI without forcing default CI to require MongoDB.

Scope:
Add a CI opt-in path for `make workerfleet-mongo-test`, keep the default gate skip-safe without `WORKERFLEET_MONGO_TEST_URI`, and document the CI contract.

Non-goals:
- Do not add Docker services, testcontainers, or new third-party dependencies.
- Do not make default `make gates` require MongoDB.
- Do not store Mongo credentials in the repository.

Files:
- .github/workflows/quality-gates.yml
- Makefile
- reference/workerfleet/README.md
- reference/workerfleet/docs/storage.md

Acceptance Tests:
- Make target remains skip-safe without `WORKERFLEET_MONGO_TEST_URI`.
- CI workflow includes an explicit opt-in workerfleet Mongo gate step.

Tests:
- `make workerfleet-mongo-test` without URI.

Docs Sync:
- reference/workerfleet/README.md
- reference/workerfleet/docs/storage.md

Validation:
- make workerfleet-mongo-test
- git diff --check

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
