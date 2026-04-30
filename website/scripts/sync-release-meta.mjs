import { readRepoFile, toTsModule, writeGeneratedFile } from './_shared.mjs';

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

export async function syncReleaseMeta() {
  const readme = await readRepoFile('README.md');
  const version = parseVersion(readme);
  const supportMatrixSection =
    readme.match(/## (?:v1 Support Matrix|Current Support Matrix)[\s\S]*?\n((?:\|.*\|\n)+)/m)?.[1] ?? '';
  const supportMatrix = parseSupportMatrix(supportMatrixSection);

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
