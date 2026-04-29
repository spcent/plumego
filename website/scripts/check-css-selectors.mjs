import { readFileSync } from 'node:fs';

const strict = process.argv.includes('--strict');
const files = [
  'src/styles/global.css',
  'src/styles/home.css',
  'src/styles/prose.css',
  'src/styles/starlight-bridge.css',
];

const stripComments = (input) => input.replace(/\/\*[\s\S]*?\*\//g, '');

const lineAt = (input, offset) => input.slice(0, offset).split('\n').length;

const selectorEntries = (file) => {
  const css = stripComments(readFileSync(file, 'utf8'));
  const entries = [];
  const stack = [];
  let tokenStart = 0;

  for (let i = 0; i < css.length; i += 1) {
    const char = css[i];

    if (char === '{') {
      const raw = css.slice(tokenStart, i).trim();
      const line = lineAt(css, tokenStart);
      const parent = stack[stack.length - 1];

      if (raw.startsWith('@')) {
        stack.push(raw);
      } else if (raw && !parent?.startsWith('@keyframes')) {
        const context = stack.filter((item) => item.startsWith('@media') || item.startsWith('@supports')).join(' ');
        for (const selector of raw.split(',').map((item) => item.trim()).filter(Boolean)) {
          entries.push({ context, file, line, selector });
        }
        stack.push('{rule}');
      } else {
        stack.push('{rule}');
      }

      tokenStart = i + 1;
      continue;
    }

    if (char === '}') {
      stack.pop();
      tokenStart = i + 1;
    }
  }

  return entries;
};

const byKey = new Map();

for (const file of files) {
  for (const entry of selectorEntries(file)) {
    const key = `${entry.context}\n${entry.selector}`;
    const matches = byKey.get(key) ?? [];
    matches.push(entry);
    byKey.set(key, matches);
  }
}

const duplicates = [...byKey.entries()]
  .filter(([, matches]) => matches.length > 1)
  .sort(([, a], [, b]) => b.length - a.length || a[0].selector.localeCompare(b[0].selector));

if (duplicates.length === 0) {
  console.log('No duplicate CSS selectors found in the checked style files.');
  process.exit(0);
}

console.log(`Duplicate CSS selector groups: ${duplicates.length}`);
for (const [key, matches] of duplicates) {
  const [context, selector] = key.split('\n');
  const label = context ? `${context} :: ${selector}` : selector;
  const locations = matches.map((entry) => `${entry.file}:${entry.line}`).join(', ');
  console.log(`- ${label}`);
  console.log(`  ${locations}`);
}

if (strict) {
  process.exit(1);
}
