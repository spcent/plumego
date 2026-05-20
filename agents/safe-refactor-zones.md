# Safe Refactor Zones

Per-package risk table. Use this before deciding whether a change needs review.
Based on `module.yaml:change_risks` and `review_checklist` fields.

---

## Risk Legend

- **Safe**: Agent can change without review; tests must still pass
- **Review required**: Make the change, but flag it for human review before merge
- **Human required**: Do not proceed without explicit approval in the task card

---

## Stable Roots

| Package | Safe | Review required | Human required |
|---|---|---|---|
| `contract` | Error message text, test files, doc comments, internal variable names | New exported type, new error code, new context accessor pair | Remove any exported symbol, change JSON wire format, change error HTTP status mapping |
| `core` | Test files, doc comments, error message text | Lifecycle order, middleware attachment, route registration error handling | Public API removal, startup/shutdown behavior change, new constructor parameter |
| `router` | Test files, doc comments, cache size constants | New route group method, param syntax, wildcard behavior | Route matching algorithm, exported API removal, RouteInfo field removal |
| `middleware` | Individual middleware option defaults, test files, doc comments | New middleware sub-package, middleware ordering, option field addition | Middleware removal, `next` call pattern change, transport behavior visible to callers |
| `security` | Test files, doc comments | New security check, new algorithm flag | Algorithm replacement, token format change, timing-safe comparison removal |
| `store` | Test files, in-memory implementation details, doc comments | New interface method (requires updating all implementations) | Interface method removal, storage contract change |
| `health` | Test files, doc comments, check description strings | New health check type | Health model field removal, status mapping change |
| `log` | Test files, doc comments, field name defaults | New log field, new backend method | Logger interface method removal, default log format change |
| `metrics` | Test files, doc comments, noop implementation details | New metric type (add to interface + all implementations) | Collector interface method removal |

---

## Extension Roots (x/*)

| Family | Safe | Review required | Human required |
|---|---|---|---|
| `x/tenant` | Test files, error messages, quota default values | Quota calculation logic, session TTL behavior, policy chain order | Tenant isolation semantics, store interface change |
| `x/ai` | Test files, provider option defaults, doc comments | New provider adapter, streaming frame handling | Token budget calculation, embedding vector contract |
| `x/observability` | Test files, devtools UI strings, metric names | New exporter type, OTel attribute mapping | Prometheus scrape format, trace propagation format |
| `x/gateway` | Test files, proxy timeout defaults, doc comments | New protocol adapter, load-balancing algorithm | Routing table format, upstream selection contract |
| `x/rest` | Test files, CRUD handler defaults | New resource method, spec field | ResourceSpec interface change |
| `x/data` | Test files, migration step details | New adapter implementation | Store interface contract |
| `x/websocket` | Test files, hub room defaults | New hub event type | Frame encoding, auth callback contract |
| `x/messaging` | Test files, retry delay defaults | New message envelope field | Queue durability contract, delivery guarantee semantics |

---

## Always Safe Everywhere

- Test files (`*_test.go`) — unless they test public API freeze behavior
- Doc comments that only describe existing behavior
- Internal unexported variable names
- Log message text (not structured field names)
- Error message text (not error codes or HTTP status codes)
- Benchmark parameters

## Never Safe Without Review

- Any exported symbol removal (even if unused in this repo)
- Any `go.mod` change
- Any file in `internal/checks/` (affects all gate results)
- Any `specs/*.yaml` file (affects all agent decisions)
- `AGENTS.md`, `docs/CANONICAL_STYLE_GUIDE.md`, `docs/architecture/`
