# Error Naming Standard for Plumego v1.0

## Current Status

This document outlines the error naming conventions adopted across the plumego codebase for consistency and clarity.

## Standard Convention

### Error Variable Naming

All exported error variables should follow this pattern:

```go
var Err<Category><Specific> = errors.New("description")
```

**Rules:**
1. Prefix with `Err`
2. Use PascalCase
3. Start with broad category, then specific detail
4. Keep concise but descriptive

### Examples

#### ✅ Good Examples

```go
// Generic errors (no prefix needed if truly generic)
var ErrNotFound = errors.New("not found")
var ErrClosed = errors.New("closed")
var ErrTimeout = errors.New("timeout")

// Domain-specific errors (include domain prefix)
var ErrCacheFull = errors.New("cache: cache full")
var ErrCacheKeyTooLong = errors.New("cache: key too long")

// Service-specific errors (include service prefix for clarity)
var ErrGitHubSignature = errors.New("github: invalid signature")
var ErrGitHubMissingHeader = errors.New("github: missing header")

var ErrStripeMissingHeader = errors.New("stripe: missing header")
var ErrStripeInvalidTimestamp = errors.New("stripe: invalid timestamp")
```

#### ❌ Inconsistent (to avoid)

```go
// Mixing prefixed and non-prefixed in same package
var ErrInvalidHex = errors.New("...")  // Should be ErrSignerInvalidHex
var SignatureError = errors.New("...")  // Should be ErrSignatureInvalid
```

## Current State Analysis

### Consistent Packages ✅

- `store/cache/` - All errors prefixed with `ErrCache*`
- `store/db/` - All errors prefixed with `ErrDb*` or domain-specific
- `store/kv/` - All errors prefixed with `ErrKey*` or `ErrStore*`
- `pubsub/` - All errors follow pattern
- `contract/` - Error types well-structured

### Packages Needing Minor Improvements

#### `net/webhookin/`

**Current:**
```go
var ErrGitHubSignature = errors.New("...")
var ErrGitHubMissingHeader = errors.New("...")
var ErrStripeMissingHeader = errors.New("...")
```

**Assessment:** ✅ Good - Service name prefixes make sense here

#### `net/webhookout/`

**Current:**
```go
var ErrInvalidHex = errors.New("...")
```

**Recommendation:**
```go
var ErrSignerInvalidHex = errors.New("webhookout: invalid hex encoding")
```

## Migration Strategy

### For v1.0

**Decision:** Keep current naming as-is for v1.0 to avoid breaking changes.

**Rationale:**
1. Current naming is mostly consistent
2. Breaking changes for minor improvements not justified
3. Users may already depend on error variable names
4. Error messages (strings) are clear and helpful

### For Future Versions

If breaking changes are needed, provide:
1. Type aliases for backward compatibility
2. Deprecation notices
3. Migration guide

Example:
```go
// Deprecated: Use ErrSignerInvalidHex instead
var ErrInvalidHex = ErrSignerInvalidHex

var ErrSignerInvalidHex = errors.New("webhookout: invalid hex encoding")
```

## Guidelines for New Code

When adding new errors:

1. **Check existing errors** in the package first
2. **Follow the established pattern** in that package
3. **Use domain prefix** for package-specific errors
4. **Keep messages clear** - what failed and why
5. **Use lowercase** in error messages (Go convention)
6. **Be specific** but not verbose

### Template

```go
// Err<Domain><What> describes when <what> fails in <domain>
var Err<Domain><What> = errors.New("<domain>: <what> <why>")

// Example:
// ErrCacheKeyTooLong is returned when a cache key exceeds maximum length
var ErrCacheKeyTooLong = errors.New("cache: key too long")
```

## Summary

**For v1.0:** Current error naming is **acceptable** and consistent enough. No breaking changes needed.

**For future:** Document any deviations and maintain consistency within each package.

**Key principle:** Consistency within a package is more important than cross-package uniformity.
