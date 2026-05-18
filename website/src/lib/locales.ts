export const DEFAULT_LOCALE = 'en';

export const LOCALES = {
  en: {
    label: 'English',
    shortLabel: 'EN',
    pathPrefix: '',
    htmlLang: 'en',
    hreflang: 'en',
    ogLocale: 'en_US',
    defaultOgImage: '/brand/og-default.svg',
    labels: {
      primaryNav: 'Primary',
      mobileNav: 'Mobile navigation',
      openMenu: 'Open menu',
      selectLanguage: 'Select language',
    },
  },
  zh: {
    label: '简体中文',
    shortLabel: '中文',
    pathPrefix: '/zh',
    htmlLang: 'zh-CN',
    hreflang: 'zh',
    ogLocale: 'zh_CN',
    defaultOgImage: '/brand/og-zh.svg',
    labels: {
      primaryNav: '主导航',
      mobileNav: '移动端导航',
      openMenu: '打开菜单',
      selectLanguage: '选择语言',
    },
  },
} as const;

export type Locale = keyof typeof LOCALES;

export const TRANSLATION_SOURCE_LOCALE: Locale = DEFAULT_LOCALE;
export const TRANSLATION_LAG_LOCALE: Locale = 'zh';

export const LOCALE_CODES = Object.keys(LOCALES) as Locale[];

export const PREFIXED_LOCALES = LOCALE_CODES
  .filter((locale) => LOCALES[locale].pathPrefix.length > 0)
  .sort((a, b) => LOCALES[b].pathPrefix.length - LOCALES[a].pathPrefix.length);

export function localeMeta(locale: Locale) {
  return LOCALES[locale];
}
