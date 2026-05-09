import { syncModules } from './sync-modules.mjs';
import { syncReleaseMeta } from './sync-release-meta.mjs';
import { syncRoadmap } from './sync-roadmap.mjs';
import { checkTranslationLag } from './check-translation-lag.mjs';

await syncModules();
await syncReleaseMeta();
await syncRoadmap();
await checkTranslationLag();
