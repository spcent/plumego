//go:build !prod

package web

import "embed"

// Assets is a dummy variable for development mode.
// In production builds (-tags prod), this is replaced with the real embedded FS.
// In development, the frontend is served from disk via x/frontend.RegisterFromDir.
//
// Run "cd web && npm run build" to populate web/dist, then "make build" or
// "go build -tags prod" to produce a single binary with embedded assets.
var Assets embed.FS

// Available indicates whether the frontend assets are embedded into the binary.
const Available = false
