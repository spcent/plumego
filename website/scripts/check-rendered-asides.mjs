import fs from 'node:fs/promises';
import path from 'node:path';
import { WEBSITE_ROOT } from './_shared.mjs';

const DIST_ROOT = path.join(WEBSITE_ROOT, 'dist');

async function listHTMLFiles(directory) {
  const entries = await fs.readdir(directory, { withFileTypes: true });
  const files = await Promise.all(entries.map(async (entry) => {
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) return listHTMLFiles(fullPath);
    return entry.isFile() && entry.name.endsWith('.html') ? [fullPath] : [];
  }));
  return files.flat();
}

const failures = [];
for (const file of await listHTMLFiles(DIST_ROOT)) {
  const html = await fs.readFile(file, 'utf8');
  if (/&lt;svg\b[\s\S]*?starlight-aside__icon/.test(html)) {
    failures.push(path.relative(WEBSITE_ROOT, file));
  }
}

if (failures.length > 0) {
  console.error('Rendered Starlight aside icons must not be HTML-escaped:');
  for (const file of failures) console.error(`- ${file}`);
  process.exit(1);
}

console.log('Rendered Starlight aside icons are valid SVG elements.');
