# Changelog

All notable changes to plumego will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Improved documentation for constructor functions that may panic
- All code formatted with `gofmt` for consistency

### Fixed
- Data race in `store/db/sharding/config/watcher_test.go` test
- Added missing package-level documentation for 17 submodules

### Documentation
- Added comprehensive package documentation for security, storage, and network modules
- Created `docs/ERROR_NAMING_STANDARD.md` outlining error naming conventions
- Created `docs/TENANT_PACKAGE_STATUS.md` documenting experimental multi-tenancy status
- Marked `tenant/` package as experimental with clear limitations

## [1.0.0-rc.1] - 2026-02-02

### Added
- Initial release candidate for v1.0
- Complete HTTP toolkit built on standard library
- Trie-based router with path parameters
- Comprehensive middleware system
- WebSocket hub with JWT authentication
- Webhook handling (GitHub, Stripe) with signature verification
- In-process pub/sub for event distribution
- Task scheduling with cron and retry support
- Embedded KV storage with WAL
- Static frontend serving
- Read/write splitting for databases
- Horizontal sharding support
- In-memory caching with TTL and LRU
- Rate limiting and abuse protection
- Security headers and input validation
- Metrics collection (Prometheus, OpenTelemetry)
- Health check endpoints
- Graceful shutdown with connection draining

### Breaking Changes
None - First release

---

## Release Notes

### v1.0.0-rc.1 - Release Candidate

This is the first release candidate for plumego v1.0. The project is feature-complete
and ready for community feedback.

**Highlights:**
- 🚀 Zero dependencies beyond Go standard library
- 🔧 Standard library-first design (`net/http`, `context`, `http.Handler`)
- 🎯 Explicit lifecycle management
- 🧩 Composable architecture
- 📦 Embedded-friendly (not a framework)

**Module Overview:**

| Module | Description | Status |
|--------|-------------|--------|
| `core/` | Application lifecycle and DI container | ✅ Stable |
| `router/` | HTTP routing with path parameters | ✅ Stable |
| `middleware/` | Request processing chain | ✅ Stable |
| `contract/` | Request context and error types | ✅ Stable |
| `config/` | Environment variable loading | ✅ Stable |
| `health/` | Health check endpoints | ✅ Stable |
| `security/` | JWT, password hashing, validation | ✅ Stable |
| `store/` | Database, cache, and KV storage | ✅ Stable |
| `net/` | HTTP client, WebSocket, webhooks | ✅ Stable |
| `pubsub/` | In-process message broker | ✅ Stable |
| `scheduler/` | Task scheduling and cron | ✅ Stable |
| `frontend/` | Static file serving | ✅ Stable |
| `tenant/` | Multi-tenancy support | ⚠️  Experimental |

**What's Stable:**
- Core HTTP routing and middleware
- Authentication and security features
- Database connectivity and sharding
- Caching and storage
- WebSocket and webhook handling
- Pub/sub messaging
- Task scheduling

**What's Experimental:**
- Multi-tenancy (`tenant/` package) - API may change

**Migration Guide:**

This is the first release - no migration needed.

**Known Limitations:**
- Multi-tenancy lacks middleware integration
- Some advanced pub/sub features have lower test coverage
- Performance benchmarks to be published separately

**Testing:**
- ✅ All tests pass: `go test ./...`
- ✅ Race detector clean: `go test -race ./...`
- ✅ Static analysis clean: `go vet ./...`
- ✅ Average test coverage: ~70%

**Requirements:**
- Go 1.24+ (see `go.mod`)
- No external dependencies for core functionality
- Optional: PostgreSQL/MySQL for database features
- Optional: Redis for distributed caching

**Getting Started:**

```bash
# Install
go get github.com/spcent/plumego

# Run reference example
cd examples/reference
go run .
```

**Documentation:**
- `README.md` - Project overview and quick start
- `CLAUDE.md` - Development guide and architecture
- `examples/` - Full-featured example applications
- `docs/` - Additional documentation and guides

**Feedback Welcome:**

This is a release candidate. We welcome feedback on:
- API design and ergonomics
- Documentation clarity
- Missing features
- Performance characteristics
- Real-world use cases

Please open issues at: https://github.com/spcent/plumego/issues

**Next Steps:**

After community feedback period:
- Address critical issues
- Finalize v1.0.0 release
- Publish performance benchmarks
- Expand example applications
- Improve documentation based on questions

---

## Versioning Policy

- **Major version (1.x.x)**: Breaking API changes
- **Minor version (x.1.x)**: New features, backward compatible
- **Patch version (x.x.1)**: Bug fixes, backward compatible

**Stability Guarantees:**

- v1.0.0+: Stable API with semantic versioning
- Experimental features clearly marked
- Deprecation notices in minor versions before removal
- Security fixes backported to recent versions

**Support:**

- Latest minor version: Full support
- Previous minor: Security fixes only
- Older versions: Community support

---

## Contributors

Thanks to all contributors who made this release possible!

Special thanks to the Go community for excellent standard library foundations.

---

[Unreleased]: https://github.com/spcent/plumego/compare/v1.0.0-rc.1...HEAD
[1.0.0-rc.1]: https://github.com/spcent/plumego/releases/tag/v1.0.0-rc.1
