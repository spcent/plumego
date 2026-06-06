package importer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ScannedFile is a single .md file found by the directory scanner.
type ScannedFile struct {
	AbsPath   string
	RelPath   string
	SizeBytes int64
}

// ScanDirectory walks root recursively and returns all .md files up to maxFiles.
// maxFiles <= 0 means no limit.
func ScanDirectory(root string, maxFiles int) ([]ScannedFile, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("scan directory: root must not be empty")
	}
	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("scan directory %q: root must be absolute", root)
	}
	root = filepath.Clean(root)

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("scan directory %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", root)
	}

	var files []ScannedFile
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			// Skip hidden directories.
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}

		files = append(files, ScannedFile{
			AbsPath:   path,
			RelPath:   rel,
			SizeBytes: fi.Size(),
		})

		if maxFiles > 0 && len(files) >= maxFiles {
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %q: %w", root, err)
	}

	return files, nil
}
