package app

import (
	"io/fs"
)

// fsSub is a thin wrapper around fs.Sub kept here so the routes file does
// not pull in io/fs directly.
func fsSub(fsys fs.FS, dir string) (fs.FS, error) {
	return fs.Sub(fsys, dir)
}
