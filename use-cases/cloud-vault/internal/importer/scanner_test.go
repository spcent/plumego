package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirectory_notExist(t *testing.T) {
	_, err := ScanDirectory("/tmp/no-such-dir-xyz-123", 0)
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestScanDirectory_rejectsRelativeRoot(t *testing.T) {
	_, err := ScanDirectory("relative/path", 0)
	if err == nil {
		t.Fatal("expected error for relative root")
	}
}

func TestScanDirectory_notDir(t *testing.T) {
	f, err := os.CreateTemp("", "*.md")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	_, err = ScanDirectory(f.Name(), 0)
	if err == nil {
		t.Fatal("expected error when path is a file")
	}
}

func TestScanDirectory_basic(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.md"), "# A")
	writeFile(t, filepath.Join(root, "b.md"), "# B")
	writeFile(t, filepath.Join(root, "c.txt"), "not markdown")

	files, err := ScanDirectory(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("want 2 .md files, got %d", len(files))
	}
	for _, f := range files {
		if filepath.Ext(f.RelPath) != ".md" {
			t.Errorf("unexpected file: %s", f.RelPath)
		}
	}
}

func TestScanDirectory_recursive(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "top.md"), "top")
	writeFile(t, filepath.Join(sub, "nested.md"), "nested")

	files, err := ScanDirectory(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("want 2 files, got %d", len(files))
	}
}

func TestScanDirectory_skipsHiddenDirs(t *testing.T) {
	root := t.TempDir()
	hidden := filepath.Join(root, ".hidden")
	if err := os.Mkdir(hidden, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(hidden, "secret.md"), "hidden")
	writeFile(t, filepath.Join(root, "visible.md"), "visible")

	files, err := ScanDirectory(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("want 1 file (hidden dir skipped), got %d", len(files))
	}
	if files[0].RelPath != "visible.md" {
		t.Errorf("unexpected file: %s", files[0].RelPath)
	}
}

func TestScanDirectory_maxFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.md", "b.md", "c.md", "d.md"} {
		writeFile(t, filepath.Join(root, name), name)
	}

	files, err := ScanDirectory(root, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("want 2 files (maxFiles=2), got %d", len(files))
	}
}

func TestScanDirectory_absPath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "doc.md"), "# Doc")

	files, err := ScanDirectory(root, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatal("expected 1 file")
	}
	if !filepath.IsAbs(files[0].AbsPath) {
		t.Errorf("AbsPath is not absolute: %s", files[0].AbsPath)
	}
	if files[0].RelPath != "doc.md" {
		t.Errorf("RelPath: want 'doc.md', got %q", files[0].RelPath)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
