//go:build !prod

package web

// Available indicates whether the frontend assets are embedded into the binary.
//
// In development mode (no "prod" build tag), the frontend is served from disk
// via x/frontend.RegisterFromDir. Run "cd web && npm run build" to populate
// web/dist, then "make build" or "go build -tags prod" to produce a single
// binary with embedded assets.
const Available = false
