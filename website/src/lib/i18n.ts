import type { Locale } from '../data/site';

export function localeFromPath(pathname: string): Locale {
  return pathname === '/zh' || pathname.startsWith('/zh/') ? 'zh' : 'en';
}

export function toLocalePath(pathname: string, locale: Locale): string {
  const clean = pathname.length > 1 && pathname.endsWith('/') ? pathname.slice(0, -1) : pathname;
  const base = clean === '/zh' ? '/' : clean.startsWith('/zh/') ? clean.slice(3) : clean;

  if (locale === 'zh') {
    return base === '/' ? '/zh' : `/zh${base}`;
  }

  return base || '/';
}

export function switchLocalePath(pathname: string): string {
  return localeFromPath(pathname) === 'zh' ? toLocalePath(pathname, 'en') : toLocalePath(pathname, 'zh');
}
