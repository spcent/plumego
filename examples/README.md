# Plumego Examples

Minimal, runnable examples demonstrating core Plumego concepts and patterns.

**All examples use only stable root packages and compile with no external dependencies beyond Plumego.**

## Examples

### [`hello/`](./hello/)

Minimal "hello world" service with path parameters and logging.

**Run:** `cd hello && go build && ./hello`

**Topics:**
- App initialization
- Route registration
- JSON responses
- Logging

### [`json-api/`](./json-api/)

REST-like JSON API with CRUD operations, request/response encoding, and error handling.

**Run:** `cd json-api && go build && ./json-api`

**Topics:**
- POST and GET routes
- JSON encoding/decoding
- Path parameters
- Error handling
- Thread-safe storage

## Learning Path

1. **Start with [`hello/`](./hello/)** to understand app basics.
2. **Move to [`json-api/`](./json-api/)** to see CRUD and error handling.
3. **Read [`reference/standard-service`](../reference/standard-service)** for a complete production-style application.

## Principles

- Examples are minimal: they teach one concept without unnecessary features.
- All examples compile and run without setup.
- Examples use only stable root packages (no `x/*` extensions).
- Each example has a `README.md` explaining the code and demonstrating usage.

## Running All Examples

Verify that all examples compile:

```bash
go build ./examples/hello/
go build ./examples/json-api/
```

Or with Make:

```bash
make validate-examples
```

(Note: This requires a `validate-examples` target in the root Makefile if you want one.)

## Next Steps

- **For advanced patterns:** See [`reference/`](../reference/).
- **For optional capabilities:** See [`docs/modules/x/`](../docs/modules/x/).
- **For style and wiring conventions:** See [`docs/reference/canonical-style-guide.md`](../docs/reference/canonical-style-guide.md).
