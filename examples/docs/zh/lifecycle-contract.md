# 生命周期契约

本文档定义后台任务的稳定生命周期语义。

## Runner 接口
```go
type Runner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

## 注册方式
- 使用 `app.Register(runner)` 或 `core.WithRunner` 注册。
- 必须在 `Boot()` 之前完成注册。

Shutdown hook：
- 使用 `app.OnShutdown(func(ctx context.Context) error { ... })` 或 `core.WithShutdownHook` 注册。
- 必须在 `Boot()` 之前完成注册。

## 启动顺序
1. 加载环境配置。
2. 挂载组件（路由 + 中间件）。
3. 初始化 HTTP server。
4. 启动组件。
5. 启动 Runner。
6. 启动 HTTP server。

Runner 必须在 HTTP server 对外服务前启动。

## 关闭顺序
1. 关闭 HTTP server（优雅退出）。
2. 停止 Runner（按注册顺序反向）。
3. 停止组件（按依赖顺序反向）。
4. 执行 shutdown hook（按注册顺序反向）。

## Context 约束
- `Start` / `Stop` 必须遵守 ctx deadline。
- 在 `ctx.Done()` 后应尽快退出。
