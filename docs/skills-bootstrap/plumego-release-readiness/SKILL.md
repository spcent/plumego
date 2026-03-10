---
name: plumego-release-readiness
description: Evaluate Plumego release readiness for rc/v1 with a strict go/no-go checklist. Use before tagging, release candidate promotion, or final v1 decision.
---

# Objective
Produce a deterministic release go/no-go decision.

# Checklist
1. Quality gates all pass:
   - go test -timeout 20s ./...
   - go vet ./...
   - gofmt -w .
2. API surface freeze status is explicit.
3. Experimental modules are clearly labeled and documented.
4. Security-critical paths include negative tests.
5. Docs/config examples are synchronized.
6. No unresolved high-severity regressions.

# Workflow
1. Gather evidence per checklist item.
2. Mark each item: PASS / WARN / FAIL.
3. Compute final decision:
   - GO: no FAIL
   - NO-GO: any FAIL
4. Provide the shortest path from NO-GO to GO.

# Output Contract
Return:
- Decision: GO or NO-GO
- Blocking Items
- Fix-First Order
- Final Verification Commands
