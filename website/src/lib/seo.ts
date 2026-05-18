import { toLocalePath } from './i18n';
import { LOCALES, LOCALE_CODES, type Locale } from './locales';

const SITE_ORIGIN = 'https://plumego.birdor.com';

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
    imagePath = LOCALES[locale].defaultOgImage,
    noindex = false,
    type = 'website',
    keywords = [],
    includeAlternates = true,
  } = input;

  const canonicalPath = toLocalePath(currentPath, locale);
  const alternates = LOCALE_CODES.map((code) => ({
    locale: code,
    path: toLocalePath(currentPath, code),
    hreflang: LOCALES[code].hreflang,
    ogLocale: LOCALES[code].ogLocale,
  }));

  return {
    siteOrigin: SITE_ORIGIN,
    canonicalPath,
    canonicalUrl: `${SITE_ORIGIN}${canonicalPath}`,
    alternates,
    imageUrl: `${SITE_ORIGIN}${imagePath}`,
    ogLocale: LOCALES[locale].ogLocale,
    alternateLocales: alternates
      .filter((alternate) => alternate.locale !== locale)
      .map((alternate) => alternate.ogLocale),
    robots: noindex ? 'noindex,follow' : 'index,follow',
    type,
    keywords: keywords.join(', '),
    description,
    title,
    includeAlternates,
  };
}
