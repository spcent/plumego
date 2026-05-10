import { MODULE_FACTS } from '../generated/modules';
import { RELEASE_FACTS } from '../generated/releases';

export type Locale = 'en' | 'zh';

export const SITE = {
  name: 'Plumego',
  githubUrl: 'https://github.com/spcent/plumego',
  repoPath: 'reference/standard-service',
  currentVersion: RELEASE_FACTS.currentVersion,
};

export const SITE_COPY: Record<Locale, { footerTagline: string }> = {
  en: {
    footerTagline: 'stdlib-first Go HTTP toolkit for explicit services.',
  },
  zh: {
    footerTagline: 'Go 服务显式 wiring 工具包，stdlib 优先。',
  },
};

export const NAV_LINKS: Record<Locale, Array<{ label: string; href: string }>> = {
  en: [
    { label: 'Docs', href: '/docs' },
    { label: 'Why Plumego', href: '/why-plumego' },
    { label: 'Examples', href: '/examples' },
    { label: 'Releases', href: '/releases' },
  ],
  zh: [
    { label: '文档', href: '/zh/docs' },
    { label: '为什么选择', href: '/zh/why-plumego' },
    { label: '示例', href: '/zh/examples' },
    { label: '发布', href: '/zh/releases' },
  ],
};

export const FOOTER_GROUPS: Record<Locale, Array<{ title: string; links: Array<{ label: string; href: string }> }>> = {
  en: [
    {
      title: 'Product',
      links: [
        { label: 'Docs', href: '/docs' },
        { label: 'Why Plumego', href: '/why-plumego' },
        { label: 'Architecture', href: '/architecture' },
        { label: 'Examples', href: '/examples' },
      ],
    },
    {
      title: 'Start',
      links: [
        { label: 'Getting Started', href: '/docs/getting-started' },
        { label: 'Reference App', href: '/docs/reference-app' },
        { label: 'FAQ', href: '/docs/faq' },
      ],
    },
    {
      title: 'Status',
      links: [
        { label: 'Roadmap', href: '/roadmap' },
        { label: 'Releases', href: '/releases' },
        { label: 'Changelog', href: `${SITE.githubUrl}/releases` },
        { label: 'GitHub', href: SITE.githubUrl },
        { label: 'Issues', href: `${SITE.githubUrl}/issues` },
      ],
    },
  ],
  zh: [
    {
      title: '产品',
      links: [
        { label: '文档', href: '/zh/docs' },
        { label: '为什么选择 Plumego', href: '/zh/why-plumego' },
        { label: '架构', href: '/zh/architecture' },
        { label: '示例', href: '/zh/examples' },
      ],
    },
    {
      title: '起步',
      links: [
        { label: '开始使用', href: '/zh/docs/getting-started' },
        { label: '参考应用', href: '/zh/docs/reference-app' },
        { label: '常见问题', href: '/zh/docs/faq' },
      ],
    },
    {
      title: '状态',
      links: [
        { label: '路线图', href: '/zh/roadmap' },
        { label: '发布', href: '/zh/releases' },
        { label: 'Changelog', href: `${SITE.githubUrl}/releases` },
        { label: 'GitHub', href: SITE.githubUrl },
        { label: 'Issues', href: `${SITE.githubUrl}/issues` },
      ],
    },
  ],
};

export const HOME_COPY = {
  en: {
    eyebrow: 'Go 1.24 · net/http compatible',
    headline: 'Go services that stay reviewable as your team grows.',
    summary:
      'Most frameworks hide routing and wiring behind conventions. Plumego keeps them visible in code — so route ownership, middleware order, and dependency entry points stay easy to inspect in review.',
    primaryCta: { label: 'Get Started', href: '/docs/getting-started' },
    secondaryCta: { label: 'See Architecture', href: '/architecture' },
    notes: ['Go 1.24+', 'net/http compatible', 'no hidden registration', 'one canonical service shape'],
    heroAudience: {
      label: 'Who it fits',
      title: 'Teams building internal APIs or platform services',
      body: 'Best when route ownership, middleware order, and dependency wiring should stay visible in code review — not hidden behind framework-managed registration.',
    },
    heroStatus: {
      label: 'Pre-v1',
      title: `${RELEASE_FACTS.currentVersion} — pre-v1, stable roots ready to evaluate`,
      body: 'Pre-v1: stable root APIs may change until v1 is released. Stable roots are the first production evaluation scope. Extension families (x/*) carry explicit maturity labels and are experimental unless stated otherwise.',
    },
    heroVisual: {
      stableLabel: 'one stable model',
      stableTitle: 'explicit request path',
      stableBody: 'Routes, middleware, handlers, and application wiring remain visible in code instead of disappearing into framework-owned registration.',
      extensionLabel: 'one expansion rule',
      extensionTitle: 'capabilities branch outward',
      extensionBody: 'Stable roots stay narrow while optional capability work starts in x/* families instead of stretching the kernel.',
      footerLabel: 'default path',
      footerBody: 'docs → reference app → request flow → maturity check → module fit',
    },
    heroPath: [
      {
        label: '01',
        title: 'Start in docs',
        body: 'Use getting-started to understand the default path before browsing deeper packages.',
      },
      {
        label: '02',
        title: 'Run the reference app',
        body: 'Treat reference/standard-service as the canonical service shape instead of a disposable demo.',
      },
      {
        label: '03',
        title: 'Trace one request',
        body: 'Read the request flow before deciding where transport responsibility ends.',
      },
      {
        label: '04',
        title: 'Read release posture deliberately',
        body: 'Check which modules are stable, which are supported references, and which are still experimental before widening adoption scope.',
      },
      {
        label: '05',
        title: 'Only then classify module ownership',
        body: 'Decide whether the next change belongs in stable core modules, app-local wiring, or extension modules.',
      },
    ],
    valueTitle: 'What this approach changes.',
    valueLead:
      'Not to add more abstraction — to keep the important structure visible after the codebase stops being small and reviewers need a shared model to work from.',
    values: [
      {
        icon: '🔍',
        kicker: 'visible wiring',
        title: 'Routes, middleware, and dependencies stay in code',
        body: 'When a new engineer opens the service, they can trace exactly how it starts, which middleware runs, and where dependencies enter — without reading framework source.',
      },
      {
        icon: '🎯',
        kicker: 'narrow surface',
        title: 'The stable surface stays intentionally small',
        body: 'Core modules carry strong compatibility guarantees. New capabilities start in extension families so the kernel does not absorb every fast-moving concern.',
      },
      {
        icon: '✅',
        kicker: 'adoption signals',
        title: 'Know what is safe to adopt before committing',
        body: 'Compatibility boundaries are documented, not inferred from package existence. Teams can distinguish a stable baseline from an experimental family without guessing.',
      },
      {
        icon: '🗂️',
        kicker: 'repo clarity',
        title: 'Change ownership stays easy to classify',
        body: 'When a PR touches middleware, a handler, and a new module, the docs and specs structure gives reviewers a shared language for deciding where each piece belongs.',
      },
      {
        icon: '🔌',
        kicker: 'stdlib first',
        title: 'net/http compatible — no framework lock-in',
        body: 'Handlers are plain func(http.ResponseWriter, *http.Request). Existing stdlib middleware, clients, and test helpers work without adapters or wrappers.',
      },
      {
        icon: '📐',
        kicker: 'canonical shape',
        title: 'One reference service shape for the whole team',
        body: 'reference/standard-service defines the default wiring pattern so every service in your org starts from the same baseline instead of reinventing structure each time.',
      },
    ],
    valueFootnote:
      'The point is not more framework. The point is a toolkit that still reads clearly after the codebase stops being small and multiple teams need one defensible default path.',
    adoptionTitle: 'Choose the next page by question, not by package.',
    adoptionBody:
      'Why Plumego answers whether Plumego fits your team. Examples walk through a runnable service. Releases tells you what is stable versus experimental. Pick the page that answers the question in front of you.',
    adoptionCards: [
      {
        kicker: 'fit',
        title: 'Why Plumego',
        body: 'Start here when the question is whether Plumego fits your team, your service shape, and your review expectations — before investing time in the technical details.',
        href: '/why-plumego',
        label: 'Evaluate fit',
      },
      {
        kicker: 'examples',
        title: 'Examples',
        body: 'Start from one runnable reference service, then follow the guided recipes to understand request flow, middleware ordering, and how extensions attach.',
        href: '/examples',
        label: 'Explore examples',
      },
      {
        kicker: 'maturity',
        title: 'Releases',
        body: 'Use the release page when the real question is not whether a module exists, but whether it is stable enough to adopt in production right now.',
        href: '/releases',
        label: 'Inspect releases',
      },
      {
        kicker: 'questions',
        title: 'FAQ',
        body: 'Common questions answered directly: how to connect a database, add JWT auth, compare with Gin or Echo, handle errors, and configure for production.',
        href: '/docs/faq',
        label: 'Read FAQ',
      },
    ],
    stableRoots: MODULE_FACTS.stableRoots,
    extensions: MODULE_FACTS.primaryExtensionFamilies,
    canonicalTitle: 'Start from one canonical request path.',
    canonicalBody:
      'Begin with reference/standard-service. Read main.go to see how the server is assembled, check app.go for explicit dependency wiring, then finish in routes.go to understand what is public and where transport control ends.',
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
        body: 'Check route registration to understand which paths are public, what middleware runs, and where transport control ends.',
      },
    ],
    mapTitle: 'Use the module map after the canonical path is clear.',
    mapLead:
      'The module map is not a second starting point. It is the next layer of orientation once you understand the default request path, release posture, and need to classify where deeper work belongs.',
    mapPanels: {
      stable: {
        title: 'Stable roots protect the default path',
        caption: 'Narrow modules with stronger compatibility burden, clearer review expectations, and a published baseline role.',
      },
      extension: {
        title: 'Extension modules protect optional capability work',
        caption: 'Product or protocol-specific work moves outward instead of stretching the kernel, and stays easier to evaluate deliberately.',
      },
      footer: 'Read the path first; use the module map when the next question is ownership or module maturity.',
    },
    finalTitle: 'Read the docs, run the reference app, then expand only when the boundary is clear.',
    finalBody:
      'Plumego works best when you start from the default path, check module maturity early, and branch into deeper boundaries only after the owning capability is obvious.',
    finalPrimary: { label: 'Read Docs', href: '/docs' },
    finalSecondary: { label: 'See Reference App', href: '/docs/reference-app' },
    contrastTitle: 'The difference shows in code review.',
    contrastLead:
      'When routes are spread across packages, a reviewer has to open each one to understand what paths exist and what middleware runs. Plumego keeps the full route map in one explicit file — and adds a structured contract layer so error and success responses stay consistent across all handlers.',
    contrastBeforeLabel: 'routes split across packages',
    contrastAfterLabel: 'plumego: one file, one contract',
  },
  zh: {
    eyebrow: 'Go 1.24 · net/http 兼容',
    headline: '仓库变大，服务代码依然可以评审。',
    summary:
      '大多数框架把路由和依赖注入藏在约定背后。Plumego 把它们留在代码里 —— 显式、可检查，在仓库和团队规模变大以后依然容易在评审中追溯。',
    primaryCta: { label: '开始使用', href: '/zh/docs/getting-started' },
    secondaryCta: { label: '查看架构', href: '/zh/architecture' },
    notes: ['Go 1.24+', 'net/http 兼容', '零隐藏注册', '一套参考服务形态'],
    heroAudience: {
      label: '适合场景',
      title: '构建内部 API 或平台服务的团队',
      body: '尤其适合需要在代码评审中直接看清 route ownership、中间件顺序和依赖 wiring 的场景，而不是把这些结构藏在框架注册机制后面。',
    },
    heroStatus: {
      label: 'Pre-v1',
      title: `${RELEASE_FACTS.currentVersion} —— pre-v1，稳定根可先评估`,
      body: 'Pre-v1：v1 发布前稳定根 API 可能调整。稳定根是最适合先进入生产评估的范围。扩展家族（x/*）明确标注成熟度，未标注为稳定则视为实验性。',
    },
    heroVisual: {
      stableLabel: '一套稳定模型',
      stableTitle: '显式请求路径',
      stableBody: 'Routes、middleware、handlers 和应用 wiring 都继续留在代码里，而不是消失在框架注册机制后面。',
      extensionLabel: '一条扩展规则',
      extensionTitle: '能力向外分叉',
      extensionBody: '稳定根保持收敛，可选能力从 x/* 家族开始，而不是继续拉长内核。',
      footerLabel: '默认路径',
      footerBody: 'docs → reference app → request flow → 成熟度确认 → 模块归属',
    },
    heroPath: [
      {
        label: '01',
        title: '先从文档进入',
        body: '先用开始使用理解默认路径，再决定要不要看更深的包。',
      },
      {
        label: '02',
        title: '运行 reference app',
        body: '把 reference/standard-service 当成 canonical 服务形态，而不是一次性 demo。',
      },
      {
        label: '03',
        title: '沿 request flow 读一遍',
        body: '先追清一条请求，判断 transport 责任到底在哪一层结束。',
      },
      {
        label: '04',
        title: '先检查发布姿态',
        body: '在扩大采用范围之前，先确认哪些模块稳定可用，哪些仍然只是支持参考或实验性能力。',
      },
      {
        label: '05',
        title: '最后判断模块归属',
        body: '再决定这项工作属于稳定核心模块、应用本地依赖注入，还是扩展模块。',
      },
    ],
    valueTitle: '这个方案改变了什么。',
    valueLead:
      '不是为了增加抽象——而是在仓库不再小、评审者需要一个共同模型时，把重要结构继续保持可见。',
    values: [
      {
        icon: '🔍',
        kicker: '可见 wiring',
        title: 'routes、middleware 和依赖都留在代码里',
        body: '当新工程师打开服务时，他们可以直接追溯服务如何启动、哪些中间件在运行、依赖从哪里进入——不需要先读框架源码。',
      },
      {
        icon: '🎯',
        kicker: '收敛的表面',
        title: '稳定表面保持刻意收敛',
        body: '核心模块承担强兼容性保证。新能力从扩展家族起步，这样内核就不会吸入每一项快速演进的工作。',
      },
      {
        icon: '✅',
        kicker: '采用信号',
        title: '在做决定前知道哪些可以安全采用',
        body: '兼容性边界有文档，而不是靠包是否存在来推断。团队可以一眼分清稳定基线和实验性家族，而不需要猜测。',
      },
      {
        icon: '🗂️',
        kicker: '仓库清晰度',
        title: '改动归属保持容易判断',
        body: '当一个 PR 同时涉及 middleware、handler 和新模块时，docs/specs 的分层结构给评审者提供共同语言，帮助决定每个部分该落在哪里。',
      },
      {
        icon: '🔌',
        kicker: 'stdlib 优先',
        title: 'net/http 兼容，不锁定框架',
        body: 'Handler 就是普通的 func(http.ResponseWriter, *http.Request)，现有的 stdlib 中间件、客户端和测试工具不需要任何适配器就能直接使用。',
      },
      {
        icon: '📐',
        kicker: 'canonical 形态',
        title: '全团队共用一套参考服务形态',
        body: 'reference/standard-service 定义了默认的 wiring 模式，让团队中每一个服务都从同一条基线出发，而不是每次都重新摸索结构。',
      },
    ],
    valueFootnote:
      '重点不是更多框架抽象。而是在仓库不再小、多个团队需要一条可辩护的默认路径时，工具包依然读起来清晰。',
    adoptionTitle: '根据问题选下一页，而不是先选包。',
    adoptionBody:
      '为什么选 Plumego 回答工具包是否适合你的团队，示例展示可运行的服务，发布页告诉你哪些稳定、哪些还在实验阶段。先选那个真正回答你当前问题的页面。',
    adoptionCards: [
      {
        kicker: 'fit',
        title: '为什么选择 Plumego',
        body: '当问题是 Plumego 是否适合你的团队、服务形态和评审预期——而不是技术细节如何使用时，从这里开始。',
        href: '/zh/why-plumego',
        label: '判断是否适合',
      },
      {
        kicker: 'examples',
        title: '示例',
        body: '从一个可运行的参考服务出发，通过引导式示例了解请求路径、中间件排序以及扩展模块如何接入。',
        href: '/zh/examples',
        label: '查看示例',
      },
      {
        kicker: 'maturity',
        title: '发布',
        body: '当真正的问题不是某个模块是否存在，而是它今天是否已经稳定到可以用于生产，就去看发布页。',
        href: '/zh/releases',
        label: '查看发布页',
      },
      {
        kicker: 'questions',
        title: '常见问题',
        body: '直接回答常见问题：如何连接数据库、添加 JWT 认证、与 Gin 或 Echo 对比、处理错误以及生产环境配置。',
        href: '/zh/docs/faq',
        label: '查看常见问题',
      },
    ],
    stableRoots: MODULE_FACTS.stableRoots,
    extensions: MODULE_FACTS.primaryExtensionFamilies,
    canonicalTitle: '先从一条 canonical 请求路径起步。',
    canonicalBody:
      '先看 reference/standard-service。读 main.go 了解 server 如何组装，看 app.go 确认依赖显式，再看 routes.go 了解哪些路径是公开的以及 transport 控制在哪里结束。',
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
        title: '然后看 internal/app/routes.go',
        body: '看清哪些路径是公开的、哪些中间件在运行，以及 transport 控制在哪里结束。',
      },
    ],
    mapTitle: '只有 canonical path 清楚以后，再看 module map。',
    mapLead:
      '模块地图不是第二个起点，而是在默认请求路径已经读清之后，帮助你继续判断更深工作归属的下一层视图。',
    mapPanels: {
      stable: {
        title: '稳定根负责保护默认路径',
        caption: '这些模块更收敛、兼容性负担更重，也承担公开基线的角色。',
      },
      extension: {
        title: '扩展模块负责承接可选能力工作',
        caption: '产品能力或协议工作向外分叉，而不是继续把内核拉长，也更适合被审慎地单独评估。',
      },
      footer: '先读路径，再用 module map 判断归属与模块成熟度。',
    },
    finalTitle: '先读文档，跑通 reference app，再在边界清楚时向外扩展。',
    finalBody: 'Plumego 最适合从默认路径进入，并尽早确认模块成熟度；只有在归属已经明确之后，再进入更深的模块边界。',
    finalPrimary: { label: '阅读文档', href: '/zh/docs' },
    finalSecondary: { label: '查看参考应用', href: '/zh/docs/reference-app' },
    contrastTitle: '差异在代码评审时最明显。',
    contrastLead:
      '当路由分散在各个包里时，评审者必须逐个打开才能知道有哪些路径和中间件在运行。Plumego 把完整路由表放在一个显式文件里——同时加入结构化的 contract 层，让所有 handler 的错误响应和成功响应保持一致。',
    contrastBeforeLabel: '路由分散在各包里',
    contrastAfterLabel: 'plumego：一个文件，一套 contract',
  },
} as const;

export const ROADMAP_COPY = {
  en: {
    title: 'Roadmap',
    description: 'What Plumego is hardening now, what comes next, and what stays out of scope.',
    eyebrow: 'Repository Direction',
    introTitle: 'How to read the roadmap.',
    introBody:
      'The roadmap is not a promise that every package moves at the same speed. It explains what is already treated as baseline, which areas are being hardened now, and which ambitions are still gated by policy, tests, and documentation.',
    guideCards: [
      {
        kicker: 'baseline',
        title: 'Read the baseline first',
        body: 'The baseline tells you what already defines the default Plumego path and should not be inferred from planned work.',
      },
      {
        kicker: 'focus',
        title: 'Use in-progress work as the real near-term signal',
        body: 'If an area is in progress, it is receiving active attention now. Planned work is directional, not a compatibility statement.',
      },
      {
        kicker: 'non-goals',
        title: 'Treat non-goals as architectural boundaries',
        body: 'The non-goals are as important as the roadmap itself because they prevent the repo from drifting back into hidden compatibility layers.',
      },
    ],
    timelineTitle: 'Current direction',
    timelineBody:
      'These sections are synchronized from the repository facts, then framed here so a reader can distinguish established baseline, active work, and deferred scope.',
    orderTitle: 'Suggested execution order',
    orderBody:
      'When multiple lines of work are open, this sequence shows the intended tightening order rather than a guarantee of exact release timing.',
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
    eyebrow: 'Repository Direction',
    introTitle: '怎么阅读这份路线图。',
    introBody:
      '路线图不是在暗示所有包会以同样速度推进。它要说明的是：哪些已经属于默认基线，哪些正在补强，哪些仍然只是方向而不是兼容性承诺。',
    guideCards: [
      {
        kicker: 'baseline',
        title: '先看基线',
        body: '基线说明哪些东西已经构成默认 Plumego 路径，不应该从计划项里反向推断。',
      },
      {
        kicker: 'focus',
        title: '进行中才代表近期真实信号',
        body: '如果某个方向处于进行中，说明它正在被主动补强。计划中的内容是方向，不是兼容性保证。',
      },
      {
        kicker: 'non-goals',
        title: '把非目标当作架构边界',
        body: '非目标和路线图本身一样重要，因为它们约束仓库不要重新滑回隐藏兼容层。',
      },
    ],
    timelineTitle: '当前方向',
    timelineBody:
      '下面这些分组由仓库事实同步而来，这里再补一层解释，帮助读者分清哪些是已建立基线、哪些是当前工作、哪些仍是后续范围。',
    orderTitle: '建议执行顺序',
    orderBody:
      '当多条工作线同时存在时，这个顺序表达的是优先收敛方向，而不是精确发布时间承诺。',
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

export const ARCHITECTURE_COPY = {
  en: {
    title: 'Architecture',
    description: 'Stable roots, extension families, and the canonical request path that define the Plumego mental model.',
    eyebrow: 'System Topology',
    introTitle: 'Read the architecture as a boundary map.',
    introBody:
      'Plumego is easiest to evaluate when you can see which responsibilities stay in the kernel, which ones branch outward into extensions, and where the default request path begins and ends.',
    guideCards: [
      {
        kicker: 'stable roots',
        title: 'Keep long-lived responsibilities narrow',
        body: 'Routing, contracts, transport middleware, and storage-facing primitives should stay legible enough to defend compatibility over time.',
      },
      {
        kicker: 'extension rule',
        title: 'Push optional capability work outward',
        body: 'Product or protocol-specific expansion should begin in x/* families so the kernel does not absorb every fast-moving concern.',
      },
      {
        kicker: 'canonical path',
        title: 'Teach one request path first',
        body: 'The website, docs, and reference app all point toward one readable route from bootstrap to write path before asking readers to branch into deeper packages.',
      },
    ],
    sectionTitle: 'Three layers you should be able to classify immediately',
    sectionBody:
      'If a new reader cannot tell whether a change belongs to stable roots, a capability family, or the canonical app path, the site is not doing enough explanatory work.',
    layers: [
      {
        kicker: 'kernel',
        title: 'Stable roots',
        body: 'Core modules carry the strongest compatibility burden. They should stay narrow, boring, and easy to reason about in code review.',
        items: MODULE_FACTS.stableRoots,
      },
      {
        kicker: 'extensions',
        title: 'Primary capability families',
        body: 'These families are useful, but they do not pretend to share the same stability profile as the smallest kernel surface.',
        items: MODULE_FACTS.primaryExtensionFamilies,
      },
      {
        kicker: 'workflow',
        title: 'Canonical request path',
        body: 'The default app path gives teams one bootstrap model, one routing flow, and one shared place to begin before extension work starts.',
        items: ['docs/getting-started', 'docs/reference-app', 'internal/app/app.go', 'internal/app/routes.go'],
      },
    ],
    flowTitle: 'How to inspect one request without guessing',
    flowBody:
      'The quickest architecture read starts at the reference app, confirms the app-local wiring, and only then expands outward into packages that obviously belong to the feature or protocol being added.',
    flowSteps: [
      {
        label: '01',
        title: 'Start in the reference app',
        body: 'Read the default service shape first so bootstrap and route ownership stay visible.',
      },
      {
        label: '02',
        title: 'Confirm the app-local constructor',
        body: 'Check where dependencies are assembled before treating helpers or packages as architectural entry points.',
      },
      {
        label: '03',
        title: 'Trace routes before handlers',
        body: 'Route registration should tell you which request path is public, what middleware runs, and where transport control ends.',
      },
    ],
  },
  zh: {
    title: '架构',
    description: '帮助读者看清 Plumego 的稳定根、扩展家族与 canonical 请求路径如何构成统一心智模型。',
    eyebrow: 'System Topology',
    introTitle: '把架构页当作边界地图来读。',
    introBody:
      '判断 Plumego 最容易的方式，是先看清哪些职责留在内核、哪些能力向 x/* 扩展家族外分，以及默认请求路径从哪里开始、在什么地方结束。',
    guideCards: [
      {
        kicker: 'stable roots',
        title: '让长期职责保持收敛',
        body: '路由、契约、传输中间件以及面向存储的基础原语，应当保持足够清晰，才能长期守住兼容性。',
      },
      {
        kicker: 'extension rule',
        title: '把可选能力向外扩展',
        body: '产品能力或协议适配应该从 x/* 家族开始，而不是不断吸入内核，最终让默认路径失去边界。',
      },
      {
        kicker: 'canonical path',
        title: '先教会读者一条默认请求路径',
        body: '网站、文档和 reference app 都应该先把从 bootstrap 到 write path 的默认路径讲清楚，再让读者进入更深的扩展区域。',
      },
    ],
    sectionTitle: '读者应该能立刻分清的三层结构',
    sectionBody:
      '如果一个新读者分不清某个改动属于稳定根、能力家族还是 canonical 应用路径，说明站点还没有把架构解释到位。',
    layers: [
      {
        kicker: 'kernel',
        title: '稳定根',
        body: '核心模块承担最强的兼容性负担，因此应保持收敛、克制，并且在代码评审时容易理解。',
        items: MODULE_FACTS.stableRoots,
      },
      {
        kicker: 'extensions',
        title: '主要能力家族',
        body: '这些扩展很重要，但不应该假装与最小内核表面享受同一等级的稳定性承诺。',
        items: MODULE_FACTS.primaryExtensionFamilies,
      },
      {
        kicker: 'workflow',
        title: 'canonical 请求路径',
        body: '默认应用路径给团队提供统一 bootstrap 模型、统一路由流向，以及开始理解仓库的共同入口。',
        items: ['docs/getting-started', 'docs/reference-app', 'internal/app/app.go', 'internal/app/routes.go'],
      },
    ],
    flowTitle: '不靠猜测地检查一条请求路径',
    flowBody:
      '最快的架构阅读方式，是先从 reference app 入手，确认 app-local wiring，然后再向外展开到那些显然属于具体功能或协议的包。',
    flowSteps: [
      {
        label: '01',
        title: '先从 reference app 开始',
        body: '先看默认服务形态，把 bootstrap 和 route ownership 读清楚。',
      },
      {
        label: '02',
        title: '确认应用本地构造器',
        body: '先检查依赖是如何被组装的，再判断辅助包是不是架构入口。',
      },
      {
        label: '03',
        title: '先追 routes，再进 handlers',
        body: 'route 注册应该先告诉你公开请求路径、middleware 顺序以及 transport 控制在哪里停止。',
      },
    ],
  },
} as const;

export const RELEASE_COPY = {
  en: {
    title: 'Releases',
    description: 'Release posture, compatibility expectations, and the current support matrix.',
    eyebrow: 'Release Posture',
    introTitle: 'How to read release posture.',
    introBody:
      'This page exists to make compatibility boundaries explicit. It should help you distinguish what is safe to adopt through the canonical path, what is still evolving, and where release ambition is intentionally held back by policy or test depth.',
    guideCards: [
      {
        kicker: 'canonical',
        title: 'Start from the canonical path',
        body: 'If you are evaluating Plumego for real use, begin with the reference app and the documented default path before reading optional surfaces as if they were all equally mature.',
      },
      {
        kicker: 'matrix',
        title: 'Read the support matrix as a boundary map',
        body: 'The matrix is not marketing copy. It tells you which areas carry stronger compatibility expectations and which still require caution.',
      },
      {
        kicker: 'roadmap',
        title: 'Separate release posture from roadmap ambition',
        body: 'A package can be interesting and actively developed without yet carrying the same stability promise as the stable roots.',
      },
    ],
    principlesTitle: 'Compatibility principles',
    principlesBody:
      'Release posture is framed by a few simple rules: the canonical path comes first, stable roots stay narrow, and optional capability families are allowed to move at different speeds.',
    principles: [
      {
        kicker: 'default path',
        title: 'The default path sets the bar',
        body: 'reference/standard-service and the stable-root path are the primary surfaces used to judge release readiness.',
      },
      {
        kicker: 'boundaries',
        title: 'x/* does not inherit stability automatically',
        body: 'Extension families are publishable, but they do not receive the same compatibility promise just by existing in the repository.',
      },
      {
        kicker: 'evidence',
        title: 'Compatibility claims follow evidence',
        body: 'Policy, tests, examples, and docs are required before a surface should be treated as frozen for compatibility.',
      },
    ],
    matrixTitle: 'Current support matrix',
    matrixBody:
      'The matrix below is synced from repository facts and is meant to be read as an adoption aid, not a branding grid.',
  },
  zh: {
    title: '发布',
    description: '发布姿态、兼容性承诺与当前支持矩阵。',
    eyebrow: 'Release Posture',
    introTitle: '怎么阅读发布姿态。',
    introBody:
      '这个页面的目标是把兼容性边界说清楚。你应该能借它分辨：哪些部分可以沿 canonical path 先采用，哪些能力仍在演进，以及哪些发布目标仍然被策略或测试深度明确卡住。',
    guideCards: [
      {
        kicker: 'canonical',
        title: '先从 canonical path 评估',
        body: '如果你准备真正使用 Plumego，先从 reference app 和已文档化的默认路径开始，不要把所有可选表面都当成同等成熟。',
      },
      {
        kicker: 'matrix',
        title: '把支持矩阵看成边界地图',
        body: '支持矩阵不是营销文案，它告诉你哪些区域承担更强兼容性预期，哪些仍然需要谨慎。',
      },
      {
        kicker: 'roadmap',
        title: '把发布姿态和路线图野心分开',
        body: '一个包可以很有价值、也在积极迭代，但这并不自动意味着它已经拥有和稳定根一样的稳定性承诺。',
      },
    ],
    principlesTitle: '兼容性原则',
    principlesBody:
      '发布姿态由几条简单规则约束：先看默认路径，稳定根保持收敛，可选能力家族允许按不同速度推进。',
    principles: [
      {
        kicker: 'default path',
        title: '默认路径决定发布门槛',
        body: 'reference/standard-service 和稳定根路径，是判断当前是否可用的主要依据。',
      },
      {
        kicker: 'boundaries',
        title: 'x/* 不会自动继承稳定性',
        body: '扩展家族可以被发布，但不会仅因为存在于仓库里就自动获得与稳定根相同的兼容性承诺。',
      },
      {
        kicker: 'evidence',
        title: '兼容性声明必须有证据支撑',
        body: '策略、测试、示例和文档都补齐之后，某个表面才应该被视为真正冻结。',
      },
    ],
    matrixTitle: '当前支持矩阵',
    matrixBody:
      '下面这张矩阵由仓库事实同步而来，应该被当成采用辅助工具，而不是品牌展示列表。',
  },
} as const;

export const SUPPORT_MATRIX = RELEASE_FACTS.supportMatrix;

export const USE_CASES_COPY = {
  en: {
    title: 'Adoption Fit',
    description: 'Use this page to decide whether Plumego matches the service in front of you and which reading path should come next.',
    eyebrow: 'Fit Check',
    introTitle: 'Use this page to decide whether Plumego matches the service in front of you.',
    introBody:
      'Architecture explains how the repository is structured. This page is narrower: it helps you decide whether Plumego is a good fit for the service you are evaluating and which entry path to open next.',
    guideCards: [
      {
        kicker: 'good fit',
        title: 'Boundary-sensitive Go services',
        body: 'Teams that need to review route wiring, handler shape, and ownership decisions directly in code benefit most from the canonical path.',
      },
      {
        kicker: 'default path',
        title: 'Internal services with one obvious layout',
        body: 'reference/standard-service gives new services a shared bootstrap and routing model instead of a repo-specific snowflake structure.',
      },
    ],
    decisionTitle: 'Three quick fit checks',
    decisionBody:
      'If your team wants the framework to disappear into conventions, Plumego may feel too explicit. If you want visible route decisions, one canonical service shape, and explicit maturity signals, it becomes much more compelling.',
    cases: [
      {
        kicker: 'internal api',
        title: 'Internal HTTP services with explicit review paths',
        body: 'Use Plumego when you want route registration, middleware order, and dependency wiring to stay easy to inspect in code review.',
        fitPoints: [
          'service boundaries are important to reviewers',
          'bootstrap and handler ownership should stay obvious',
          'the team already thinks in net/http terms',
        ],
        startHref: '/docs/getting-started',
        startLabel: 'Start with Getting Started',
      },
      {
        kicker: 'platform',
        title: 'Shared platform services that need one canonical shape',
        body: 'Plumego works well when multiple teams need the same service skeleton and you want drift to stay low across new services.',
        fitPoints: [
          'new services should start from one reference layout',
          'application-local wiring should remain explicit',
          'reviewers need a small set of accepted patterns',
        ],
        startHref: '/docs/reference-app',
        startLabel: 'Read the Reference App',
      },
      {
        kicker: 'extensions',
        title: 'Capability work that should stay out of the kernel',
        body: "When your repository has optional transport or product capability layers, Plumego's stable-root versus x/* split gives that work a clearer home.",
        fitPoints: [
          'optional capability work should not redefine the default path',
          'tenant, AI, gateway, or data logic should stay outside stable roots',
          'ownership matters more than convenience wrappers',
        ],
        startHref: '/docs/modules/overview',
        startLabel: 'Open Modules Overview',
      },
      {
        kicker: 'release discipline',
        title: 'Repositories that need explicit maturity signals',
        body: 'Plumego is a better fit when you want release posture and roadmap direction called out separately instead of implied by package existence.',
        fitPoints: [
          'compatibility claims should follow evidence',
          'stable and experimental surfaces should be easy to distinguish',
          'adoption decisions need a published boundary map',
        ],
        startHref: '/releases',
        startLabel: 'Inspect Releases',
      },
    ],
  },
  zh: {
    title: '采用判断',
    description: '用这个页面判断当前服务是否适合 Plumego，以及下一步应该沿哪条阅读路径进入。',
    eyebrow: 'Fit Check',
    introTitle: '用这一页判断 Plumego 是否匹配你眼前的服务。',
    introBody:
      'Architecture 负责解释仓库如何分层；这一页更收敛，只负责帮你判断 Plumego 是否适合当前服务，以及下一步该打开哪条入口路径。',
    guideCards: [
      {
        kicker: 'good fit',
        title: '边界敏感的 Go 服务',
        body: '当团队需要在代码评审里直接看清 route wiring、handler 形态和 ownership 决策时，canonical path 会更有价值。',
      },
      {
        kicker: 'default path',
        title: '需要统一起步形态的内部服务',
        body: 'reference/standard-service 能给新服务提供共同的 bootstrap 与 routing 模型，而不是每个仓库自己长一套目录变体。',
      },
    ],
    decisionTitle: '三个快速判断问题',
    decisionBody:
      '如果你的团队更希望框架藏进约定里，Plumego 可能会显得过于显式。相反，如果你想要可见的路由决策、统一服务形态与显式成熟度信号，它会更有吸引力。',
    cases: [
      {
        kicker: 'internal api',
        title: '需要显式评审路径的内部 HTTP 服务',
        body: '当你希望在代码评审中直接看清 route 注册、中间件顺序和依赖 wiring 时，Plumego 很合适。',
        fitPoints: [
          'reviewer 很关注服务边界',
          'bootstrap 与 handler ownership 需要保持直观',
          '团队本来就按 net/http 心智在思考',
        ],
        startHref: '/zh/docs/getting-started',
        startLabel: '从开始使用进入',
      },
      {
        kicker: 'platform',
        title: '需要统一骨架的共享平台服务',
        body: '当多个团队需要使用同一套服务骨架，而且你希望新服务之间的结构漂移尽量小，Plumego 会更顺手。',
        fitPoints: [
          '新服务应该从同一参考布局起步',
          '应用本地 wiring 需要继续显式',
          'review 只想接受一小组明确模式',
        ],
        startHref: '/zh/docs/reference-app',
        startLabel: '阅读参考应用',
      },
      {
        kicker: 'extensions',
        title: '应当留在内核之外的能力工作',
        body: '当仓库里存在可选 transport 或产品能力层时，Plumego 的 stable-root 与 x/* 分层能给这类工作更清晰的归属。',
        fitPoints: [
          '可选能力不应该重定义默认路径',
          'tenant、AI、gateway、data 逻辑应留在稳定根之外',
          'ownership 比便捷包装更重要',
        ],
        startHref: '/zh/docs/modules/overview',
        startLabel: '查看模块总览',
      },
      {
        kicker: 'release discipline',
        title: '需要显式成熟度信号的仓库',
        body: '如果你希望把 release posture 和 roadmap ambition 明确分开，而不是看到包存在就默认稳定，Plumego 更匹配。',
        fitPoints: [
          '兼容性声明需要证据支撑',
          '稳定与实验表面应该容易区分',
          '采用决策需要一张公开的边界图',
        ],
        startHref: '/zh/releases',
        startLabel: '查看发布页',
      },
    ],
  },
} as const;

export const EXAMPLES_COPY = {
  en: {
    title: 'Examples',
    description: 'The example paths Plumego can stand behind today, from the canonical runnable service to guided boundary-reading examples.',
    eyebrow: 'Practical Paths',
    introTitle: 'One canonical service, then five focused recipe tracks.',
    introBody:
      'The examples path has one runnable baseline and five guided recipe tracks. Start from reference/standard-service to see the default service shape, then pick a recipe to answer one concrete question — request flow, module fit, maturity check, or repository shape — before branching further.',
    guideCards: [
      {
        kicker: 'runnable',
        title: 'One canonical runnable example',
        body: 'reference/standard-service is the example Plumego expects users to run first and compare against when creating a new service.',
      },
      {
        kicker: 'guided',
        title: 'Guided examples explain the repository around it',
        body: 'Request flow, module boundaries, and release posture are examples of how to navigate the repository rather than extra mini-apps.',
      },
      {
        kicker: 'honest scope',
        title: 'The site only promotes examples the repo can actually support',
        body: 'This page is intentionally narrower than a framework showcase. It favors examples that reinforce the canonical path instead of distracting from it.',
      },
    ],
    runnableTitle: 'Runnable now',
    runnableBody:
      'If you want one example to clone, run, and compare against your own service, use the reference app first.',
    runnableExamples: [
      {
        kicker: 'canonical',
        title: 'reference/standard-service',
        body: 'The default app layout for bootstrap, route wiring, and transport-only handlers.',
        details: [
          'best starting point for new services',
          'shows the default directory shape',
          'maps directly to docs and validation guidance',
        ],
        href: '/docs/reference-app',
        label: 'Open the reference app guide',
      },
    ],
    guidedTitle: 'Guided examples',
    guidedBody:
      "These examples are not separate runnable apps. They are release-grade walkthroughs that teach the repository's intended reading path and ownership rules.",
    guidedExamples: [
      {
        kicker: 'request path',
        title: 'Read one request from route to write path',
        body: 'Use the request-flow page when you need to trace where HTTP work begins and where transport responsibility should end.',
        href: '/docs/concepts/request-flow',
        label: 'Open request flow',
      },
      {
        kicker: 'module fit',
        title: 'Classify a change before opening packages',
        body: 'Use modules overview, stable roots, and x/* family pages to decide whether work belongs in the stable surface or an extension family.',
        href: '/docs/modules/overview',
        label: 'Open modules overview',
      },
      {
        kicker: 'maturity',
        title: 'Check whether an area is stable enough to adopt',
        body: 'Use release posture and the public releases page to keep compatibility assumptions explicit.',
        href: '/docs/release-posture',
        label: 'Open release posture',
      },
    ],
    advancedTitle: 'Advanced reference apps',
    advancedBody:
      'These reference apps demonstrate capability-specific service shapes. Each one extends reference/standard-service by adding one x/* family. Read the canonical reference app first, then open one of these when you are ready to evaluate a specific capability.',
    advancedExamples: [
      {
        kicker: 'ai service',
        title: 'reference/with-ai',
        body: 'Multi-provider AI with streaming responses and tool routing. Adds x/ai to the stable HTTP kernel without altering the canonical route shape.',
        maturity: 'x/ai — Experimental',
        details: [
          'provider abstraction separate from HTTP transport',
          'streaming responses through standard net/http handlers',
          'stable roots stay compatible as AI API shapes evolve',
        ],
        href: '/docs/modules/x-ai',
        label: 'Open x/ai primer',
        secondaryHref: '/releases',
        secondaryLabel: 'Check maturity posture',
      },
      {
        kicker: 'multi-tenant saas',
        title: 'reference/with-tenant',
        body: 'Per-tenant routing, quota enforcement, and JWT-backed policy evaluation. Adds x/tenant outside the HTTP kernel so stable roots stay compatible when tenant logic evolves.',
        maturity: 'x/tenant — Experimental',
        details: [
          'tenant identity from JWT or header, evaluated at the transport layer',
          'per-tenant quota and policy separate from route wiring',
          'stable roots carry the HTTP baseline; x/tenant carries the tenant layer',
        ],
        href: '/docs/modules/x-tenant',
        label: 'Open x/tenant primer',
        secondaryHref: '/releases',
        secondaryLabel: 'Check maturity posture',
      },
    ],
    referenceMatrixTitle: 'All capability reference apps',
    referenceMatrixBody:
      'Each app adds one x/* family to the standard-service baseline. Read the canonical reference app first, then open the one that matches your capability.',
    referenceMatrix: [
      { name: 'reference/with-gateway', kicker: 'x/gateway', description: 'Edge proxy, load balancing, and route rewriting', href: '/docs/modules/x-gateway', maturity: 'Experimental' },
      { name: 'reference/with-messaging', kicker: 'x/messaging', description: 'Async message publishing and subscription wiring', href: '/docs/modules/x-messaging', maturity: 'Experimental' },
      { name: 'reference/with-websocket', kicker: 'x/websocket', description: 'WebSocket real-time transport', href: '/docs/modules/x-websocket', maturity: 'Experimental' },
      { name: 'reference/with-webhook', kicker: 'x/webhook', description: 'Webhook receiver with signature verification', href: '/docs/modules/x-webhook', maturity: 'Experimental' },
      { name: 'reference/with-rest', kicker: 'x/rest', description: 'CRUD resource controllers and REST conventions', href: '/docs/modules/x-rest', maturity: 'Experimental' },
      { name: 'reference/with-ops', kicker: 'x/ops', description: 'Protected admin and operations surfaces', href: '/docs/modules/x-ops', maturity: 'Experimental' },
      { name: 'reference/production-service', kicker: 'stable roots', description: 'Production-hardened variant with full lifecycle, TLS, and tests', href: '/docs/reference-app', maturity: 'Supported reference' },
    ],
    workerfleetTitle: 'Production-scale reference: reference/workerfleet',
    workerfleetBody:
      'reference/workerfleet is a full-depth production reference app — distributed worker fleet management with domain models, MongoDB-backed stores, Kubernetes pod discovery, Prometheus metrics with custom collectors, alert engine with deduplication and threshold evaluation, and Feishu/webhook notifications. Use it to evaluate Plumego\'s capability depth beyond tutorial services.',
    workerfleetDetails: [
      'domain-driven design: task, worker, pod, alert, and event models',
      'MongoDB stores with index management and integration tests',
      'Kubernetes watch-based pod sync and discovery',
      'Prometheus metrics with custom collectors and Grafana dashboards',
      'alert engine with deduplication, threshold rules, and notifiers (Feishu, webhook)',
    ],
    workerfleetMaturity: 'Production reference — full-depth example',
    workerfleetLabel: 'Read workerfleet README',
  },
  zh: {
    title: '示例',
    description: 'Plumego 当前真正能站得住的示例路径：从 canonical 可运行服务，到围绕它展开的引导式示例。',
    eyebrow: 'Practical Paths',
    introTitle: '一个 canonical 服务，加五条聚焦的 recipe 路径。',
    introBody:
      '示例路径包含一个可运行的基线，以及五条引导式 recipe。先从 reference/standard-service 看默认服务形态，再按需选一条 recipe 回答具体问题——请求路径、模块归属、成熟度确认或仓库分层——然后再向更深的方向展开。',
    guideCards: [
      {
        kicker: 'runnable',
        title: '只有一个 canonical 可运行示例',
        body: 'reference/standard-service 是 Plumego 期望用户最先运行、也最先拿来对照新服务的示例。',
      },
      {
        kicker: 'guided',
        title: '引导式示例解释它周围的仓库结构',
        body: '请求路径、模块边界和发布姿态这些页面，本质上是教你如何读仓库，而不是再做一堆迷你应用。',
      },
      {
        kicker: 'honest scope',
        title: '站点只推广仓库真正能支持的示例',
        body: '这页会比传统框架 showcase 更窄，因为它优先强化 canonical path，而不是分散注意力。',
      },
    ],
    runnableTitle: '现在就能跑的示例',
    runnableBody:
      '如果你只准备跑一个示例，并把它当成自己服务的对照物，就先用 reference app。',
    runnableExamples: [
      {
        kicker: 'canonical',
        title: 'reference/standard-service',
        body: '默认应用布局，覆盖 bootstrap、route wiring 和 transport-only handler。',
        details: [
          '最适合作为新服务起点',
          '展示默认目录结构',
          '可以直接映射到文档与验证路径',
        ],
        href: '/zh/docs/reference-app',
        label: '打开参考应用说明',
      },
    ],
    guidedTitle: '引导式示例',
    guidedBody:
      '这些示例不是额外的可运行应用，而是发布级 walkthrough，用来教你理解这套仓库的默认阅读路径和 ownership 规则。',
    guidedExamples: [
      {
        kicker: 'request path',
        title: '从 route 到写回路径读一遍请求',
        body: '当你需要追踪 HTTP 工作从哪里开始、transport 责任在哪里结束时，就先看 request-flow 页面。',
        href: '/zh/docs/concepts/request-flow',
        label: '打开 Request Flow',
      },
      {
        kicker: 'module fit',
        title: '在翻包前先判断改动归属',
        body: '用 modules overview、stable roots 和 x/* family 页面先决定工作应该落在稳定表面还是扩展家族。',
        href: '/zh/docs/modules/overview',
        label: '打开模块总览',
      },
      {
        kicker: 'maturity',
        title: '先判断成熟度，再决定是否采用',
        body: '用 release posture 和发布页保持兼容性假设显式，而不是默认它已经稳定。',
        href: '/zh/docs/release-posture',
        label: '打开发布姿态',
      },
    ],
    advancedTitle: '进阶参考应用',
    advancedBody:
      '这些参考应用演示了特定能力的服务形态，每一个都在 reference/standard-service 的基础上加入一个 x/* 家族。先读 canonical 参考应用，再根据你需要评估的具体能力选择进入哪一个。',
    advancedExamples: [
      {
        kicker: 'ai service',
        title: 'reference/with-ai',
        body: '带多 provider 抽象、streaming 响应和 tool 路由的 AI 服务。在稳定 HTTP 内核基础上加入 x/ai，不改变 canonical route 结构。',
        maturity: 'x/ai — 实验性',
        details: [
          'provider 抽象与 HTTP transport 分离',
          'streaming 响应通过标准 net/http handler 处理',
          'AI API 形态演进时，稳定根保持兼容',
        ],
        href: '/zh/docs/modules/x-ai',
        label: '打开 x/ai 模块手册',
        secondaryHref: '/zh/releases',
        secondaryLabel: '查看成熟度姿态',
      },
      {
        kicker: 'multi-tenant saas',
        title: 'reference/with-tenant',
        body: '带 per-tenant 路由、配额执行和 JWT 策略评估的多租户服务。x/tenant 在 HTTP 内核之外承接租户层，稳定根不受租户逻辑演进影响。',
        maturity: 'x/tenant — 实验性',
        details: [
          '租户身份来自 JWT 或 header，在传输层评估',
          'per-tenant 配额与策略与 route wiring 分开',
          '稳定根承担 HTTP 基线，x/tenant 承担租户层',
        ],
        href: '/zh/docs/modules/x-tenant',
        label: '打开 x/tenant 模块手册',
        secondaryHref: '/zh/releases',
        secondaryLabel: '查看成熟度姿态',
      },
    ],
    referenceMatrixTitle: '所有能力参考应用',
    referenceMatrixBody:
      '每个参考应用都在 standard-service 的基础上加入一个 x/* 家族。先读 canonical 参考应用，再按你需要评估的能力选择进入。',
    referenceMatrix: [
      { name: 'reference/with-gateway', kicker: 'x/gateway', description: '边缘代理、负载均衡与路由重写', href: '/zh/docs/modules/x-gateway', maturity: '实验性' },
      { name: 'reference/with-messaging', kicker: 'x/messaging', description: '异步消息发布与订阅接线', href: '/zh/docs/modules/x-messaging', maturity: '实验性' },
      { name: 'reference/with-websocket', kicker: 'x/websocket', description: 'WebSocket 实时传输', href: '/zh/docs/modules/x-websocket', maturity: '实验性' },
      { name: 'reference/with-webhook', kicker: 'x/webhook', description: '带签名校验的 Webhook 接收器', href: '/zh/docs/modules/x-webhook', maturity: '实验性' },
      { name: 'reference/with-rest', kicker: 'x/rest', description: 'CRUD 资源控制器与 REST 规范', href: '/zh/docs/modules/x-rest', maturity: '实验性' },
      { name: 'reference/with-ops', kicker: 'x/ops', description: '受保护的管理与运维表面', href: '/zh/docs/modules/x-ops', maturity: '实验性' },
      { name: 'reference/production-service', kicker: 'stable roots', description: '带完整生命周期、TLS 和测试的生产级加固变体', href: '/zh/docs/reference-app', maturity: '受支持参考' },
    ],
    workerfleetTitle: '生产规模参考：reference/workerfleet',
    workerfleetBody:
      'reference/workerfleet 是一个完整深度的生产参考应用——分布式 worker 机队管理，包含领域模型、MongoDB 存储、Kubernetes Pod 发现、Prometheus 指标、告警引擎（带去重与阈值评估）以及飞书/webhook 通知。适合用来评估 Plumego 在超出教程级别时的能力深度。',
    workerfleetDetails: [
      '领域驱动设计：task、worker、pod、alert 和 event 模型',
      'MongoDB 存储，含索引管理和集成测试',
      'Kubernetes watch-based pod 同步与发现',
      'Prometheus 指标，含自定义 collector 和 Grafana 看板',
      '告警引擎，含去重、阈值规则和通知器（飞书、webhook）',
    ],
    workerfleetMaturity: '生产参考 — 完整深度示例',
    workerfleetLabel: '阅读 workerfleet README',
  },
} as const;

export const NOT_FOUND_COPY = {
  en: {
    title: 'Page Not Found',
    description: 'The page you requested does not exist or has moved. Start again from docs, the reference app, or the release pages.',
    eyebrow: '404',
    body:
      'The URL may have moved while the website structure was being tightened. Use one of the primary entry points below instead of guessing deeper paths.',
    actions: [
      { label: 'Go to Docs', href: '/docs' },
      { label: 'Open Getting Started', href: '/docs/getting-started' },
      { label: 'View Releases', href: '/releases' },
      { label: '切换到中文', href: '/zh' },
    ],
  },
  zh: {
    title: '页面不存在',
    description: '你访问的页面不存在，或者路径已经调整。请从文档、参考应用或发布页重新进入。',
    eyebrow: '404',
    body:
      '站点结构在收敛过程中，一部分路径可能已经调整。不要继续猜更深的 URL，直接从下面这些主入口重新进入。',
    actions: [
      { label: '进入文档', href: '/zh/docs' },
      { label: '打开开始使用', href: '/zh/docs/getting-started' },
      { label: '查看发布页', href: '/zh/releases' },
      { label: 'Switch to English', href: '/' },
    ],
  },
} as const;
