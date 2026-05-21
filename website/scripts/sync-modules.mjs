import { readRepoFile, writeGeneratedFile, toTsModule } from './_shared.mjs';

export async function syncModules() {
  const repoSpec = await readRepoFile('specs/repo.yaml');
  const routingSpec = await readRepoFile('specs/task-routing.yaml');

  const stableRootsBlock = repoSpec.match(/stable:\n[\s\S]*?paths:\n((?:\s+- .+\n)+)/m)?.[1] ?? '';
  const extensionPathsBlock = repoSpec.match(/extension:\n[\s\S]*?paths:\n((?:\s+- .+\n)+)/m)?.[1] ?? '';
  const discouragedRootsBlock =
    repoSpec.match(/discouraged_roots:\n((?:\s+- .+\n)+)/m)?.[1] ?? '';
  const topLevelRulesBlock = repoSpec.match(/rules:\n((?:\s+- .+\n)+)\n\nagent_workflow:/m)?.[1] ?? '';
  const betaFamiliesBlock =
    routingSpec.match(/beta_families:\n((?:\s+- .+\n)+)/m)?.[1] ?? '';
  const primaryFamiliesBlock =
    routingSpec.match(/primary_families:\n((?:\s+- .+\n)+)/m)?.[1] ?? '';
  const stableRootWorkIntent =
    routingSpec.match(/stable_root_work:\n\s+intent:\s+(.+)\n/m)?.[1]?.trim() ?? '';
  const extensionWorkIntent =
    routingSpec.match(/extension_work:\n\s+intent:\s+(.+)\n/m)?.[1]?.trim() ?? '';

  const stableRoots = stableRootsBlock
    .split('\n')
    .map((line) => line.replace(/^\s*-\s+/, '').trim())
    .filter(Boolean);
  const allExtensionPaths = extensionPathsBlock
    .split('\n')
    .map((line) => line.replace(/^\s*-\s+/, '').trim())
    .filter(Boolean);
  const discouragedRoots = discouragedRootsBlock
    .split('\n')
    .map((line) => line.replace(/^\s*-\s+/, '').trim())
    .filter(Boolean);
  const topLevelRules = topLevelRulesBlock
    .split('\n')
    .map((line) => line.replace(/^\s*-\s+/, '').trim())
    .filter(Boolean);
  const betaExtensionFamilies = betaFamiliesBlock
    .split('\n')
    .map((line) => line.replace(/^\s*-\s+/, '').trim())
    .filter(Boolean);
  const primaryExtensionFamilies = primaryFamiliesBlock
    .split('\n')
    .map((line) => line.replace(/^\s*-\s+/, '').trim())
    .filter(Boolean);
  const betaSet = new Set(betaExtensionFamilies);
  const experimentalFamilies = primaryExtensionFamilies.filter((f) => !betaSet.has(f));

  await writeGeneratedFile(
    'modules.ts',
    toTsModule('MODULE_FACTS', {
      stableRoots,
      allExtensionPaths,
      betaExtensionFamilies,
      experimentalFamilies,
      primaryExtensionFamilies,
      discouragedRoots,
      topLevelRules,
      stableRootWorkIntent,
      extensionWorkIntent,
    }),
  );
}

if (import.meta.url === `file://${process.argv[1]}`) {
  await syncModules();
}
