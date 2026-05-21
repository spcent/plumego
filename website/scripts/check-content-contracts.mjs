import fs from 'node:fs/promises';
import path from 'node:path';
import { REPO_ROOT, WEBSITE_ROOT } from './_shared.mjs';

const PAGES_ROOT = path.join(WEBSITE_ROOT, 'src', 'pages');
const SITE_DATA_FILE = path.join(WEBSITE_ROOT, 'src', 'data', 'site.ts');
const ASTRO_CONFIG_FILE = path.join(WEBSITE_ROOT, 'astro.config.mjs');

const expectedJourney = {
  en: [
    { label: 'Why Plumego', href: '/why-plumego' },
    { label: 'Examples', href: '/examples' },
    { label: 'Releases', href: '/releases' },
  ],
  zh: [
    { label: '为什么选择', href: '/zh/why-plumego' },
    { label: '示例', href: '/zh/examples' },
    { label: '发布', href: '/zh/releases' },
  ],
};

const referenceInventoryExceptions = new Set([
  'benchmark',
  'standard-service',
  'workerfleet',
]);

const canonicalExtensionNames = new Map([
  ['x/cache', 'x/data/cache'],
  ['x/devtools', 'x/observability/devtools'],
  ['x/discovery', 'x/gateway/discovery'],
  ['x/ipc', 'x/gateway/ipc'],
  ['x/mq', 'x/messaging/mq'],
  ['x/ops', 'x/observability/ops'],
  ['x/pubsub', 'x/messaging/pubsub'],
  ['x/scheduler', 'x/messaging/scheduler'],
  ['x/webhook', 'x/messaging/webhook'],
]);

async function listFiles(dir, predicate) {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  const files = await Promise.all(
    entries.map(async (entry) => {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) return listFiles(fullPath, predicate);
      if (entry.isFile() && predicate(fullPath)) return [fullPath];
      return [];
    }),
  );
  return files.flat();
}

function rel(file) {
  return path.relative(REPO_ROOT, file);
}

function lineNumber(text, index) {
  return text.slice(0, index).split('\n').length;
}

function extractJourneyBlock(text) {
  const constMatch = text.match(/const\s+journeySteps\s*=\s*\[([\s\S]*?)\];/);
  if (constMatch) return constMatch[1];

  const inlineMatch = text.match(/<JourneyBar[\s\S]*?steps=\{\[([\s\S]*?)\]\}/);
  return inlineMatch?.[1] ?? '';
}

function journeyPairs(block) {
  return [...block.matchAll(/\{\s*label:\s*'([^']+)'\s*,\s*href:\s*'([^']+)'\s*\}/g)]
    .map((match) => ({ label: match[1], href: match[2] }));
}

function sameJourney(actual, expected) {
  return actual.length === expected.length
    && actual.every((item, index) => item.label === expected[index].label && item.href === expected[index].href);
}

async function checkJourneyBars(failures) {
  const pageFiles = await listFiles(PAGES_ROOT, (file) => file.endsWith('.astro'));

  for (const file of pageFiles) {
    const text = await fs.readFile(file, 'utf8');
    if (!text.includes('<JourneyBar')) continue;

    const locale = file.includes(`${path.sep}zh${path.sep}`) ? 'zh' : 'en';
    const actual = journeyPairs(extractJourneyBlock(text));
    if (!sameJourney(actual, expectedJourney[locale])) {
      failures.push(
        `${rel(file)} JourneyBar must use ${JSON.stringify(expectedJourney[locale])}; found ${JSON.stringify(actual)}`,
      );
    }
  }
}

async function checkReferenceInventory(failures) {
  const entries = await fs.readdir(path.join(REPO_ROOT, 'reference'), { withFileTypes: true });
  const dirs = entries.filter((entry) => entry.isDirectory()).map((entry) => entry.name).sort();
  const site = await fs.readFile(SITE_DATA_FILE, 'utf8');
  const listed = new Set([...site.matchAll(/name:\s*'reference\/([^']+)'/g)].map((match) => match[1]));

  for (const dir of dirs) {
    if (referenceInventoryExceptions.has(dir)) continue;
    if (!listed.has(dir)) {
      failures.push(`reference/${dir} exists but is missing from EXAMPLES_COPY referenceMatrix`);
    }
  }

  for (const dir of listed) {
    if (!dirs.includes(dir)) {
      failures.push(`reference/${dir} is listed in EXAMPLES_COPY referenceMatrix but no directory exists`);
    }
  }
}

async function checkExtensionDisplayNames(failures) {
  const files = [
    SITE_DATA_FILE,
    ASTRO_CONFIG_FILE,
    ...(await listFiles(PAGES_ROOT, (file) => file.endsWith('.astro'))),
  ];

  for (const file of files) {
    const text = await fs.readFile(file, 'utf8');
    for (const [alias, canonical] of canonicalExtensionNames.entries()) {
      const quoted = new RegExp(`(['"\`])${alias.replace('/', '\\/')}\\1`, 'g');
      for (const match of text.matchAll(quoted)) {
        failures.push(`${rel(file)}:${lineNumber(text, match.index)} display ${alias} as ${canonical}`);
      }
    }
  }
}

const failures = [];
await checkJourneyBars(failures);
await checkReferenceInventory(failures);
await checkExtensionDisplayNames(failures);

if (failures.length > 0) {
  console.error('Website content contract check failed:');
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exit(1);
}

console.log('Website content contract check passed.');
