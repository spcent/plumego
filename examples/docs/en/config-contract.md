# Config & Startup Contract

This document defines the recommended configuration and startup model for Plumego
applications. It standardizes patterns without requiring any external library.

## Config Struct Pattern
Define a single config struct that embeds core settings and your app settings:

```go
type Config struct {
	Core core.AppConfig

	// App-specific settings
	FeatureXEnabled bool
	ExternalAPIURL  string
}
```

Recommended functions:
- `Defaults() Config`
- `LoadConfig() (Config, error)`
- `Validate(cfg Config) error`

## Env Override Strategy
Use a clear precedence order:
1. Defaults in `Defaults()`
2. `.env` file (optional) via `core.WithEnvPath` + `core.Boot()`
3. Environment variables (explicit overrides)
4. Command-line flags (highest priority)

Naming conventions:
- Use UPPER_SNAKE for env vars.
- Prefer a stable prefix (e.g. `APP_`) for app-specific settings.
- Keep `env.example` in sync with any new env vars.

## Startup Sequence (Recommended)
1. Parse flags (override variables).
2. Build config defaults.
3. Apply env overrides.
4. Validate config.
5. Build `core.New(...)` options from config.
6. Register middleware/components/runners.
7. Call `app.Boot()`.

This keeps startup deterministic and avoids hidden side effects.

## Standard Example (Reference Style)

```go
func Defaults() Config {
	return Config{
		Core: core.AppConfig{
			Addr: ":8080",
			Debug: false,
		},
		FeatureXEnabled: true,
	}
}

func LoadConfig() (Config, error) {
	cfg := Defaults()

	cfg.Core.Addr = config.GetString("APP_ADDR", cfg.Core.Addr)
	cfg.Core.Debug = config.GetBool("APP_DEBUG", cfg.Core.Debug)
	cfg.FeatureXEnabled = config.GetBool("APP_FEATURE_X_ENABLED", cfg.FeatureXEnabled)

	return cfg, nil
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	app := core.New(
		core.WithAddr(cfg.Core.Addr),
		core.WithDebug(),
	)

	if err := app.Boot(); err != nil {
		log.Fatalf("boot error: %v", err)
	}
}
```
