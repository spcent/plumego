# Card 0924

Priority: P1
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/config.go`
Depends On:

Goal:
- Remove the three exported `*FromConfig` adapter types that wrap a `ConfigManager` to satisfy a sub-interface, since `InMemoryConfigManager` already implements all three sub-interfaces directly.
- Replace them with a documented pattern: embed the three provider interfaces in `ConfigManager` so any implementation must satisfy all of them.

Problem:
- `config.go` exports three wrapper structs (`QuotaConfigProviderFromConfig:61`, `PolicyConfigProviderFromConfig:78`, `RateLimitConfigProviderFromConfig:97`) whose sole purpose is to adapt a `ConfigManager` to `QuotaConfigProvider`, `PolicyConfigProvider`, and `RateLimitConfigProvider`.
- `InMemoryConfigManager` (`config.go:114`) already implements `GetTenantConfig`, `QuotaConfig`, `PolicyConfig`, and `RateLimitConfig` directly, meaning the adapters are unnecessary when using the provided implementation.
- The adapters are only useful if someone writes a custom `ConfigManager` that does not implement the sub-interfaces — a use case the package does not document, guide, or test.
- Three nearly-identical boilerplate wrapper types with no tests of their own add noise, create a false impression that they are required, and can silently produce bugs (e.g., wrapping `InMemoryConfigManager` in an adapter bypasses caching or direct method calls).
- The clean design is to embed the three provider interfaces in `ConfigManager` itself, so the interface contract is complete and no adapters are needed.

Scope:
- Add `QuotaConfigProvider`, `PolicyConfigProvider`, and `RateLimitConfigProvider` as embedded interfaces in `ConfigManager`.
- Delete `QuotaConfigProviderFromConfig`, `PolicyConfigProviderFromConfig`, and `RateLimitConfigProviderFromConfig` struct types and their methods.
- Verify `InMemoryConfigManager` already satisfies the expanded `ConfigManager` interface (it does).
- Remove any test or usage of the adapter types.

Non-goals:
- Do not change `InMemoryConfigManager` behavior.
- Do not add a new constructor or factory.
- Do not touch any package outside `x/tenant/core`.

Files:
- `x/tenant/core/config.go`
- `x/tenant/core/config_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- `ConfigManager` interface embeds `QuotaConfigProvider`, `PolicyConfigProvider`, and `RateLimitConfigProvider`.
- The three `*FromConfig` adapter structs and their methods are deleted.
- `InMemoryConfigManager` satisfies the combined interface without changes.
- All tests pass.

Outcome:
- Pending.
