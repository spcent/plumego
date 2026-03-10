---
name: plumego-task-slicer
description: Split large Plumego requests into minimal reversible task cards under token limits. Use when user asks roadmap, phased iteration, or a request spans multiple modules and needs decomposition before coding.
---

# Objective
Decompose work into small cards that fit one Codex turn.

# Workflow
1. Extract objective, constraints, module boundary, and expected outcome.
2. Split into cards with strict limits:
   - one primary module
   - up to 5 files
   - up to 3 validation commands
   - reversible in one commit
3. Prioritize by risk:
   - P0: correctness/security/regression
   - P1: API/style consistency
   - P2: docs/examples
4. For each card, output exactly:
   - Goal
   - Scope
   - Non-goals
   - Files
   - Tests
   - Docs Sync
   - Done Definition
5. Keep the first response compact; expand only when asked.

# Output Contract
Always return:
- Now: one card to execute next
- Queue: next 3 cards
- Risks: top 3 risks and mitigations
