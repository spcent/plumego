import fs from 'node:fs/promises';
import path from 'node:path';

import { readRepoFile, REPO_ROOT, toTsModule, writeGeneratedFile } from './_shared.mjs';

function parseVersion(readme) {
  const versionMatch = readme.match(/version-(v[0-9A-Za-z.]+(?:--[0-9A-Za-z.]+)*)-blue/);
  if (versionMatch) return versionMatch[1].replaceAll('--', '-');

  const statusMatch = readme.match(/status-([0-9A-Za-z.]+(?:--[0-9A-Za-z.]+)*)-[A-Za-z]+/);
  return statusMatch ? statusMatch[1].replaceAll('--', '-') : 'v0.0.0';
}

function parseSupportMatrix(section) {
  const rows = section
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.startsWith('|'));

  return rows.slice(2).map((row) => {
    const [area, status, promise, modules] = row
      .split('|')
      .slice(1, -1)
      .map((cell) => cell.trim().replaceAll('`', ''));

    return { area, status, promise, modules };
  });
}

function parseListBlock(text, pattern) {
  const source = text.endsWith('\n') ? text : `${text}\n`;
  const block = source.match(pattern)?.[1] ?? '';
  return block
    .split('\n')
    .map((line) => line.replace(/^\s*-\s+/, '').trim())
    .filter(Boolean);
}

function yamlSection(text, startPattern, endPattern) {
  const start = text.search(startPattern);
  if (start === -1) return '';
  const rest = text.slice(start);
  const end = rest.search(endPattern);
  return end === -1 ? rest : rest.slice(0, end);
}

async function moduleManifestPaths(root) {
  const entries = await fs.readdir(root, { withFileTypes: true });
  const files = await Promise.all(
    entries.map(async (entry) => {
      const fullPath = path.join(root, entry.name);
      if (entry.isDirectory()) {
        return moduleManifestPaths(fullPath);
      }
      return entry.isFile() && entry.name === 'module.yaml' ? [fullPath] : [];
    }),
  );
  return files.flat();
}

async function extensionModulePaths() {
  const manifests = await moduleManifestPaths(path.join(REPO_ROOT, 'x'));
  return manifests
    .map((manifest) => path.relative(REPO_ROOT, path.dirname(manifest)))
    .sort();
}

async function fallbackSupportMatrix() {
  const repoSpec = await readRepoFile('specs/repo.yaml');
  const routingSpec = await readRepoFile('specs/task-routing.yaml');

  const stableLayer = yamlSection(repoSpec, /\n  stable:\n/m, /\n\s*\n  extension:\n/m);
  const stableRoots = parseListBlock(stableLayer, /paths:\n((?:\s+- .+\n)+)/m);
  const betaFamilies = parseListBlock(routingSpec, /beta_families:\n((?:\s+- .+\n)+)/m);
  const primaryFamilies = parseListBlock(routingSpec, /primary_families:\n((?:\s+- .+\n)+)/m);
  const extensionModules = await extensionModulePaths();
  const primarySet = new Set(primaryFamilies);
  const betaSet = new Set(betaFamilies);
  const appFacingExperimental = primaryFamilies.filter((name) => !betaSet.has(name));
  const subordinate = extensionModules.filter((name) => !primarySet.has(name) && !betaSet.has(name));

  return [
    {
      area: 'Stable library roots',
      status: 'ga',
      promise: 'Public package surface carries the v1 stable-root compatibility promise',
      modules: stableRoots.join(', '),
    },
    {
      area: 'Canonical reference app',
      status: 'supported reference',
      promise:
        'Kept aligned with the canonical bootstrap and stable-root usage, but not treated as a reusable extension catalog',
      modules: 'reference/standard-service',
    },
    {
      area: 'CLI',
      status: 'supported tool',
      promise:
        'Supported as a command-line tool, not as a Go import surface; command behavior and generated output must stay aligned with canonical docs',
      modules: 'cmd/plumego',
    },
    {
      area: 'Beta extension families',
      status: 'beta',
      promise:
        'API surface frozen between minor release refs; promoted after two consecutive tagged refs with no exported API changes and owner sign-off',
      modules: betaFamilies.join(', '),
    },
    {
      area: 'App-facing extension families',
      status: 'experimental',
      promise: 'Included in repo quality gates and release scope, but API/config compatibility is not frozen',
      modules: appFacingExperimental.join(', '),
    },
    {
      area: 'Subordinate extension primitives',
      status: 'experimental',
      promise:
        'Maintained and tested, but discovery should start from the owning family entrypoint and compatibility is not frozen',
      modules: subordinate.join(', '),
    },
  ];
}

export async function syncReleaseMeta() {
  const readme = await readRepoFile('README.md');
  const version = parseVersion(readme);
  const supportMatrixSection =
    readme.match(/## (?:v1 Support Matrix|Current Support Matrix)[\s\S]*?\n((?:\|.*\|\n)+)/m)?.[1] ?? '';
  const supportMatrixFromReadme = parseSupportMatrix(supportMatrixSection);
  const supportMatrix =
    supportMatrixFromReadme.length > 0 ? supportMatrixFromReadme : await fallbackSupportMatrix();

  if (supportMatrix.length === 0) {
    throw new Error('release support matrix source produced no rows');
  }
  for (const row of supportMatrix) {
    if (!row.modules) {
      throw new Error(`release support matrix row ${row.area} has no modules`);
    }
  }

  await writeGeneratedFile(
    'releases.ts',
    toTsModule('RELEASE_FACTS', {
      currentVersion: version,
      supportMatrix,
    }),
  );
}

if (import.meta.url === `file://${process.argv[1]}`) {
  await syncReleaseMeta();
}
