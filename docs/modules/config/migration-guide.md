# Migration Guide

> **Package**: `github.com/spcent/plumego/config`

Guide for migrating from other configuration systems to Plumego config.

---

## From Viper

### Before (Viper)

```go
import "github.com/spf13/viper"

viper.SetDefault("port", 8080)
viper.AutomaticEnv()
viper.SetConfigFile(".env")
viper.ReadInConfig()

port := viper.GetInt("port")
debug := viper.GetBool("debug")
```

### After (Plumego)

```go
import "github.com/spcent/plumego/config"

cfg := config.Load()

port := cfg.GetInt("PORT", 8080)
debug := cfg.GetBool("DEBUG", false)
```

---

## From godotenv

### Before (godotenv)

```go
import "github.com/joho/godotenv"

godotenv.Load()

addr := os.Getenv("APP_ADDR")
if addr == "" {
    addr = ":8080"
}

debugStr := os.Getenv("APP_DEBUG")
debug := debugStr == "true"
```

### After (Plumego)

```go
import "github.com/spcent/plumego/config"

cfg := config.Load()

addr := cfg.Get("APP_ADDR", ":8080")
debug := cfg.GetBool("APP_DEBUG", false)
```

---

## From envconfig

### Before (envconfig)

```go
import "github.com/kelseyhightower/envconfig"

type Config struct {
    Port  int    `envconfig:"PORT" default:"8080"`
    Debug bool   `envconfig:"DEBUG" default:"false"`
}

var cfg Config
envconfig.Process("", &cfg)
```

### After (Plumego)

```go
import "github.com/spcent/plumego/config"

type Config struct {
    Port  int
    Debug bool
}

func LoadConfig() Config {
    cfg := config.Load()
    return Config{
        Port:  cfg.GetInt("PORT", 8080),
        Debug: cfg.GetBool("DEBUG", false),
    }
}

cfg := LoadConfig()
```

---

## From os.Getenv

### Before (os.Getenv)

```go
addr := os.Getenv("APP_ADDR")
if addr == "" {
    addr = ":8080"
}

portStr := os.Getenv("PORT")
port := 8080
if portStr != "" {
    if p, err := strconv.Atoi(portStr); err == nil {
        port = p
    }
}

debugStr := os.Getenv("DEBUG")
debug := false
if debugStr == "true" || debugStr == "1" {
    debug = true
}
```

### After (Plumego)

```go
import "github.com/spcent/plumego/config"

cfg := config.Load()

addr := cfg.Get("APP_ADDR", ":8080")
port := cfg.GetInt("PORT", 8080)
debug := cfg.GetBool("DEBUG", false)
```

---

## Migration Checklist

### 1. Replace Import

```go
// Remove
import "github.com/spf13/viper"
// or
import "github.com/joho/godotenv"

// Add
import "github.com/spcent/plumego/config"
```

### 2. Replace Initialization

```go
// Remove
viper.AutomaticEnv()
viper.ReadInConfig()
// or
godotenv.Load()

// Add
cfg := config.Load()
```

### 3. Replace Get Methods

```go
// Before
port := viper.GetInt("port")
debug := viper.GetBool("debug")

// After
port := cfg.GetInt("PORT", 8080)
debug := cfg.GetBool("DEBUG", false)
```

### 4. Add Defaults

```go
// Viper often uses SetDefault()
viper.SetDefault("port", 8080)
port := viper.GetInt("port")

// Plumego uses inline defaults
port := cfg.GetInt("PORT", 8080)
```

### 5. Update Environment Variable Names

```go
// Many systems use lowercase or mixed case
viper.Get("port")
viper.Get("debugMode")

// Plumego convention: UPPERCASE with underscores
cfg.Get("PORT", "8080")
cfg.GetBool("DEBUG_MODE", false)
```

---

## Benefits of Migration

1. **No Dependencies**: Plumego config uses only standard library
2. **Simpler API**: Type-safe methods with defaults
3. **Consistent**: Works seamlessly with other Plumego modules
4. **Lightweight**: Minimal overhead, fast startup

---

## Complete Example

### Before (Multiple Libraries)

```go
package main

import (
    "github.com/joho/godotenv"
    "github.com/spf13/viper"
    "os"
)

func main() {
    godotenv.Load()
    viper.AutomaticEnv()

    addr := os.Getenv("APP_ADDR")
    if addr == "" {
        addr = ":8080"
    }

    port := viper.GetInt("PORT")
    if port == 0 {
        port = 8080
    }

    debug := viper.GetBool("DEBUG")

    // Use config...
}
```

### After (Plumego)

```go
package main

import "github.com/spcent/plumego/config"

func main() {
    cfg := config.Load()

    addr := cfg.Get("APP_ADDR", ":8080")
    port := cfg.GetInt("PORT", 8080)
    debug := cfg.GetBool("DEBUG", false)

    // Use config...
}
```

---

**Complete**: [Config Overview](README.md)

---

🎉 **P0 Documentation Complete!**
