// Package ai documents Plumego's AI capability family.
//
// The x/ai module remains experimental overall. Its manifest defines the
// canonical stability split:
//   - stable-tier subpackages: provider, session, streaming, tool
//   - experimental subpackages: orchestration, semanticcache, marketplace,
//     distributed, resilience
//
// Additional supporting subpackages under x/ai should be treated as
// experimental unless and until the module manifest says otherwise.
//
// Import the owning subpackage directly; this root package is only a module
// marker and does not provide a canonical bootstrap entrypoint.
//
// Keep AI wiring explicit in handlers or owning extensions. Do not add hidden
// provider globals, implicit registration, or stable-root dependencies.
package ai
