import { DEFAULT_LOCALE, LOCALES, LOCALE_CODES, PREFIXED_LOCALES, type Locale } from './locales';

export function localeFromPath(pathname: string): Locale {
  const clean = normalizePath(pathname);

  for (const locale of PREFIXED_LOCALES) {
    const prefix = LOCALES[locale].pathPrefix;
    if (clean === prefix || clean.startsWith(`${prefix}/`)) {
      return locale;
    }
  }

  return DEFAULT_LOCALE;
}

export function toLocalePath(pathname: string, locale: Locale): string {
  const base = basePath(pathname);
  const prefix = LOCALES[locale].pathPrefix;

  if (!prefix) {
    return base;
  }

  return base === '/' ? prefix : `${prefix}${base}`;
}

export function switchLocalePath(pathname: string): string {
  const current = localeFromPath(pathname);
  const next = LOCALE_CODES.find((locale) => locale !== current) ?? DEFAULT_LOCALE;
  return toLocalePath(pathname, next);
}

function basePath(pathname: string): string {
  const clean = normalizePath(pathname);

  for (const locale of PREFIXED_LOCALES) {
    const prefix = LOCALES[locale].pathPrefix;
    if (clean === prefix) {
      return '/';
    }
    if (clean.startsWith(`${prefix}/`)) {
      return clean.slice(prefix.length) || '/';
    }
  }

  return clean;
}

function normalizePath(pathname: string): string {
  if (!pathname || pathname === '/') {
    return '/';
  }

  const clean = pathname.endsWith('/') ? pathname.slice(0, -1) : pathname;
  return clean.startsWith('/') ? clean : `/${clean}`;
}
