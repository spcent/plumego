import { MODULE_FACTS } from '../generated/modules';
import { RELEASE_FACTS } from '../generated/releases';

export type Locale = 'en' | 'zh';

export const SITE = {
  name: 'Plumego',
  githubUrl: 'https://github.com/spcent/plumego',
  repoPath: 'reference/standard-service',
  currentVersion: RELEASE_FACTS.currentVersion,
};

export const NAV_LINKS: Record<Locale, Array<{ label: string; href: string }>> = {
  en: [
    { label: 'Docs', href: '/docs' },
    { label: 'Modules', href: '/docs/modules/overview' },
    { label: 'Roadmap', href: '/roadmap' },
    { label: 'Releases', href: '/releases' },
    { label: 'GitHub', href: SITE.githubUrl },
  ],
  zh: [
    { label: '文档', href: '/zh/docs' },
    { label: '模块', href: '/zh/docs/modules/overview' },
    { label: '路线图', href: '/zh/roadmap' },
    { label: '发布', href: '/zh/releases' },
    { label: 'GitHub', href: SITE.githubUrl },
  ],
};

export const HOME_COPY = {
  en: {
    eyebrow: 'Standard Library First',
    headline: 'A Go web toolkit that stays close to net/http.',
    summary:
      'Plumego is a stdlib-first toolkit for building explicit HTTP services in Go. Routing, middleware, transport contracts, and application wiring stay visible in code instead of being hidden behind framework conventions.',
    primaryCta: { label: 'Get Started', href: '/docs/getting-started' },
    secondaryCta: { label: 'View Reference App', href: '/docs/reference-app' },
    notes: ['Go 1.24+', 'stable roots + x/* extensions', 'agent-friendly repo control plane'],
    heroVisual: {
      stableLabel: 'stable roots',
      stableTitle: 'small public surface',
      stableBody: 'core, router, contract, and the rest of the long-lived API stay narrow on purpose.',
      extensionLabel: 'extensions',
      extensionTitle: 'optional capability families',
      extensionBody: 'x/* keeps fast-moving features out of the kernel and out of the default learning path.',
      footerLabel: 'reading path',
      footerBody: 'reference app → module boundary → extension family',
    },
    valueTitle: 'Built for explicit services, not framework magic.',
    valueLead:
      'Each section below describes a deliberate repository choice: keep the HTTP model familiar, keep boundaries visible, and keep control-plane intent inspectable.',
    values: [
      {
        kicker: 'HTTP model',
        title: 'Stay on the standard library',
        body: 'Plumego keeps the net/http model intact so handlers, middleware, and tests remain familiar.',
      },
      {
        kicker: 'Ownership',
        title: 'Keep boundaries visible',
        body: 'Stable roots own narrow responsibilities. Fast-moving capabilities live under x/* instead of leaking into the kernel.',
      },
      {
        kicker: 'Workflow',
        title: 'Make repository intent legible',
        body: 'Docs, specs, manifests, and task cards make ownership easier to classify for both humans and coding agents.',
      },
    ],
    mapTitle: 'One stable core, optional capability packs.',
    mapLead:
      'The public surface stays small on purpose. Stable roots define the long-lived API, while x/* families carry optional or faster-evolving features.',
    mapPanels: {
      stable: {
        title: 'Stable Roots',
        caption: 'Long-lived public packages with explicit responsibilities.',
      },
      extension: {
        title: 'x/* Extensions',
        caption: 'Optional or faster-moving families that should not redefine the core path.',
      },
      footer: 'Default routing rule: kernel and transport changes stay in stable roots; capability work starts in the owning x/* family.',
    },
    stableRoots: MODULE_FACTS.stableRoots,
    extensions: MODULE_FACTS.primaryExtensionFamilies,
    canonicalTitle: 'Start from one canonical path.',
    canonicalBody:
      'Begin with reference/standard-service. It shows the directory shape, bootstrap flow, and route wiring Plumego treats as the canonical application layout.',
    canonicalCta: { label: 'Read the Reference App Guide', href: '/docs/reference-app' },
    canonicalSteps: [
      {
        label: 'bootstrap',
        title: 'Read main.go first',
        body: 'Confirm where the application starts and how the server is assembled.',
      },
      {
        label: 'wiring',
        title: 'Open internal/app/app.go',
        body: 'Inspect the app-local constructor and keep dependencies explicit.',
      },
      {
        label: 'routes',
        title: 'Finish in internal/app/routes.go',
        body: 'Check route registration before touching handlers or deeper modules.',
      },
    ],
    controlTitle: 'Docs for humans. Specs for tools. Tasks for execution.',
    controlBody:
      'Plumego keeps prose, machine-readable rules, and execution cards separate so repository changes can stay inside explicit boundaries.',
    controlItems: [
      {
        title: 'docs/',
        body: 'Explain intent, architecture, and the canonical path.',
      },
      {
        title: 'specs/',
        body: 'Carry machine-readable ownership, routing, and dependency rules.',
      },
      {
        title: 'tasks/',
        body: 'Define execution cards and milestone sequencing for change work.',
      },
    ],
    finalTitle: 'Read the docs, inspect the reference app, then build your own service.',
    finalBody:
      'The shortest path into Plumego is to follow the documented reference flow first and only expand outward when the ownership boundary is clear.',
    finalPrimary: { label: 'Read Docs', href: '/docs' },
    finalSecondary: { label: 'Open GitHub', href: SITE.githubUrl },
  },
  zh: {
    eyebrow: 'Standard Library First',
    headline: '一个尽量贴近 net/http 的 Go Web 工具包。',
    summary:
      'Plumego 是一个 stdlib-first 的 Go HTTP 工具包。它让路由、中间件、传输 contract 与应用 wiring 保持显式，而不是把关键控制流藏进框架约定里。',
    primaryCta: { label: '开始使用', href: '/zh/docs/getting-started' },
    secondaryCta: { label: '查看参考应用', href: '/zh/docs/reference-app' },
    notes: ['Go 1.24+', '稳定根 + x/* 扩展', 'agent-friendly 仓库控制面'],
    heroVisual: {
      stableLabel: 'stable roots',
      stableTitle: '收敛的公开表面',
      stableBody: 'core、router、contract 等长期 API 有意保持窄边界，而不是一路长成框架能力目录。',
      extensionLabel: 'extensions',
      extensionTitle: '可选能力家族',
      extensionBody: 'x/* 承接快速演进功能，让内核与默认学习路径保持克制。',
      footerLabel: 'reading path',
      footerBody: 'reference app → 模块边界 → 扩展家族',
    },
    valueTitle: '服务要显式，不要框架魔法。',
    valueLead:
      '下面每个区块都对应一个明确的工程决策：HTTP 模型保持熟悉，模块边界保持可见，仓库控制面保持可判断。',
    values: [
      {
        kicker: 'HTTP 模型',
        title: '默认站在标准库上',
        body: '保留 net/http 的心智模型，让 handler、中间件与测试方式都保持熟悉。',
      },
      {
        kicker: 'Ownership',
        title: '边界保持可见',
        body: '稳定根只承担窄职责，快速演进能力统一放在 x/*，不把内核做成能力目录。',
      },
      {
        kicker: 'Workflow',
        title: '让仓库意图更容易判断',
        body: 'docs、specs、manifest 与 task card 分层明确，方便人和 agent 判断 ownership 并做小步改动。',
      },
    ],
    mapTitle: '稳定核心一层，能力扩展一层。',
    mapLead: '公开表面保持收敛是有意设计。稳定根承接长期 API，x/* 家族承接可选能力与快速演进功能。',
    mapPanels: {
      stable: {
        title: '稳定根',
        caption: '长期公开包，职责显式且边界窄。',
      },
      extension: {
        title: 'x/* 扩展',
        caption: '可选能力与快速演进家族，不重定义核心学习路径。',
      },
      footer: '默认路由规则：内核与 transport 变化归稳定根；能力工作从对应的 x/* 家族开始。',
    },
    stableRoots: MODULE_FACTS.stableRoots,
    extensions: MODULE_FACTS.primaryExtensionFamilies,
    canonicalTitle: '先从唯一的 canonical 路径起步。',
    canonicalBody:
      '先看 reference/standard-service。它定义了 Plumego 认可的目录结构、bootstrap 流程与 route wiring 方式。',
    canonicalCta: { label: '阅读参考应用说明', href: '/zh/docs/reference-app' },
    canonicalSteps: [
      {
        label: 'bootstrap',
        title: '先读 main.go',
        body: '确认应用从哪里启动，以及 server 是怎样被组装起来的。',
      },
      {
        label: 'wiring',
        title: '再看 internal/app/app.go',
        body: '检查应用本地 wiring，保持依赖显式。',
      },
      {
        label: 'routes',
        title: '最后看 internal/app/routes.go',
        body: '在继续深入 handler 或模块之前，先看清 route 注册。',
      },
    ],
    controlTitle: '给人看的 docs，给工具看的 specs，给执行的 tasks。',
    controlBody: 'Plumego 把说明、机器可读规则与执行卡拆开，让仓库变更更容易落在明确边界内。',
    controlItems: [
      {
        title: 'docs/',
        body: '解释设计意图、架构与 canonical 路径。',
      },
      {
        title: 'specs/',
        body: '承载机器可读的 ownership、路由与依赖规则。',
      },
      {
        title: 'tasks/',
        body: '定义执行卡与 milestone 顺序，约束实际改动过程。',
      },
    ],
    finalTitle: '先读文档，理解参考应用，再写你自己的服务。',
    finalBody: '进入 Plumego 的最短路径，是先遵循已定义的参考流程，再在 ownership 清晰的前提下逐步展开。',
    finalPrimary: { label: '阅读文档', href: '/zh/docs' },
    finalSecondary: { label: '打开 GitHub', href: SITE.githubUrl },
  },
} as const;

export const ROADMAP_COPY = {
  en: {
    title: 'Roadmap',
    description: 'What Plumego is hardening now, what comes next, and what stays out of scope.',
    sections: [
      {
        title: 'Current baseline',
        items: [
          'stable roots with explicit boundaries',
          'canonical reference app under reference/standard-service',
          'repo-native control plane under docs/, specs/, and tasks/',
        ],
      },
      {
        title: 'In progress',
        items: [
          'x/ai stabilization and runnable examples',
          'docs and onboarding sync for v1 users',
          'deeper extension test coverage where compatibility is still experimental',
        ],
      },
      {
        title: 'Planned',
        items: [
          'x/tenant production-readiness hardening',
          'clearer x/data and x/fileapi operational guidance',
          'extension stability criteria before any promotion discussion',
        ],
      },
      {
        title: 'Non-goals',
        items: [
          'reintroducing hidden compatibility layers',
          'moving tenant or data-topology logic back into stable roots',
          'marking x/* as GA without policy, tests, and docs',
        ],
      },
    ],
  },
  zh: {
    title: '路线图',
    description: 'Plumego 当前在补什么、下一步做什么，以及明确不做什么。',
    sections: [
      {
        title: '当前基线',
        items: [
          '稳定根已经建立清晰边界',
          'reference/standard-service 作为 canonical 参考应用',
          'docs/、specs/、tasks/ 已形成仓库控制面',
        ],
      },
      {
        title: '进行中',
        items: [
          'x/ai 的稳定化与可运行示例',
          'v1 用户视角的文档与 onboarding 同步',
          '仍属 experimental 的扩展模块补强测试深度',
        ],
      },
      {
        title: '计划中',
        items: [
          'x/tenant 的生产可用性加强',
          'x/data 与 x/fileapi 的运维指引补齐',
          '在扩展晋级前先定义统一稳定性标准',
        ],
      },
      {
        title: '明确不做',
        items: [
          '重新引入隐藏兼容层',
          '把租户或数据拓扑逻辑塞回稳定根',
          '在缺少策略、测试和文档时把 x/* 标成 GA',
        ],
      },
    ],
  },
} as const;

export const RELEASE_COPY = {
  en: {
    title: 'Releases',
    description: 'Release posture, compatibility expectations, and the current support matrix.',
  },
  zh: {
    title: '发布',
    description: '发布姿态、兼容性承诺与当前支持矩阵。',
  },
} as const;

export const SUPPORT_MATRIX = RELEASE_FACTS.supportMatrix;
