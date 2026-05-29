// Package web embeds the guardus React + Vite SPA.
//
// The static/ subdirectory is the output of `make frontend-build` (which
// runs `npm run build` inside web/app/); rebuild it whenever the SPA
// source changes.
package web

import "embed"

//go:embed static
var FileSystem embed.FS

const (
	RootPath  = "static"
	IndexPath = RootPath + "/index.html"
)
