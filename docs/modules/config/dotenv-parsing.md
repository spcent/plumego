# .env File Parsing

> **Package**: `github.com/spcent/plumego/config`

The current package supports `.env` parsing in two explicit forms:

- `config.LoadEnvFile(path, overwrite)` to write values into process environment
- `config.NewFileSource(path, config.FormatEnv, watch)` to load `.env` content into a `config.Manager`

## File format

```bash
APP_ADDR=:8080
APP_DEBUG=true
DB_URL="postgres://localhost/db"
OPTIONAL_KEY=
PORT = 8080
```

Comments beginning with `#` are ignored. Keys are normalized to snake_case in `Manager` source loading.

## Load into process environment

```go
if err := config.LoadEnvFile(".env", false); err != nil {
    log.Fatal(err)
}
```

Use this in `main` when you explicitly want process-global mutation.

## Load into a manager without mutating process environment

```go
cfg := config.NewManager(plumelog.NewGLogger())
_ = cfg.AddSource(config.NewFileSource(".env.production", config.FormatEnv, false))
if err := cfg.Load(context.Background()); err != nil {
    log.Fatal(err)
}
```

## Watching a `.env` file

```go
source := config.NewFileSource(".env", config.FormatEnv, true).WithWatchInterval(time.Second)
```

## Recommended file layout

```bash
.env.example
.env.local
.env.production
.env.staging
.env.test
```

## Priority guidance

Use explicit source ordering rather than relying on hidden default lookup:

1. base file
2. environment-specific file
3. real environment variables

```go
_ = cfg.AddSource(config.NewFileSource(".env", config.FormatEnv, false))
_ = cfg.AddSource(config.NewFileSource(".env.production", config.FormatEnv, false))
_ = cfg.AddSource(config.NewEnvSource(""))
```
