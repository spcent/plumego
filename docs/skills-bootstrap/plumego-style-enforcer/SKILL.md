---
name: plumego-style-enforcer
description: Enforce Plumego canonical coding style for handlers, routes, JSON decode, middleware, and error response shape. Use when implementing/refactoring HTTP-facing code.
---

# Objective
Keep one obvious coding style across modules.

# Canonical Rules
1. Handler shape: func(http.ResponseWriter, *http.Request)
2. Explicit route registration: one method + path + handler per line
3. JSON decode path: json.NewDecoder(r.Body).Decode(...)
4. Single error write path: contract.WriteError with structured error code
5. Middleware shape: func(http.Handler) http.Handler
6. Constructor-based DI; no context service-locator

# Workflow
1. Scan changed code for style violations.
2. Apply minimal edits aligned with docs/CANONICAL_STYLE_GUIDE.md.
3. Preserve behavior; avoid introducing new helper families.
4. Add or adjust tests only where behavior changes.

# Output Contract
Return:
- Violations Found
- Fixes Applied
- Residual Style Risk
