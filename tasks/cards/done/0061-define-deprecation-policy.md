# Card 0061

Priority: P1

Goal:
- Define a formal deprecation policy for stable library roots and extension
  packages so maintainers and consumers have a clear, written contract for how
  API evolution will be managed post-v1.

Scope:
- Write `docs/DEPRECATION.md` covering:
  - Which packages the policy applies to (stable roots vs. `x/*`).
  - The v1 compatibility promise for stable roots.
  - A four-step deprecation process (mark, document, maintain, remove).
  - Extension package evolution rules (Experimental; no freeze).
  - What is not considered a breaking change.
  - Governance (owner approval requirements by risk level).
  - Cross-links to `<module>/module.yaml`, `docs/ROADMAP.md`, `AGENTS.md`.

Non-goals:
- Do not deprecate any existing symbol in this card.
- Do not change any code in this card.
- Do not extend the policy to CLI output format or env-var defaults.

Files:
- `docs/DEPRECATION.md` (new)

Tests:
- `go run ./internal/checks/reference-layout` (verifies no unexpected top-level
  changes)

Docs Sync:
- `docs/ROADMAP.md`: reference the new policy document under Phase 6 exit
  criteria.

Done Definition:
- `docs/DEPRECATION.md` exists and is linked from ROADMAP Phase 6 notes.
- The four-step process, extension exemption, and governance section are all
  present.
