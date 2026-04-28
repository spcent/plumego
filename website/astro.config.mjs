import { defineConfig } from 'astro/config';
import mdx from '@astrojs/mdx';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://plumego.dev',
  output: 'static',
  trailingSlash: 'never',
  integrations: [
    starlight({
      title: 'Plumego',
      description: 'A stdlib-first Go web toolkit with clear module boundaries.',
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
        './src/styles/theme.css',
      ],
      components: {
        Head: './src/components/starlight/Head.astro',
        Header: './src/components/starlight/Header.astro',
        ThemeSelect: './src/components/shared/ThemeSelect.astro',
        LanguageSelect: './src/components/shared/LanguageSelect.astro',
      },
      sidebar: [
        {
          label: 'Prologue',
          translations: { 'zh-CN': '序言' },
          items: [
            { label: 'Introduction', slug: 'docs', translations: { 'zh-CN': '介绍' } },
            { label: 'Release Posture', slug: 'docs/release-posture', translations: { 'zh-CN': '发布策略' } },
            { label: 'Stable Roots', slug: 'docs/stable-roots', translations: { 'zh-CN': '稳定根' } },
            { label: 'X Family', slug: 'docs/x-family', translations: { 'zh-CN': 'X 家族' } },
          ],
        },
        {
          label: 'Getting Started',
          translations: { 'zh-CN': '快速上手' },
          items: [
            { label: 'Getting Started', slug: 'docs/getting-started', translations: { 'zh-CN': '开始使用' } },
            { label: 'Reference App', slug: 'docs/reference-app', translations: { 'zh-CN': '参考应用' } },
            { label: 'FAQ', slug: 'docs/faq', translations: { 'zh-CN': '常见问题' } },
          ],
        },
        {
          label: 'Core Concepts',
          translations: { 'zh-CN': '核心概念' },
          items: [
            { label: 'Request Flow', slug: 'docs/concepts/request-flow', translations: { 'zh-CN': '请求流程' } },
            { label: 'Middleware Model', slug: 'docs/concepts/middleware-model', translations: { 'zh-CN': '中间件模型' } },
            { label: 'Error Model', slug: 'docs/concepts/error-model', translations: { 'zh-CN': '错误模型' } },
            { label: 'Configuration', slug: 'docs/concepts/configuration-model', translations: { 'zh-CN': '配置模型' } },
            { label: 'Extension Maturity', slug: 'docs/concepts/extension-maturity', translations: { 'zh-CN': '扩展成熟度' } },
            { label: 'Repo Control Plane', slug: 'docs/concepts/repo-control-plane', translations: { 'zh-CN': '仓库控制面' } },
          ],
        },
        {
          label: 'Stable Modules',
          translations: { 'zh-CN': '稳定模块' },
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
        {
          label: 'Extensions (x/*)',
          translations: { 'zh-CN': '扩展模块 (x/*)' },
          collapsed: true,
          items: [
            { label: 'x/rest', slug: 'docs/modules/x-rest' },
            { label: 'x/cache', slug: 'docs/modules/x-cache' },
            { label: 'x/data', slug: 'docs/modules/x-data' },
            { label: 'x/messaging', slug: 'docs/modules/x-messaging' },
            { label: 'x/messaging (sub)', slug: 'docs/modules/x-messaging-subordinates' },
            { label: 'x/websocket', slug: 'docs/modules/x-websocket' },
            { label: 'x/scheduler', slug: 'docs/modules/x-scheduler' },
            { label: 'x/resilience', slug: 'docs/modules/x-resilience' },
            { label: 'x/observability', slug: 'docs/modules/x-observability' },
            { label: 'x/gateway', slug: 'docs/modules/x-gateway' },
            { label: 'x/ai', slug: 'docs/modules/x-ai' },
            { label: 'x/fileapi', slug: 'docs/modules/x-fileapi' },
            { label: 'x/frontend', slug: 'docs/modules/x-frontend' },
            { label: 'x/tenant', slug: 'docs/modules/x-tenant' },
            { label: 'x/webhook', slug: 'docs/modules/x-webhook' },
            { label: 'x/discovery', slug: 'docs/modules/x-discovery' },
            { label: 'x/ops', slug: 'docs/modules/x-ops' },
            { label: 'x/devtools', slug: 'docs/modules/x-devtools' },
          ],
        },
        {
          label: 'Guides',
          translations: { 'zh-CN': '实践指南' },
          collapsed: true,
          items: [
            { label: 'Build a REST Resource', slug: 'docs/guides/build-rest-resource', translations: { 'zh-CN': '构建 REST 资源' } },
            { label: 'Add JWT Auth', slug: 'docs/guides/add-jwt-auth', translations: { 'zh-CN': '添加 JWT 认证' } },
            { label: 'Custom Middleware', slug: 'docs/guides/custom-middleware', translations: { 'zh-CN': '自定义中间件' } },
            { label: 'Handle Errors', slug: 'docs/guides/handle-errors', translations: { 'zh-CN': '错误处理' } },
            { label: 'Health & Readiness', slug: 'docs/guides/health-and-readiness', translations: { 'zh-CN': '健康检查' } },
            { label: 'Structured Logging', slug: 'docs/guides/structured-logging', translations: { 'zh-CN': '结构化日志' } },
            { label: 'Connect Database', slug: 'docs/guides/connect-database', translations: { 'zh-CN': '连接数据库' } },
            { label: 'Testing Handlers', slug: 'docs/guides/testing-handlers', translations: { 'zh-CN': '测试 Handler' } },
            { label: 'File Uploads', slug: 'docs/guides/file-uploads', translations: { 'zh-CN': '文件上传' } },
            { label: 'WebSocket', slug: 'docs/guides/websocket', translations: { 'zh-CN': 'WebSocket' } },
            { label: 'Graceful Shutdown', slug: 'docs/guides/graceful-shutdown', translations: { 'zh-CN': '优雅停机' } },
            { label: 'Deploy with Docker', slug: 'docs/guides/deploy-with-docker', translations: { 'zh-CN': 'Docker 部署' } },
            { label: 'Observability', slug: 'docs/guides/observability-integration', translations: { 'zh-CN': '可观测性集成' } },
            { label: 'Integrate AI', slug: 'docs/guides/integrate-ai', translations: { 'zh-CN': '集成 AI' } },
            { label: 'Multi-tenancy', slug: 'docs/guides/multi-tenancy', translations: { 'zh-CN': '多租户' } },
            { label: 'Migration & Upgrades', slug: 'docs/guides/migration-and-upgrades', translations: { 'zh-CN': '迁移与升级' } },
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
