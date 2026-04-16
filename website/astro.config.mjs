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
        './src/styles/global.css',
        './src/styles/prose.css',
        './src/styles/theme.css',
      ],
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
