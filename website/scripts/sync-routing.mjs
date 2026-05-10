import { readRepoFile, writeGeneratedFile, toTsModule } from './_shared.mjs';

// ── Manual annotations ────────────────────────────────────────────────────────
// classificationReason lives here, not in the YAML.
// task-routing.yaml is a machine-readable rule file; prose belongs in scripts.
const ANNOTATIONS = {
  en: {
    add_rate_limiting: {
      chipLabel: 'Add inbound rate limiting',
      taskDescription:
        'Add rate limiting to an inbound HTTP endpoint (e.g. the greet route).',
      classificationReason:
        '"Inbound" maps to transport layer. Rate limiting at the request boundary is middleware work — not outbound circuit protection (x/resilience). Destination: middleware/.',
    },
    fix_handler_bug: {
      chipLabel: 'Fix handler response bug',
      taskDescription:
        'Fix a defect where an HTTP handler returns the wrong status or malformed response body.',
      classificationReason:
        'Endpoint defect → explicit http_endpoint_bugfix recipe. Starts from contract + router; avoids middleware to keep transport wiring clean.',
    },
    tenant_jwt_policy: {
      chipLabel: 'Add multi-tenant JWT policy',
      taskDescription:
        'Add per-tenant JWT validation, quota enforcement, or policy evaluation.',
      classificationReason:
        'Tenant identity and policy live in x/tenant, not generic auth middleware. Stable roots carry the HTTP baseline; x/tenant carries the tenant layer.',
    },
    ai_streaming: {
      chipLabel: 'Build AI streaming endpoint',
      taskDescription:
        'Add an endpoint that streams responses from an AI provider (e.g. OpenAI-compatible SSE).',
      classificationReason:
        'AI capability is product work, not transport infrastructure. extension_work → x/ai. Stable roots remain compatible as AI API shapes evolve.',
    },
    websocket_support: {
      chipLabel: 'Add WebSocket support',
      taskDescription: 'Add WebSocket upgrade handling to an existing HTTP server.',
      classificationReason:
        'Protocol adaptation (HTTP → WS) is extension work. x/websocket coexists with standard HTTP routes using the same middleware chain.',
    },
    api_gateway_proxy: {
      chipLabel: 'Set up API gateway proxy',
      taskDescription:
        'Add reverse proxy, upstream routing, or load balancing at the edge.',
      classificationReason:
        'Edge transport is gateway_edge_transport task → x/gateway. Does not change the stable routing model used by upstream services.',
    },
    file_upload: {
      chipLabel: 'Add file upload endpoint',
      taskDescription: 'Add multipart file upload and download handling.',
      classificationReason:
        'File upload is product capability, not a storage primitive. file_api task → x/fileapi. Avoids store/file and x/data/file to keep primitives separate.',
    },
    route_registration: {
      chipLabel: 'Update route registration',
      taskDescription:
        'Add or restructure routes and handler wiring in an existing service.',
      classificationReason:
        'Route registration is app-local wiring, not a stable root change. app_wiring → reference/standard-service internal/app/routes.go.',
    },
    change_arch_rules: {
      chipLabel: 'Change dependency boundary rule',
      taskDescription:
        'Update the allowed import directions between layers in specs/dependency-rules.yaml.',
      classificationReason:
        'Architecture rule changes belong in the control plane. repo_rules → specs/. Do not edit Go packages for this — the rule is machine-readable only.',
    },
    new_stable_module: {
      chipLabel: 'Review a new stable module',
      taskDescription:
        'Evaluate or add a new package to the stable roots (e.g. a new transport primitive).',
      classificationReason:
        'Stable root expansion requires a boundary review recipe before coding. stable_root_boundary_review → reads all nine existing stable root manifests.',
    },
  },
  zh: {
    add_rate_limiting: {
      chipLabel: '加入站限流',
      taskDescription: '给某个入站 HTTP 端点加入限流（如 greet 路由）。',
      classificationReason:
        '"入站" 映射到传输层。请求边界的限流是中间件工作，不是出站熔断（x/resilience）。目标：middleware/。',
    },
    fix_handler_bug: {
      chipLabel: '修复 handler 响应缺陷',
      taskDescription: '修复 HTTP handler 返回错误状态码或响应体格式错误的缺陷。',
      classificationReason:
        '端点缺陷 → 使用 http_endpoint_bugfix recipe。从 contract + router 开始；避免 middleware，保持传输配线的清晰。',
    },
    tenant_jwt_policy: {
      chipLabel: '多租户 JWT 策略',
      taskDescription: '为每个租户加入 JWT 校验、quota 限制或策略评估。',
      classificationReason:
        '租户身份与策略在 x/tenant，不在通用 auth middleware。稳定根承担 HTTP 基线；x/tenant 承担租户层。',
    },
    ai_streaming: {
      chipLabel: '构建 AI 流式端点',
      taskDescription: '加入从 AI provider 流式返回响应的端点（如 OpenAI 兼容的 SSE）。',
      classificationReason:
        'AI 能力是产品工作，不是传输基础设施。extension_work → x/ai。稳定根在 AI API 形态演化时保持兼容。',
    },
    websocket_support: {
      chipLabel: '加入 WebSocket 支持',
      taskDescription: '在现有 HTTP 服务中加入 WebSocket 升级处理。',
      classificationReason:
        '协议适配（HTTP → WS）是扩展工作。x/websocket 与 HTTP 路由共存，使用同一个中间件链。',
    },
    api_gateway_proxy: {
      chipLabel: 'API 网关反向代理',
      taskDescription: '在边缘加入反向代理、上游路由或负载均衡。',
      classificationReason:
        '边缘传输是 gateway_edge_transport 任务 → x/gateway。不改变上游服务使用的稳定路由模型。',
    },
    file_upload: {
      chipLabel: '加入文件上传接口',
      taskDescription: '加入 multipart 文件上传和下载处理。',
      classificationReason:
        '文件上传是产品能力，不是 store primitive。file_api 任务 → x/fileapi。避免 store/file 和 x/data/file，保持 primitive 分离。',
    },
    route_registration: {
      chipLabel: '更新路由注册',
      taskDescription: '在现有服务中新增或重构路由和 handler 配线。',
      classificationReason:
        '路由注册是应用本地配线，不是稳定根变更。app_wiring → reference/standard-service internal/app/routes.go。',
    },
    change_arch_rules: {
      chipLabel: '修改依赖边界规则',
      taskDescription: '更新 specs/dependency-rules.yaml 中各层之间的 import 方向。',
      classificationReason:
        '架构规则变更属于控制面。repo_rules → specs/。这里不改 Go 包——规则是纯机器可读的。',
    },
    new_stable_module: {
      chipLabel: '审查新稳定模块',
      taskDescription: '评估或添加一个新包到稳定根（如新的传输 primitive）。',
      classificationReason:
        '稳定根扩展在编码前需要边界审查 recipe。stable_root_boundary_review → 读取全部九个现有稳定根的 manifest。',
    },
  },
};

// ── YAML extraction helpers ───────────────────────────────────────────────────

function extractTaskBlock(yaml, taskKey) {
  // Match "  taskKey:\n" followed by indented content until next top-level key
  const pattern = new RegExp(
    `  ${taskKey}:\\n((?:    .+\\n)*)`,
    'm',
  );
  const match = yaml.match(pattern);
  return match ? match[1] : '';
}

function extractList(block, key) {
  const pattern = new RegExp(`    ${key}:\\n((?:      - .+\\n)+)`, 'm');
  const match = block.match(pattern);
  if (!match) return [];
  return match[1]
    .split('\n')
    .map((l) => l.replace(/^\s*-\s+/, '').trim())
    .filter(Boolean);
}

function extractIntent(block) {
  const m = block.match(/    intent:\s+(.+)/);
  return m ? m[1].trim() : '';
}

// ── Scenario definitions ──────────────────────────────────────────────────────
// Maps scenario id → { ruleName, taskYamlKey, ruleSection ('routing_rules'|'tasks'), primerPath? }
const SCENARIO_DEFS = [
  {
    id: 'add_rate_limiting',
    ruleName: 'middleware',
    taskYamlKey: 'middleware',
    ruleSection: 'tasks',
    destination: 'middleware/',
    primerPath: '/docs/modules/middleware',
  },
  {
    id: 'fix_handler_bug',
    ruleName: 'http_endpoint_bugfix',
    taskYamlKey: 'http_endpoint_bugfix',
    ruleSection: 'tasks',
    destination: 'contract/ + router/',
    primerPath: '/docs/modules/contract',
  },
  {
    id: 'tenant_jwt_policy',
    ruleName: 'tenant_policy_change',
    taskYamlKey: 'tenant_policy_change',
    ruleSection: 'tasks',
    destination: 'x/tenant/',
    primerPath: '/docs/modules/x-tenant',
  },
  {
    id: 'ai_streaming',
    ruleName: 'extension_work → x/ai',
    taskYamlKey: null,
    ruleSection: 'routing_rules',
    destination: 'x/ai/',
    startWith: ['x/ai/module.yaml', 'docs/modules/x-ai/README.md', 'x/ai/entrypoints.go'],
    avoid: ['core', 'router', 'middleware'],
    primerPath: '/docs/modules/x-ai',
  },
  {
    id: 'websocket_support',
    ruleName: 'websocket',
    taskYamlKey: 'websocket',
    ruleSection: 'tasks',
    destination: 'x/websocket/',
    primerPath: '/docs/modules/x-websocket',
  },
  {
    id: 'api_gateway_proxy',
    ruleName: 'gateway_edge_transport',
    taskYamlKey: 'gateway_edge_transport',
    ruleSection: 'tasks',
    destination: 'x/gateway/',
    primerPath: '/docs/modules/x-gateway',
  },
  {
    id: 'file_upload',
    ruleName: 'file_api',
    taskYamlKey: 'file_api',
    ruleSection: 'tasks',
    destination: 'x/fileapi/',
    primerPath: '/docs/modules/x-fileapi',
  },
  {
    id: 'route_registration',
    ruleName: 'app_wiring',
    taskYamlKey: null,
    ruleSection: 'routing_rules',
    destination: 'reference/standard-service',
    startWith: [
      'reference/standard-service/main.go',
      'reference/standard-service/internal/app/app.go',
      'reference/standard-service/internal/app/routes.go',
    ],
    avoid: ['x/rest', 'x/gateway', 'x/webhook'],
    primerPath: '/docs/reference-app',
  },
  {
    id: 'change_arch_rules',
    ruleName: 'repo_rules',
    taskYamlKey: null,
    ruleSection: 'routing_rules',
    destination: 'specs/',
    startWith: ['specs/dependency-rules.yaml', 'specs/repo.yaml', 'AGENTS.md'],
    avoid: ['reference/standard-service', 'cmd/plumego'],
    primerPath: '/docs/concepts/repo-control-plane',
  },
  {
    id: 'new_stable_module',
    ruleName: 'stable_root_boundary_review',
    taskYamlKey: 'stable_root_boundary_review',
    ruleSection: 'tasks',
    destination: 'stable roots (new package)',
    primerPath: '/docs/stable-roots',
  },
];

// ── Main sync ─────────────────────────────────────────────────────────────────

export async function syncRouting() {
  const yaml = await readRepoFile('specs/task-routing.yaml');

  const tasksSection = yaml.match(/\ntasks:\n([\s\S]*)$/m)?.[1] ?? '';

  const scenariosEn = [];
  const scenariosZh = [];

  for (const def of SCENARIO_DEFS) {
    let startWith = def.startWith ?? [];
    let avoid = def.avoid ?? [];
    let destination = def.destination ?? '';
    let intent = '';

    if (def.taskYamlKey && def.ruleSection === 'tasks') {
      const block = extractTaskBlock(tasksSection, def.taskYamlKey);
      if (block) {
        const sw = extractList(block, 'start_with');
        const av = extractList(block, 'avoid');
        const intentVal = extractIntent(block);
        if (sw.length) startWith = sw;
        if (av.length) avoid = av;
        if (intentVal) intent = intentVal;
      }
    }

    const annotEn = ANNOTATIONS.en[def.id];
    const annotZh = ANNOTATIONS.zh[def.id];

    scenariosEn.push({
      id: def.id,
      chipLabel: annotEn.chipLabel,
      taskDescription: annotEn.taskDescription,
      ruleName: def.ruleName,
      intent: intent || annotEn.taskDescription,
      destination,
      startWith,
      avoid,
      classificationReason: annotEn.classificationReason,
      primerPath: def.primerPath ?? null,
    });

    scenariosZh.push({
      id: def.id,
      chipLabel: annotZh.chipLabel,
      taskDescription: annotZh.taskDescription,
      ruleName: def.ruleName,
      intent: intent || annotZh.taskDescription,
      destination,
      startWith,
      avoid,
      classificationReason: annotZh.classificationReason,
      primerPath: def.primerPath ?? null,
    });
  }

  await writeGeneratedFile(
    'routing.ts',
    toTsModule('ROUTING_SCENARIOS', { en: scenariosEn, zh: scenariosZh }),
  );

  console.log(`  routing: wrote ${scenariosEn.length} scenarios`);
}

if (import.meta.url === `file://${process.argv[1]}`) {
  await syncRouting();
}
