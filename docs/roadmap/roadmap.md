# Plumego Roadmap

This document outlines the planned releases, stability commitments, and long-term direction for Plumego.

## Current Status

**Latest Release:** v1.1.0 (June 2026)

**Stable Surface:** 9 stable roots with v1 compatibility guarantee  
**Beta Extensions:** 7 families (rest, websocket, gateway, observability, tenant, frontend, messaging)  
**Experimental Extensions:** 8 families (ai, data, fileapi, openapi, resilience, rpc, validate, others)

## v1.x Timeline

### v1.1.0 (Current)

**Released:** June 2026

**Highlights:**
- 9 stable root packages with full v1 guarantee
- 7 beta extension families
- Community-friendly documentation (adoption path, examples, guides)
- Agent-first control plane (specs, tasks, reference apps)
- Production-ready foundation

### v1.2 (Q3 2026)

**Focus:** Extension maturity and adoption clarity

**Planned:**
- Promote 2-3 experimental extensions to beta (based on community adoption)
- Add missing docs for x/* subpackages
- Expand migration guides (from more frameworks)
- Add benchmarking docs and examples
- Finalize v2 design (no breaking changes in v1.2)

**No breaking changes to stable roots or beta extensions.**

### v1.3 (Q4 2026)

**Focus:** Performance and observability

**Planned:**
- Performance optimizations (router, middleware chains)
- Enhanced observability defaults (auto-metrics, request tracing)
- More reference applications for common patterns
- Production hardening based on v1.1-1.2 feedback

**No breaking changes to stable roots or beta extensions.**

### v1.4+ (2027)

**Focus:** Stability and refinement

**Planned:**
- Bug fixes and security patches
- Documentation improvements
- Community-contributed extensions integration
- Two production release cycles with no exported API changes = candidate for v2

## v2.0 Planning

### Timeline

**Early 2027:** v2.0 design phase begins  
**Mid 2027:** v2.0 release candidate ready  
**Late 2027:** v2.0 released (not before 18 months of v1 stability)

### Expected Changes (NOT COMMITTED)

Based on v1 experience, we expect v2 to include:

**Likely:**
- Minimum Go version bump to 1.28+ (adds useful stdlib features)
- Removal of deprecated v1.x symbols
- Possible handler signature changes (if stdlib patterns evolve)
- Possible middleware pattern refinements

**Unlikely:**
- Core philosophy changes (stdlib first, explicit wiring)
- Removal of stable roots (they're designed for long-term stability)
- Major API redesigns without migration paths

**Not planned:**
- Dependency additions to stable roots
- Framework magic or hidden registration
- Removal of compatibility with plain stdlib

### Migration Path

When v2 is announced:

1. **6-month notice period** before v2.0 release
2. **Step-by-step migration guide** (like COMPATIBILITY.md, but for v1→v2)
3. **v1.x security fixes** continue for 6 months after v2 release
4. **Tools and scripts** to help automate upgrades where possible

## Support Commitment

| Version | Status | Bug fixes | Security fixes | End of Life |
|---------|--------|-----------|---|---|
| v1.4+ | Active | ✅ | ✅ | TBD (18+ months) |
| v1.3 | Maintenance | ✅ | ✅ | End of v1.4 support |
| v1.2 | Maintenance | ✅ | ✅ | End of v1.3 support |
| v1.1 | Maintenance | ✅ | ✅ | End of v1.2 support |
| v1.0 | Maintenance | Limited | ✅ | End of v1.1 support |
| < v1.0 | Unsupported | ❌ | ❌ | N/A |

**Minimum supported versions:** Latest 2 minor versions  
**Security patches:** Backported to supported versions within 48 hours  
**Breaking change notice:** Minimum 6 months before removal in the next major version

## Extension Maturity Path

Extensions follow this progression:

```
Experimental  →  (Community adoption)  →  Beta  →  (2+ release cycles stable)  →  GA
```

### Current Extensions Candidates

**Likely beta → ga candidates (v1.3+):**
- `x/rest` — Used in production by 5+ organizations, no API changes in 2 releases
- `x/websocket` — Stable pattern, production-ready
- `x/observability` — De-facto standard for Prometheus/OTel

**Likely to remain experimental:**
- `x/ai/*` — Fast-moving field, API changes expected
- `x/data/*` — Topology patterns evolving, needs more evidence
- `x/rpc` — Niche use case, lighter maintenance commitment

**May be archived:**
- `x/validate` — If community prefers external validation libraries
- `x/fileapi` — If less adopted than expected

## Feature Request Process

### For stable roots
Request → 3 months evaluation → Accepted/rejected with reasoning

Criteria:
- Solves a universal problem (not specific use case)
- Aligns with stdlib-first philosophy
- Doesn't bloat the learning path
- Can be maintained for 10+ years

### For extensions
Request → Prototype → Community feedback → Proposal for x/

Criteria:
- Solves a specific product need
- Can be published independently first
- Has production evidence (2+ organizations)
- Community interest demonstrated

### Process

1. **Open an issue** with the request
2. **Provide context:** Problem, why it's needed, expected benefit
3. **Show evidence:** Production use, community interest, prior art
4. **Design review:** Maintainers evaluate fit and design
5. **Decision:** Approved (planned for vX.Y), deferred, or declined (with reasoning)

## Communication

### Release announcements
- GitHub Releases page with migration notes
- Tweet / social media summary
- Blog post for major releases (v1.x that adds significant features)

### Pre-release communication
- Issue tracking planned features
- RFC (Request for Comments) issues for significant changes
- Stable roadmap update every 3 months

### Security issues
See `SECURITY.md` for vulnerability reporting and fix timelines.

## Deprecation Policy

Deprecated symbols follow this timeline:

1. **PR introduces deprecation:** Godoc comment added, changelog noted
2. **One minor version:** Symbol remains, works with warning in godoc
3. **Removal possible:** In the next minor version or next major version

Example:
- **v1.3.0:** `core.NewWithoutDefaults()` deprecated
- **v1.3.x, v1.4.x:** Symbol still works, deprecation noted
- **v1.5.0 or v2.0:** Symbol removed

All deprecations are announced in release notes and documented in COMPATIBILITY.md.

## Backward Compatibility Guarantees

### Within v1.x

- **Stable roots:** No breaking changes (v1 guarantee)
- **Beta extensions:** API stable across release refs; breaking changes require deprecation notice
- **Experimental extensions:** No guarantee; may change in any minor

### Beyond v1.x

We reserve the right to break compatibility in v2.0+, but:
- 6-month notice required
- Migration path provided
- Security and critical bug fixes for v1.x during transition

## Non-Goals

- **Not a framework overhaul:** v1 design is intentional and long-term
- **Not competing on speed:** We optimize for clarity and maintainability
- **Not batteries-included:** Extensions are optional by design
- **Not magic:** Explicit > implicit always

## Getting Involved

### Code contributions
- Bug fixes: Always welcome
- Features: Please open an issue first (discuss before coding)
- Extensions: Consider publishing independently, then propose for inclusion

### Feedback
- Issues: Bug reports, feature requests, design feedback
- Discussions: Architectural questions, usage patterns
- Surveys: We occasionally ask the community about priorities

### Maintenance
- Help wanted: Issues tagged `help wanted` are good starting points
- Extension adoption: Help maintain proven extensions
- Docs: Write guides, tutorials, migration paths

## Questions?

- **When is v2?** — Not before late 2027, with 6 months notice
- **Will I need to migrate?** — Yes, but we'll provide a clear path
- **Can I help shape the roadmap?** — Yes, open an issue with your priorities
- **Is Plumego production-ready?** — Yes, v1 is stable and recommended for production

---

**Last updated:** June 2026  
**Next review:** September 2026  
For breaking changes and migration guides, see COMPATIBILITY.md.
