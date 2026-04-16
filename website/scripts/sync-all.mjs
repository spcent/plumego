import { syncModules } from './sync-modules.mjs';
import { syncReleaseMeta } from './sync-release-meta.mjs';
import { syncRoadmap } from './sync-roadmap.mjs';

await syncModules();
await syncReleaseMeta();
await syncRoadmap();
