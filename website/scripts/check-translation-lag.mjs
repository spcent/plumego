/**
 * check-translation-lag.mjs
 *
 * Compares en and zh doc files by git commit timestamp.
 * Any zh file that is more than LAG_DAYS behind its en counterpart gets
 * flagged in src/generated/translation-lag.ts so components can render
 * a visible stale-translation notice.
 *
 * Usage:
 *   node ./scripts/check-translation-lag.mjs
 *
 * Output:
 *   src/generated/translation-lag.ts  — map of translationKey → lagDays
 */

import { execSync } from 'node:child_process';
import { readdirSync, statSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { writeGeneratedFile, toTsModule, WEBSITE_ROOT } from './_shared.mjs';

const LAG_DAYS = 14;

const DOCS_EN = path.join(WEBSITE_ROOT, 'src', 'content', 'docs', 'docs');
const DOCS_ZH = path.join(WEBSITE_ROOT, 'src', 'content', 'docs', 'zh', 'docs');

function gitTimestamp(filePath) {
  try {
    const out = execSync(`git log -1 --format=%ct -- "${filePath}"`, {
      cwd: path.resolve(WEBSITE_ROOT, '..'),
      encoding: 'utf8',
    }).trim();
    return out ? parseInt(out, 10) : null;
  } catch {
    return null;
  }
}

function extractTranslationKey(filePath) {
  const content = readFileSync(filePath, 'utf8');
  const m = content.match(/^translationKey:\s*(.+)$/m);
  return m ? m[1].trim() : null;
}

function walkMdx(dir, lang, keyMap) {
  function recurse(current) {
    for (const entry of readdirSync(current)) {
      const full = path.join(current, entry);
      if (statSync(full).isDirectory()) {
        recurse(full);
      } else if (entry.endsWith('.mdx')) {
        const key = extractTranslationKey(full);
        if (key) {
          const existing = keyMap.get(key) ?? {};
          if (lang === 'en') {
            existing.enPath = full;
          } else {
            existing.zhPath = full;
          }
          keyMap.set(key, existing);
        }
      }
    }
  }
  recurse(dir);
}

export async function checkTranslationLag() {
  const keyMap = new Map();

  walkMdx(DOCS_EN, 'en', keyMap);
  walkMdx(DOCS_ZH, 'zh', keyMap);

  const lagMap = {};

  for (const [key, { enPath, zhPath }] of keyMap.entries()) {
    if (!enPath || !zhPath) continue;

    const enTs = gitTimestamp(enPath);
    const zhTs = gitTimestamp(zhPath);

    if (!enTs || !zhTs) continue;

    const lagSeconds = enTs - zhTs;
    const lagDays = Math.floor(lagSeconds / 86400);

    if (lagDays > LAG_DAYS) {
      lagMap[key] = lagDays;
    }
  }

  await writeGeneratedFile(
    'translation-lag.ts',
    toTsModule('TRANSLATION_LAG', lagMap),
  );

  const flagged = Object.keys(lagMap).length;
  if (flagged > 0) {
    console.warn(
      `[check-translation-lag] ${flagged} zh page(s) are more than ${LAG_DAYS} days behind their en source.`,
    );
    for (const [key, days] of Object.entries(lagMap)) {
      console.warn(`  ${key}: ${days} days behind`);
    }
  } else {
    console.log(`[check-translation-lag] All zh pages are within ${LAG_DAYS}-day threshold.`);
  }

  return flagged;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const strict = process.argv.includes('--strict');
  const flagged = await checkTranslationLag();
  if (strict && flagged > 0) {
    process.exit(1);
  }
}
