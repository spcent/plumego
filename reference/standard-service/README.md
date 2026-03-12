# Standard Service Reference

`reference/standard-service` is Plumego's canonical reference application.

It is the single source of truth for:

- default application layout
- bootstrap flow in `main.go`
- route registration style
- handler and response conventions

Run it with:

```bash
go run ./reference/standard-service
```

When templates or docs need a project structure example, they should follow this directory instead of `examples/`.
