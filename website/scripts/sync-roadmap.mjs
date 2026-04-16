import { markdownBulletList, readRepoFile, toTsModule, writeGeneratedFile } from './_shared.mjs';

function parsePhases(markdown) {
  const matches = markdown.matchAll(/## (Phase [^\n]+)\n\nStatus: ([^\n]+)\n/g);
  const phases = [];

  for (const match of matches) {
    phases.push({
      title: match[1].trim(),
      status: match[2].trim(),
    });
  }

  return phases;
}

export async function syncRoadmap() {
  const roadmap = await readRepoFile('docs/ROADMAP.md');
  const currentBaseline = markdownBulletList(
    roadmap.match(/## Current Baseline\s+([\s\S]*?)(?=\n## |\s*$)/)?.[1] ?? '',
  );
  const suggestedExecutionOrder = (roadmap.match(/## Suggested Execution Order\s+([\s\S]*?)(?=\n## |\s*$)/)?.[1] ?? '')
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => /^\d+\.\s/.test(line))
    .map((line) => line.replace(/^\d+\.\s+/, ''));
  const phases = parsePhases(roadmap);

  await writeGeneratedFile(
    'roadmap.ts',
    toTsModule('ROADMAP_FACTS', {
      currentBaseline,
      inProgress: phases.filter((phase) => phase.status === 'in progress').map((phase) => phase.title),
      planned: phases.filter((phase) => phase.status === 'planned').map((phase) => phase.title),
      suggestedExecutionOrder,
    }),
  );
}

if (import.meta.url === `file://${process.argv[1]}`) {
  await syncRoadmap();
}
