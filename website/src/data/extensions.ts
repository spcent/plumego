// Extension display metadata — descriptions, entrypoints, doc hrefs, and search tags.
// Maturity status (beta / experimental) is authoritative in MODULE_FACTS from
// generated/modules.ts. Keep descriptions here aligned with module primer summaries.

export interface ExtensionMeta {
  description: string;
  entrypoint: string;
  docHref: string;
  tags: string[];
}

export const EXTENSION_META_EN: Record<string, ExtensionMeta> = {
  'x/rest': {
    description: 'CRUD resource conventions, typed handlers, and REST response contracts over the stable HTTP kernel.',
    entrypoint: 'x/rest',
    docHref: '/docs/modules/x-rest',
    tags: ['api', 'rest', 'crud'],
  },
  'x/gateway': {
    description: 'Reverse proxy, upstream health, load balancing, and route composition at the service edge.',
    entrypoint: 'x/gateway',
    docHref: '/docs/modules/x-gateway',
    tags: ['proxy', 'edge', 'lb'],
  },
  'x/websocket': {
    description: 'Full-duplex WebSocket connections managed by a hub, coexisting with standard HTTP routes.',
    entrypoint: 'x/websocket',
    docHref: '/docs/modules/x-websocket',
    tags: ['realtime', 'ws'],
  },
  'x/observability': {
    description: 'Prometheus metrics, OpenTelemetry traces, and structured log exporters wired at the service boundary.',
    entrypoint: 'x/observability',
    docHref: '/docs/modules/x-observability',
    tags: ['metrics', 'traces', 'otel'],
  },
  'x/tenant': {
    description: 'Per-tenant JWT identity, quota enforcement, and policy evaluation at the transport layer.',
    entrypoint: 'x/tenant',
    docHref: '/docs/modules/x-tenant',
    tags: ['saas', 'multi-tenant', 'jwt'],
  },
  'x/frontend': {
    description: 'Static asset serving, SPA fallback routing, and embedded filesystem support for frontend integration.',
    entrypoint: 'x/frontend',
    docHref: '/docs/modules/x-frontend',
    tags: ['static', 'spa', 'assets'],
  },
  'x/messaging': {
    description: 'Event publishing, webhook dispatch, and messaging service-level API over pluggable transports.',
    entrypoint: 'x/messaging',
    docHref: '/docs/modules/x-messaging',
    tags: ['events', 'pubsub', 'webhooks'],
  },
  'x/ai': {
    description: 'Multi-provider AI abstraction, session lifecycle, streaming SSE responses, and tool-call routing.',
    entrypoint: 'x/ai/provider',
    docHref: '/docs/modules/x-ai',
    tags: ['ai', 'llm', 'streaming'],
  },
  'x/data': {
    description: 'File storage, idempotency keys, and data access primitives above the stable store interface.',
    entrypoint: 'x/data',
    docHref: '/docs/modules/x-data',
    tags: ['storage', 'data', 'files'],
  },
  'x/resilience': {
    description: 'Circuit breakers, retry policies, and timeout wrappers for outbound HTTP and service calls.',
    entrypoint: 'x/resilience',
    docHref: '/docs/modules/x-resilience',
    tags: ['resilience', 'circuit-breaker', 'retry'],
  },
  'x/rpc': {
    description: 'gRPC + HTTP dual-protocol service wiring that shares the same stable middleware chain.',
    entrypoint: 'x/rpc',
    docHref: '/docs/modules/x-rpc',
    tags: ['grpc', 'rpc', 'protocol'],
  },
  'x/openapi': {
    description: 'OpenAPI 3 spec generation from route definitions, with schema inference and doc server.',
    entrypoint: 'x/openapi',
    docHref: '/docs/modules/x-openapi',
    tags: ['openapi', 'swagger', 'docs'],
  },
  'x/validate': {
    description: 'Request body validation middleware with struct tag rules and structured error responses.',
    entrypoint: 'x/validate',
    docHref: '/docs/modules/x-validate',
    tags: ['validation', 'input'],
  },
  'x/fileapi': {
    description: 'Multipart upload handlers, chunk resumption, and file metadata management over the stable HTTP kernel.',
    entrypoint: 'x/fileapi',
    docHref: '/docs/modules/x-fileapi',
    tags: ['upload', 'files', 'multipart'],
  },
};

export const EXTENSION_META_ZH: Record<string, ExtensionMeta> = {
  'x/rest': {
    description: 'CRUD 资源规范、类型化 handler 和基于稳定 HTTP 内核的 REST 响应 contract。',
    entrypoint: 'x/rest',
    docHref: '/zh/docs/modules/x-rest',
    tags: ['api', 'rest', 'crud'],
  },
  'x/gateway': {
    description: '反向代理、上游健康检查、负载均衡以及服务边缘的路由组合。',
    entrypoint: 'x/gateway',
    docHref: '/zh/docs/modules/x-gateway',
    tags: ['proxy', 'edge', 'lb'],
  },
  'x/websocket': {
    description: '由 hub 管理的全双工 WebSocket 连接，与标准 HTTP 路由共存。',
    entrypoint: 'x/websocket',
    docHref: '/zh/docs/modules/x-websocket',
    tags: ['realtime', 'ws'],
  },
  'x/observability': {
    description: 'Prometheus 指标、OpenTelemetry 链路追踪和结构化日志导出，接入服务边界。',
    entrypoint: 'x/observability',
    docHref: '/zh/docs/modules/x-observability',
    tags: ['metrics', 'traces', 'otel'],
  },
  'x/tenant': {
    description: '在传输层进行 per-tenant JWT 身份识别、配额执行和策略评估。',
    entrypoint: 'x/tenant',
    docHref: '/zh/docs/modules/x-tenant',
    tags: ['saas', 'multi-tenant', 'jwt'],
  },
  'x/frontend': {
    description: '静态资产服务、SPA fallback 路由和嵌入式文件系统支持，用于前端集成。',
    entrypoint: 'x/frontend',
    docHref: '/zh/docs/modules/x-frontend',
    tags: ['static', 'spa', 'assets'],
  },
  'x/messaging': {
    description: '基于可插拔 transport 的事件发布、webhook 分发和消息服务层 API。',
    entrypoint: 'x/messaging',
    docHref: '/zh/docs/modules/x-messaging',
    tags: ['events', 'pubsub', 'webhooks'],
  },
  'x/ai': {
    description: '多 provider AI 抽象、session 生命周期、streaming SSE 响应以及 tool-call 路由。',
    entrypoint: 'x/ai/provider',
    docHref: '/zh/docs/modules/x-ai',
    tags: ['ai', 'llm', 'streaming'],
  },
  'x/data': {
    description: '基于稳定 store 接口之上的文件存储、幂等键和数据访问原语。',
    entrypoint: 'x/data',
    docHref: '/zh/docs/modules/x-data',
    tags: ['storage', 'data', 'files'],
  },
  'x/resilience': {
    description: '用于出站 HTTP 和服务调用的熔断器、重试策略和超时包装器。',
    entrypoint: 'x/resilience',
    docHref: '/zh/docs/modules/x-resilience',
    tags: ['resilience', 'circuit-breaker', 'retry'],
  },
  'x/rpc': {
    description: 'gRPC + HTTP 双协议服务装配，共享同一条稳定中间件链。',
    entrypoint: 'x/rpc',
    docHref: '/zh/docs/modules/x-rpc',
    tags: ['grpc', 'rpc', 'protocol'],
  },
  'x/openapi': {
    description: '从路由定义生成 OpenAPI 3 规范，支持 schema 推断和文档服务器。',
    entrypoint: 'x/openapi',
    docHref: '/zh/docs/modules/x-openapi',
    tags: ['openapi', 'swagger', 'docs'],
  },
  'x/validate': {
    description: '带 struct tag 规则和结构化错误响应的请求体验证中间件。',
    entrypoint: 'x/validate',
    docHref: '/zh/docs/modules/x-validate',
    tags: ['validation', 'input'],
  },
  'x/fileapi': {
    description: '基于稳定 HTTP 内核的 multipart 上传 handler、分片续传和文件元数据管理。',
    entrypoint: 'x/fileapi',
    docHref: '/zh/docs/modules/x-fileapi',
    tags: ['upload', 'files', 'multipart'],
  },
};
