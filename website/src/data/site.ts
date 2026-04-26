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
    { label: 'Why Plumego', href: '/why-plumego' },
    { label: 'Architecture', href: '/architecture' },
    { label: 'Examples', href: '/examples' },
    { label: 'Releases', href: '/releases' },
    { label: 'GitHub', href: SITE.githubUrl },
  ],
  zh: [
    { label: '文档', href: '/zh/docs' },
    { label: '为什么选择', href: '/zh/why-plumego' },
    { label: '架构', href: '/zh/architecture' },
    { label: '示例', href: '/zh/examples' },
    { label: '发布', href: '/zh/releases' },
    { label: 'GitHub', href: SITE.githubUrl },
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
        { label: 'GitHub', href: SITE.githubUrl },
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
        { label: 'GitHub', href: SITE.githubUrl },
      ],
    },
  ],
};

export const HOME_COPY = {
  en: {
    eyebrow: 'Plumego',
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
      label: 'Ready now',
      title: `${RELEASE_FACTS.currentVersion} — stable path ready to adopt`,
      body: 'Stable roots are production-ready. Extension families carry explicit maturity labels so you always know which surfaces hold compatibility promises.',
    },
    heroVisual: {
      stableLabel: 'one stable model',
      stableTitle: 'explicit request path',
      stableBody: 'Routes, middleware, handlers, and application wiring remain visible in code instead of disappearing into framework-owned registration.',
      extensionLabel: 'one expansion rule',
      extensionTitle: 'capabilities branch outward',
      extensionBody: 'Stable roots stay narrow while optional capability work starts in x/* families instead of stretching the kernel.',
      footerLabel: 'default path',
      footerBody: 'docs → reference app → request flow → release posture → module fit',
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
        body: 'Check which surfaces are stable, supported reference, or still experimental before widening adoption scope.',
      },
      {
        label: '05',
        title: 'Only then classify module ownership',
        body: 'Decide whether the next change belongs in stable roots, app-local wiring, or x/* families.',
      },
    ],
    valueTitle: 'Why Plumego exists.',
    valueLead:
      'Plumego does not try to out-abstract the Go HTTP model. It exists to keep the important parts visible as the repository and the team get larger, and to make adoption decisions easier to defend in review.',
    values: [
      {
        kicker: 'explicitness',
        title: 'Keep handlers, routes, and wiring visible',
        body: 'The framework value is not hidden registration. You can still inspect how a service starts, how routes are wired, and where dependencies enter.',
      },
      {
        kicker: 'boundaries',
        title: 'Keep the stable surface intentionally narrow',
        body: 'Stable roots carry long-lived responsibilities. Optional or fast-moving capability work starts in x/* families instead of leaking into the kernel.',
      },
      {
        kicker: 'release posture',
        title: 'Publish what is safe to adopt today',
        body: 'Compatibility is explained in public instead of being inferred from package existence alone, so teams can distinguish baseline from experimentation.',
      },
      {
        kicker: 'repo control',
        title: 'Make repository intent easier to classify',
        body: 'Docs, specs, generated facts, and task cards make change ownership easier to reason about for both humans and coding agents.',
      },
    ],
    valueFootnote:
      'The point is not more framework. The point is a toolkit that still reads clearly after the codebase stops being small and when multiple teams need one defensible default path.',
    adoptionTitle: 'Choose the next page by question, not by package.',
    adoptionBody:
      'Architecture explains how the repository is shaped. Examples show the canonical runnable path. Releases tells you what can be adopted now versus what should stay deliberate. Pick the page that answers your next decision.',
    adoptionCards: [
      {
        kicker: 'fit',
        title: 'Why Plumego',
        body: 'Use the dedicated adoption page when the question is not package topology but whether Plumego actually matches the service, team posture, and repository shape in front of you.',
        href: '/why-plumego',
        label: 'Evaluate fit',
      },
      {
        kicker: 'topology',
        title: 'Architecture',
        body: 'See how stable roots, extension families, and the canonical request path fit together before deciding where a change belongs.',
        href: '/architecture',
        label: 'Explore architecture',
      },
      {
        kicker: 'paths',
        title: 'Examples',
        body: 'Start from the one canonical runnable example, then move into guided recipes that explain request flow, boundaries, and adoption posture.',
        href: '/examples',
        label: 'Explore examples',
      },
      {
        kicker: 'adoption',
        title: 'Releases',
        body: 'Use release posture when the question is not “does this package exist?” but “which surfaces are mature enough to trust today?”',
        href: '/releases',
        label: 'Inspect releases',
      },
    ],
    stableRoots: MODULE_FACTS.stableRoots,
    extensions: MODULE_FACTS.primaryExtensionFamilies,
    canonicalTitle: 'Start from one canonical request path.',
    canonicalBody:
      'Begin with reference/standard-service. The point is not just one service skeleton—it is one readable path from bootstrap to route registration, release posture, and module classification before deeper boundary work begins.',
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
      {
        label: 'release posture',
        title: 'Verify maturity before widening scope',
        body: 'Use the release page to separate stable roots, supported reference surfaces, and experimental families.',
      },
      {
        label: 'boundaries',
        title: 'Only then compare module ownership',
        body: 'Use architecture and modules overview to decide whether the next change belongs in stable roots or x/* families.',
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
        title: 'x/* families protect optional capability work',
        caption: 'Product or protocol-specific work moves outward instead of stretching the kernel, and stays easier to evaluate deliberately.',
      },
      footer: 'Read the path first; use the module map when the next question is ownership or release posture.',
    },
    finalTitle: 'Read the docs, run the reference app, then expand only when the boundary is clear.',
    finalBody:
      'Plumego works best when you start from the default path, check release posture early, and branch into deeper boundaries only after the owning capability is obvious.',
    finalPrimary: { label: 'Read Docs', href: '/docs' },
    finalSecondary: { label: 'Open GitHub', href: SITE.githubUrl },
  },
  zh: {
    eyebrow: 'Plumego',
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
      label: '现在就可以用',
      title: `${RELEASE_FACTS.currentVersion} —— 稳定路径已经可以采用`,
      body: '稳定根已经生产可用。扩展家族明确标注成熟度，让你随时知道哪些表面有兼容性承诺，哪些仍在快速演进。',
    },
    heroVisual: {
      stableLabel: '一套稳定模型',
      stableTitle: '显式请求路径',
      stableBody: 'Routes、middleware、handlers 和应用 wiring 都继续留在代码里，而不是消失在框架注册机制后面。',
      extensionLabel: '一条扩展规则',
      extensionTitle: '能力向外分叉',
      extensionBody: '稳定根保持收敛，可选能力从 x/* 家族开始，而不是继续拉长内核。',
      footerLabel: '默认路径',
      footerBody: 'docs → reference app → request flow → 发布姿态 → 模块归属',
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
        body: '在扩大采用范围之前，先确认哪些表面属于稳定根，哪些仍然只是支持参考或实验性能力。',
      },
      {
        label: '05',
        title: '最后判断模块归属',
        body: '再决定这项工作属于稳定根、应用本地 wiring，还是 x/* 能力家族。',
      },
    ],
    valueTitle: 'Plumego 为什么存在。',
    valueLead:
      'Plumego 不是要重新发明 Go 的 HTTP 模型。它的目标，是在仓库和团队变大以后，继续把关键结构保持可见，并让采用决策在评审中更容易被解释。',
    values: [
      {
        kicker: '显式性',
        title: '让 handler、route 和 wiring 保持可见',
        body: '价值不在于隐藏注册过程，而在于你仍然能直接看清服务如何启动、路由如何连接、依赖从哪里进入。',
      },
      {
        kicker: '边界',
        title: '让稳定表面保持收敛',
        body: '稳定根承接长期职责，可选或快速演进能力从 x/* 家族进入，而不是一路渗进内核。',
      },
      {
        kicker: '发布姿态',
        title: '把今天真正能安全采用的范围公开出来',
        body: '兼容性不是靠“仓库里有这个包”来暗示，而是由页面明确解释，让团队能区分基线与实验区。',
      },
      {
        kicker: '仓库控制面',
        title: '让仓库意图更容易判断',
        body: 'docs、specs、同步事实与 task card 分层明确，让人和 agent 都更容易判断改动该落在哪。',
      },
    ],
    valueFootnote:
      '重点不是更多框架抽象，而是让服务工具包在仓库不再小、团队开始协作分工之后，依然有一条可辩护的默认路径。',
    adoptionTitle: '根据问题选下一页，而不是先选包。',
    adoptionBody:
      'Architecture 负责解释仓库如何分层；Examples 负责展示 canonical 可运行路径；Releases 负责解释哪些区域现在就适合采用。先选真正能回答你下一个决策的问题的页面。',
    adoptionCards: [
      {
        kicker: 'fit',
        title: '为什么选择 Plumego',
        body: '如果你关心的不再是包拓扑，而是 Plumego 是否真的匹配眼前这个服务、团队姿态与仓库目标，就先走专门的 adoption 判断页。',
        href: '/zh/why-plumego',
        label: '判断是否适合',
      },
      {
        kicker: 'topology',
        title: '架构',
        body: '先看稳定根、扩展家族与 canonical 请求路径如何组合，再判断某个改动究竟该落在哪一层。',
        href: '/zh/architecture',
        label: '查看架构',
      },
      {
        kicker: 'paths',
        title: '示例',
        body: '从唯一的 canonical 可运行示例起步，再进入解释请求路径、边界和采用姿态的引导式示例。',
        href: '/zh/examples',
        label: '查看示例',
      },
      {
        kicker: 'adoption',
        title: '发布',
        body: '当问题不再是“仓库里有没有这个包”，而是“今天哪些表面真正值得信任”时，就去看发布姿态。',
        href: '/zh/releases',
        label: '查看发布页',
      },
    ],
    stableRoots: MODULE_FACTS.stableRoots,
    extensions: MODULE_FACTS.primaryExtensionFamilies,
    canonicalTitle: '先从一条 canonical 请求路径起步。',
    canonicalBody:
      '先看 reference/standard-service。重点不只是服务骨架本身，而是一条从 bootstrap 到 route 注册、再到发布姿态与模块归属都保持可读的默认路径。',
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
        body: '在继续深入 handler 或模块之前，先看清 route 注册。',
      },
      {
        label: 'release posture',
        title: '确认成熟度再扩大范围',
        body: '用发布页分清哪些是稳定根，哪些属于支持参考表面，哪些仍然是实验性家族。',
      },
      {
        label: 'boundaries',
        title: '最后再判断模块归属',
        body: '用 architecture 与 modules overview 判断下一项工作属于稳定根还是 x/* 家族。',
      },
    ],
    mapTitle: '只有 canonical path 清楚以后，再看 module map。',
    mapLead:
      'module map 不是第二个起点，而是在默认请求路径与发布姿态已经读清之后，帮助你继续判断更深工作归属的下一层视图。',
    mapPanels: {
      stable: {
        title: '稳定根负责保护默认路径',
        caption: '这些模块更收敛、兼容性负担更重，也承担公开基线的角色。',
      },
      extension: {
        title: 'x/* 家族负责承接可选能力工作',
        caption: '产品能力或协议工作向外分叉，而不是继续把内核拉长，也更适合被审慎地单独评估。',
      },
      footer: '先读路径，再用 module map 判断 ownership 与发布姿态。',
    },
    finalTitle: '先读文档，跑通 reference app，再在边界清楚时向外扩展。',
    finalBody: 'Plumego 最适合从默认路径进入，并尽早检查发布姿态；只有在 capability ownership 已经明确之后，再进入更深的模块边界。',
    finalPrimary: { label: '阅读文档', href: '/zh/docs' },
    finalSecondary: { label: '打开 GitHub', href: SITE.githubUrl },
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
      {
        kicker: 'repo control',
        title: 'Repositories that want agents to classify work reliably',
        body: 'Plumego’s docs/specs/tasks split helps both humans and agents decide where a change belongs before the diff starts to sprawl.',
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
        body: 'When your repository has optional transport or product capability layers, Plumego’s stable-root versus x/* split gives that work a clearer home.',
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
      {
        kicker: 'repo control',
        title: '希望 agent 也能稳定分类工作的仓库',
        body: 'docs/specs/tasks 的分层让人和 agent 都能在 diff 失控之前先判断改动归属。',
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
    introTitle: 'Start from one runnable example, then branch into guided examples.',
    introBody:
      'Plumego is intentionally conservative about what counts as a canonical example. Today the primary runnable example is reference/standard-service; the surrounding pages show how to read that example, how to classify ownership, and how to expand from it without treating the repo as a feature catalog.',
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
      'These examples are not separate runnable apps. They are release-grade walkthroughs that teach the repository’s intended reading path and ownership rules.',
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
  },
  zh: {
    title: '示例',
    description: 'Plumego 当前真正能站得住的示例路径：从 canonical 可运行服务，到围绕它展开的引导式示例。',
    eyebrow: 'Practical Paths',
    introTitle: '先从一个可运行示例开始，再进入引导式示例。',
    introBody:
      'Plumego 对“什么算 canonical 示例”非常克制。当前最主要的可运行示例是 reference/standard-service；围绕它的页面则负责说明如何读这个示例、如何判断 ownership，以及如何从它扩展，而不是把仓库当成功能目录。',
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
