//go:build prod

package web

import "embed"

// Assets contains the production frontend build output (web/dist/).
//
// Only compiled into the binary when the "prod" build tag is set:
//
//	go build -tags prod .
//
// The Makefile target "build" sets this flag automatically after running
// the frontend build step.
//
//go:embed dist
var Assets embed.FS

// Available indicates whether the frontend assets are embedded into the binary.
const Available = true
