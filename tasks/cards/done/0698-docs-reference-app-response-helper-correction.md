# 0698 - Correct Reference App Response Helper Guidance

State: done
Priority: P1
Primary module: website docs

## Goal

Fix the misleading "do not infer" language in the Reference App page so it
clearly discourages feature-specific response helper families and reinforces the
single canonical response and error write paths.

## Scope

- Correct the English and Chinese Reference App pages.
- Add a short anti-pattern example only if it improves clarity without expanding
  the page too much.
- Cross-link to the Contract Primer and Error Model where useful.

## Non-goals

- Do not change Go code.
- Do not introduce a new response helper API.
- Do not rewrite the full Reference App page.

## Page Revision Checklist

| Target page | Modification goal | Acceptance points |
| --- | --- | --- |
| `website/src/content/docs/docs/reference-app.mdx` | Correct the row that currently implies feature-specific response helper families are encouraged. | The row explicitly says feature-specific response helper families are not encouraged; the positive guidance remains `contract.WriteResponse` and `contract.WriteError`. |
| `website/src/content/docs/zh/docs/reference-app.mdx` | Apply the same correction in Chinese. | Chinese wording clearly discourages per-feature response helper families. |
| `website/src/content/docs/docs/modules/contract.mdx` | Optionally add or tighten a sentence that `contract` owns the canonical write path. | The Contract Primer remains the source for the response/error contract and has no conflicting guidance. |
| `website/src/content/docs/zh/docs/modules/contract.mdx` | Mirror any Contract Primer wording change. | Localized wording stays consistent with the English page. |
| `website/src/content/docs/docs/concepts/error-model.mdx` | Optionally add a cross-reference from error construction to Reference App usage. | The page still treats `NewErrorBuilder` and `WriteError` as the single error path. |
| `website/src/content/docs/zh/docs/concepts/error-model.mdx` | Mirror any localized cross-reference. | Chinese copy reinforces the same single-path guidance. |

## Files

- `website/src/content/docs/docs/reference-app.mdx`
- `website/src/content/docs/zh/docs/reference-app.mdx`
- Optional: `website/src/content/docs/docs/modules/contract.mdx`
- Optional: `website/src/content/docs/zh/docs/modules/contract.mdx`
- Optional: `website/src/content/docs/docs/concepts/error-model.mdx`
- Optional: `website/src/content/docs/zh/docs/concepts/error-model.mdx`

## Tests

- `cd website && pnpm check:docs-api` - passed
- `cd website && pnpm check` - passed
- `cd website && pnpm build` - passed

## Planned Tests

- `cd website && pnpm check:docs-api`
- `cd website && pnpm check`

## Docs Sync

This is a documentation correction. Keep English and Chinese pages aligned.

## Done Definition

- No Reference App page suggests feature-specific response helper families are
  encouraged.
- The positive guidance points to `contract.WriteResponse`,
  `contract.WriteError`, and `contract.NewErrorBuilder`.
- Related contract/error pages do not contain conflicting wording.
