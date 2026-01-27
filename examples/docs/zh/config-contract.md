# 配置与启动契约

本文档定义 Plumego 推荐的配置与启动模型，统一模式但不强制依赖外部库。

## Config Struct 模式
建议使用单一配置结构体，包含 core 配置与业务配置：

```go
type Config struct {
	Core core.AppConfig

	// 业务配置
	FeatureXEnabled bool
	ExternalAPIURL  string
}
```

推荐函数：
- `Defaults() Config`
- `LoadConfig() (Config, error)`
- `Validate(cfg Config) error`

## Env 覆盖策略
推荐优先级：
1. `Defaults()` 默认值
2. `.env`（可选，`core.WithEnvPath` + `core.Boot()`）
3. 环境变量（显式覆盖）
4. 命令行参数（最高优先级）

命名建议：
- 环境变量使用大写下划线。
- 应用级变量推荐统一前缀（如 `APP_`）。
- 新增变量需同步更新 `env.example`。

## 启动顺序（推荐）
1. 解析命令行参数（覆盖变量）。
2. 构建默认配置。
3. 应用环境变量覆盖。
4. 校验配置。
5. 根据配置组装 `core.New(...)` options。
6. 注册 middleware / component / runner。
7. 调用 `app.Boot()`。

保证启动确定性，避免隐式副作用。

## 标准示例（参考风格）

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
