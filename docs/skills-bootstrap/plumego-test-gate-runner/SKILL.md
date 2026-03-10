---
name: plumego-test-gate-runner
description: Run Plumego tests efficiently with token-aware staging and required quality gates. Use after any code change or when preparing merge/release confidence checks.
---

# Objective
Minimize test cost while preserving confidence.

# Workflow
1. Infer change type from touched files:
   - routing, middleware, security, tenant, store, docs-only
2. Run staged validation:
   - Stage A: targeted tests for changed modules
   - Stage B: required full gates
3. Required full gates:
   - go test -timeout 20s ./...
   - go vet ./...
   - gofmt -w .
4. Add change-specific checks:
   - routing: static/param/group/reverse routing tests
   - middleware: ordering/error-path tests
   - security: invalid token/signature negatives
   - tenant: quota/policy/isolation tests
5. Summarize only failing signals first, then concise pass summary.

# Output Contract
Return:
- Commands Run
- Failures (if any)
- Risk Assessment
- Merge Readiness
