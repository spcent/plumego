# Card XXXX

Milestone:
Recipe: specs/change-recipes/<recipe>.yaml
Priority:
State: active
Primary Module:
Owned Files:
Depends On:

## Goal

<!-- One sentence. Observable end-state, not steps. -->

## Scope

<!-- Exact files or packages in scope. Max 5 files. -->

## Non-goals

<!-- Hard stops. Agent must not cross these. -->

## Files

<!-- Max 5 files. List each on its own line. -->

## Acceptance Tests

<!-- Pre-written failing test function(s) that define done.
     Format: <file>: <TestFunctionName>
     Example:
       middleware/timeout_test.go: TestTimeoutHeaderPresent
       contract/errors_test.go: TestWriteErrorCode
     Leave blank if the card has no behavior change. -->

## Tests

<!-- Additional tests beyond acceptance tests (edge cases, negative paths). -->

## Docs Sync

<!-- Docs files to update if behavior/API/config/security changes. -->

## Validation

<!-- Exact commands that must exit 0.
     Minimum: go test -race -timeout 60s ./<module>/... -->

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

<!-- Agent fills this after completion: what changed and why. -->
