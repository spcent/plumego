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
        './src/styles/home.css',
      ],
      components: {
        Head: './src/components/starlight/Head.astro',
        Header: './src/components/starlight/Header.astro',
        ThemeSelect: './src/components/shared/ThemeSelect.astro',
        LanguageSelect: './src/components/shared/LanguageSelect.astro',
      },
      sidebar: [
        {
          label: 'Documentation',
          autogenerate: {
            directory: 'docs',
          },
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
