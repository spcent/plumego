# Extension Stability Policy

This policy defines the status levels used by `x/*` extension families and their
sub-packages. It is the authoritative source referenced by
`docs/concepts/extension-maturity.md`.

## Status Values

### `experimental`

- API may change in any minor version without notice.
- Not recommended for production use without explicit project-level stabilization.
- May be promoted to `beta` once release evidence requirements are met (see
  `specs/extension-beta-evidence.yaml`).

### `beta`

- API is stable across the release refs cited in the beta evidence doc.
- Breaking changes require a deprecation notice in the same PR.
- Suitable for adoption with the understanding that edge cases may still be
  rough.
- May be promoted to `ga` once the module has two production release cycles
  with no exported API changes and full test coverage at the stable tier.

### `ga` (General Availability)

- API follows the same compatibility guarantee as stable roots.
- Breaking changes require a major version bump or explicit migration path.
- Only stable roots currently carry `ga` status; extensions reach `ga` via the
  promotion criteria in `specs/extension-maturity.yaml`.

## Promotion Criteria

| From | To | Requirements |
|---|---|---|
| experimental | beta | Two consecutive release refs with no exported API changes; owner sign-off; beta evidence recorded in `docs/evidence/extension/`. |
| beta | ga | Two production release cycles; full coverage at stable tier; boundary review sign-off; entry in `docs/evidence/stable-api/`. |

## Deprecation Within a Status Level

Deprecated symbols must be removed in the same PR that migrates their last
caller. No dead wrappers may remain at merge time. This applies at all maturity
levels.
