import { defineConfig } from 'astro/config';
import mdx from '@astrojs/mdx';
import sitemap from '@astrojs/sitemap';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://plumego.birdor.com',
  output: 'static',
  trailingSlash: 'never',
  integrations: [
    sitemap({
      filter: (page) => !page.includes('/404'),
      i18n: {
        defaultLocale: 'en',
        locales: {
          en: 'en',
          zh: 'zh-CN',
        },
      },
    }),
    starlight({
      title: 'Plumego',
      description: 'stdlib-only Go HTTP toolkit — explicit by design, agent-ready by structure. Visible routing, narrow stable roots, machine-readable specs.',
      favicon: '/favicon.svg',
      disable404Route: true,
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/spcent/plumego',
        },
      ],
      locales: {
        root: {
          label: 'English',
          lang: 'en',
        },
        zh: {
          label: '简体中文',
          lang: 'zh-CN',
        },
      },
      customCss: [
        './src/styles/tokens.css',
        './src/styles/starlight-bridge.css',
        './src/styles/global.css',
        './src/styles/prose.css',
        './src/styles/home.css',
      ],
      components: {
        Head: './src/components/starlight/Head.astro',
        Header: './src/components/starlight/Header.astro',
        ThemeSelect: './src/components/shared/ThemeSelect.astro',
        LanguageSelect: './src/components/shared/LanguageSelect.astro',
        Banner: './src/components/starlight/Banner.astro',
      },
      sidebar: [
        {
          label: 'Getting Started',
          translations: { 'zh-CN': '快速上手' },
          items: [
            { label: 'Introduction', slug: 'docs', translations: { 'zh-CN': '介绍' } },
            { label: 'Getting Started', slug: 'docs/getting-started', translations: { 'zh-CN': '开始使用' } },
            { label: 'Reference App', slug: 'docs/reference-app', translations: { 'zh-CN': '参考应用' } },
            { label: 'FAQ', slug: 'docs/faq', translations: { 'zh-CN': '常见问题' } },
            { label: 'Release Posture', slug: 'docs/release-posture', translations: { 'zh-CN': '发布策略' } },
          ],
        },
        {
          label: 'Core Modules',
          translations: { 'zh-CN': '核心模块' },
          items: [
            { label: 'Stable Roots Overview', slug: 'docs/stable-roots', translations: { 'zh-CN': '稳定根总览' } },
            { label: 'Request Flow', slug: 'docs/concepts/request-flow', translations: { 'zh-CN': '请求流程' } },
            { label: 'Middleware Model', slug: 'docs/concepts/middleware-model', translations: { 'zh-CN': '中间件模型' } },
            { label: 'Error Model', slug: 'docs/concepts/error-model', translations: { 'zh-CN': '错误模型' } },
            { label: 'Configuration', slug: 'docs/concepts/configuration-model', translations: { 'zh-CN': '配置模型' } },
            {
              label: 'Modules',
              translations: { 'zh-CN': '稳定模块' },
              collapsed: false,
              items: [
                { label: 'Overview', slug: 'docs/modules/overview', translations: { 'zh-CN': '总览' } },
                { label: 'core', slug: 'docs/modules/core' },
                { label: 'contract', slug: 'docs/modules/contract' },
                { label: 'router', slug: 'docs/modules/router' },
                { label: 'middleware', slug: 'docs/modules/middleware' },
                { label: 'security', slug: 'docs/modules/security' },
                { label: 'health', slug: 'docs/modules/health' },
                { label: 'log', slug: 'docs/modules/log' },
                { label: 'metrics', slug: 'docs/modules/metrics' },
                { label: 'store', slug: 'docs/modules/store' },
              ],
            },
            { label: 'When Not to Use', slug: 'docs/when-not-to-use', translations: { 'zh-CN': '不适用场景' } },
          ],
        },
        {
          label: 'Extensions (x/*)',
          translations: { 'zh-CN': '扩展模块 (x/*)' },
          items: [
            { label: 'X Family Overview', slug: 'docs/x-family', translations: { 'zh-CN': 'X 家族总览' } },
            { label: 'Extension Maturity', slug: 'docs/concepts/extension-maturity', translations: { 'zh-CN': '扩展成熟度' } },
            {
              label: 'Beta',
              translations: { 'zh-CN': 'Beta 级' },
              collapsed: false,
              items: [
                { label: 'x/rest', slug: 'docs/modules/x-rest' },
                { label: 'x/gateway', slug: 'docs/modules/x-gateway' },
                { label: 'x/websocket', slug: 'docs/modules/x-websocket' },
                { label: 'x/observability', slug: 'docs/modules/x-observability' },
                { label: 'x/tenant', slug: 'docs/modules/x-tenant' },
                { label: 'x/frontend', slug: 'docs/modules/x-frontend' },
                { label: 'x/messaging', slug: 'docs/modules/x-messaging' },
                { label: 'x/messaging (primitives)', slug: 'docs/modules/x-messaging-subordinates' },
              ],
            },
            {
              label: 'Experimental',
              translations: { 'zh-CN': '实验性' },
              collapsed: true,
              items: [
                { label: 'x/ai', slug: 'docs/modules/x-ai' },
                { label: 'x/data', slug: 'docs/modules/x-data' },
                { label: 'x/fileapi', slug: 'docs/modules/x-fileapi' },
                { label: 'x/data/cache', slug: 'docs/modules/x-cache' },
                { label: 'x/resilience', slug: 'docs/modules/x-resilience' },
                { label: 'x/messaging/scheduler', slug: 'docs/modules/x-scheduler' },
                { label: 'x/messaging/webhook', slug: 'docs/modules/x-webhook' },
                { label: 'x/observability/ops', slug: 'docs/modules/x-ops' },
                { label: 'x/gateway/discovery', slug: 'docs/modules/x-discovery' },
                { label: 'x/observability/devtools', slug: 'docs/modules/x-devtools' },
                { label: 'x/gateway/ipc', slug: 'docs/modules/x-ipc' },
                { label: 'x/rpc', slug: 'docs/modules/x-rpc' },
                { label: 'x/openapi', slug: 'docs/modules/x-openapi' },
                { label: 'x/validate', slug: 'docs/modules/x-validate' },
              ],
            },
          ],
        },
        {
          label: 'Guides & Reference',
          translations: { 'zh-CN': '指南与参考' },
          items: [
            {
              label: 'Build & Wire',
              translations: { 'zh-CN': '构建与装配' },
              collapsed: false,
              items: [
                { label: 'Build a REST Resource', slug: 'docs/guides/build-rest-resource', translations: { 'zh-CN': '构建 REST 资源' } },
                { label: 'Custom Middleware', slug: 'docs/guides/custom-middleware', translations: { 'zh-CN': '自定义中间件' } },
                { label: 'Add JWT Auth', slug: 'docs/guides/add-jwt-auth', translations: { 'zh-CN': '添加 JWT 认证' } },
                { label: 'Handle Errors', slug: 'docs/guides/handle-errors', translations: { 'zh-CN': '错误处理' } },
                { label: 'Testing Handlers', slug: 'docs/guides/testing-handlers', translations: { 'zh-CN': '测试 Handler' } },
              ],
            },
            {
              label: 'Data & Protocol',
              translations: { 'zh-CN': '数据与协议' },
              collapsed: true,
              items: [
                { label: 'Connect Database', slug: 'docs/guides/connect-database', translations: { 'zh-CN': '连接数据库' } },
                { label: 'File Uploads', slug: 'docs/guides/file-uploads', translations: { 'zh-CN': '文件上传' } },
                { label: 'WebSocket', slug: 'docs/guides/websocket', translations: { 'zh-CN': 'WebSocket' } },
                { label: 'Integrate AI', slug: 'docs/guides/integrate-ai', translations: { 'zh-CN': '集成 AI' } },
                { label: 'Multi-tenancy', slug: 'docs/guides/multi-tenancy', translations: { 'zh-CN': '多租户' } },
              ],
            },
            {
              label: 'Observe & Deploy',
              translations: { 'zh-CN': '可观测与部署' },
              collapsed: true,
              items: [
                { label: 'Health & Readiness', slug: 'docs/guides/health-and-readiness', translations: { 'zh-CN': '健康检查' } },
                { label: 'Structured Logging', slug: 'docs/guides/structured-logging', translations: { 'zh-CN': '结构化日志' } },
                { label: 'Observability', slug: 'docs/guides/observability-integration', translations: { 'zh-CN': '可观测性集成' } },
                { label: 'Graceful Shutdown', slug: 'docs/guides/graceful-shutdown', translations: { 'zh-CN': '优雅停机' } },
                { label: 'Deploy with Docker', slug: 'docs/guides/deploy-with-docker', translations: { 'zh-CN': 'Docker 部署' } },
                { label: 'Dev Server', slug: 'docs/guides/dev-server', translations: { 'zh-CN': '开发服务器' } },
              ],
            },
            {
              label: 'Migrate',
              translations: { 'zh-CN': '迁移' },
              collapsed: true,
              items: [
                { label: 'Adoption Path', slug: 'docs/guides/adoption-path', translations: { 'zh-CN': '采用路径' } },
                { label: 'Migration & Upgrades', slug: 'docs/guides/migration-and-upgrades', translations: { 'zh-CN': '迁移与升级' } },
                { label: 'Migrate from Gin/Echo', slug: 'docs/guides/migrate-from-gin-echo', translations: { 'zh-CN': '从 Gin/Echo 迁移' } },
                { label: 'Migrate from Chi', slug: 'docs/guides/migrate-from-chi', translations: { 'zh-CN': '从 Chi 迁移' } },
              ],
            },
            {
              label: 'Style Guide',
              translations: { 'zh-CN': '编码规范' },
              collapsed: true,
              items: [
                { label: 'Style Guide', slug: 'docs/guides/style-guide', translations: { 'zh-CN': '编码规范' } },
              ],
            },
            {
              label: 'API Reference',
              translations: { 'zh-CN': 'API 参考' },
              collapsed: true,
              items: [
                { label: 'Overview', slug: 'docs/reference', translations: { 'zh-CN': '总览' } },
                { label: 'core API', slug: 'docs/reference/api-core', translations: { 'zh-CN': 'core API' } },
                { label: 'contract API', slug: 'docs/reference/api-contract', translations: { 'zh-CN': 'contract API' } },
                { label: 'router API', slug: 'docs/reference/api-router', translations: { 'zh-CN': 'router API' } },
                { label: 'Error Reference', slug: 'docs/reference/errors', translations: { 'zh-CN': '错误参考' } },
                { label: 'Stability & Deprecation', slug: 'docs/reference/stability', translations: { 'zh-CN': '稳定性与弃用' } },
              ],
            },
            {
              label: 'Advanced',
              translations: { 'zh-CN': '高级主题' },
              collapsed: true,
              items: [
                { label: 'Repo Control Plane', slug: 'docs/concepts/repo-control-plane', translations: { 'zh-CN': '仓库控制面' } },
                { label: 'Agent-first Workflow', slug: 'docs/concepts/agent-first-workflow', translations: { 'zh-CN': 'Agent 工作流' } },
                { label: 'Core Boundaries', slug: 'docs/concepts/core-boundaries', translations: { 'zh-CN': '核心边界' } },
                { label: 'Extension Boundaries', slug: 'docs/concepts/extension-boundaries', translations: { 'zh-CN': '扩展边界' } },
              ],
            },
          ],
        },
      ],
    }),
    mdx(),
  ],
  markdown: {
    shikiConfig: {
      themes: {
        light: 'github-light',
        dark: 'github-dark',
      },
      wrap: true,
    },
  },
});
