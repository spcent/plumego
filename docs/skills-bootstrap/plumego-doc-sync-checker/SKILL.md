---
name: plumego-doc-sync-checker
description: Keep Plumego documentation and config examples synchronized with behavior changes. Use when API, defaults, security behavior, env vars, or startup/shutdown semantics change.
---

# Objective
Prevent code-doc drift for user-facing behavior.

# Sync Targets
- README.md
- README_CN.md
- AGENTS.md
- CLAUDE.md
- env.example

# Workflow
1. Detect whether change affects:
   - public API
   - config/env/defaults
   - security behavior
   - lifecycle semantics
2. For affected areas, patch required sync targets.
3. Keep bilingual docs consistent in scope and meaning.
4. Avoid speculative docs; document only implemented behavior.

# Output Contract
Return:
- Behavior Changes
- Files Updated
- Remaining Doc Gaps
