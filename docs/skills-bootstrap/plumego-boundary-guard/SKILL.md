---
name: plumego-boundary-guard
description: Enforce Plumego module boundaries and architecture constraints. Use when reviewing planned or actual code changes, especially for core/router/middleware/security packages.
---

# Objective
Prevent cross-module responsibility leakage.

# Workflow
1. Classify touched files by module: core/router/middleware/contract/security/tenant/store/x/ai/utils.
2. Check boundary rules from AGENTS.md:
   - no routing behavior in core
   - no business logic in middleware
   - no persistence/business logic in utils
   - preserve net/http compatibility
3. Detect risky patterns:
   - hidden global side effects
   - context service-locator DI
   - new non-stdlib dependency in main module
4. Produce decision:
   - PASS
   - WARN (acceptable with mitigation)
   - FAIL (must refactor)
5. For WARN/FAIL, give minimal refactor path with file-level actions.

# Output Contract
Return:
- Boundary Status
- Violations
- Minimal Fix Plan
- Required Extra Tests
