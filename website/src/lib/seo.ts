import type { Locale } from '../data/site';
import { toLocalePath } from './i18n';

const SITE_ORIGIN = 'https://plumego.dev';

export interface MarketingSeoInput {
  title: string;
  description: string;
  currentPath: string;
  locale: Locale;
  imagePath?: string;
  noindex?: boolean;
  type?: 'website' | 'article';
  keywords?: string[];
  includeAlternates?: boolean;
}

export function buildMarketingSeo(input: MarketingSeoInput) {
  const {
    title,
    description,
    currentPath,
    locale,
    imagePath = locale === 'zh' ? '/brand/og-zh.svg' : '/brand/og-default.svg',
    noindex = false,
    type = 'website',
    keywords = [],
    includeAlternates = true,
  } = input;

  const canonicalPath = toLocalePath(currentPath, 'en');
  const zhPath = toLocalePath(currentPath, 'zh');
  const canonicalUrl = `${SITE_ORIGIN}${locale === 'zh' ? zhPath : canonicalPath}`;

  return {
    siteOrigin: SITE_ORIGIN,
    canonicalPath,
    zhPath,
    canonicalUrl,
    imageUrl: `${SITE_ORIGIN}${imagePath}`,
    ogLocale: locale === 'zh' ? 'zh_CN' : 'en_US',
    alternateLocale: locale === 'zh' ? 'en_US' : 'zh_CN',
    robots: noindex ? 'noindex,follow' : 'index,follow',
    type,
    keywords: keywords.join(', '),
    description,
    title,
    includeAlternates,
  };
}
