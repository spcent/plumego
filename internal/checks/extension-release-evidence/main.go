package main

import (
	"archive/tar"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	modulePattern := flag.String("module", "", "package pattern to snapshot, for example ./x/rest or ./x/rest/...")
	baseRef := flag.String("base", "", "base git ref")
	headRef := flag.String("head", "", "head git ref")
	outDir := flag.String("out-dir", "", "optional directory for generated snapshots")
	flag.Parse()

	if *modulePattern == "" {
		failf("-module is required")
	}
	if *baseRef == "" {
		failf("-base is required")
	}
	if *headRef == "" {
		failf("-head is required")
	}

	repoRoot, err := gitOutput("", "rev-parse", "--show-toplevel")
	if err != nil {
		failf("resolve repo root: %v", err)
	}
	repoRoot = strings.TrimSpace(repoRoot)

	report, err := run(repoRoot, *modulePattern, *baseRef, *headRef, *outDir)
	if err != nil {
		failf("%v", err)
	}
	fmt.Print(report)
}

func run(repoRoot, modulePattern, baseRef, headRef, outDir string) (string, error) {
	tmpRoot, err := os.MkdirTemp("", "plumego-extension-release-evidence-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpRoot)

	baseTree := filepath.Join(tmpRoot, "base")
	headTree := filepath.Join(tmpRoot, "head")
	if err := extractRef(repoRoot, baseTree, baseRef); err != nil {
		return "", err
	}
	if err := extractRef(repoRoot, headTree, headRef); err != nil {
		return "", err
	}

	snapshotDir := tmpRoot
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return "", fmt.Errorf("create out dir: %w", err)
		}
		snapshotDir = outDir
	}

	baseSnapshot := filepath.Join(snapshotDir, "base.snapshot")
	headSnapshot := filepath.Join(snapshotDir, "head.snapshot")
	if err := snapshotRef(baseTree, modulePattern, baseSnapshot); err != nil {
		return "", fmt.Errorf("snapshot base %s: %w", baseRef, err)
	}
	if err := snapshotRef(headTree, modulePattern, headSnapshot); err != nil {
		return "", fmt.Errorf("snapshot head %s: %w", headRef, err)
	}

	compare, err := command(repoRoot, "go", "run", "./internal/checks/extension-api-snapshot", "-compare", baseSnapshot, headSnapshot)
	var report bytes.Buffer
	fmt.Fprintf(&report, "module\t%s\n", modulePattern)
	fmt.Fprintf(&report, "base\t%s\n", baseRef)
	fmt.Fprintf(&report, "head\t%s\n", headRef)
	if outDir != "" {
		fmt.Fprintf(&report, "base_snapshot\t%s\n", filepath.ToSlash(baseSnapshot))
		fmt.Fprintf(&report, "head_snapshot\t%s\n", filepath.ToSlash(headSnapshot))
	}
	if err != nil {
		report.WriteString("api\tchanged\n")
		report.WriteString(compare)
		return report.String(), fmt.Errorf("release evidence found exported API changes")
	}
	report.WriteString("api\tunchanged\n")
	report.WriteString(compare)
	return report.String(), nil
}

func extractRef(repoRoot, path, ref string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create source tree for %s: %w", ref, err)
	}
	archive, err := commandBytes(repoRoot, "git", "archive", "--format=tar", ref)
	if err != nil {
		return fmt.Errorf("archive %s: %w", ref, err)
	}
	return extractTar(bytes.NewReader(archive), path)
}

func snapshotRef(worktree, modulePattern, outPath string) error {
	_, err := command(worktree, "go", "run", "./internal/checks/extension-api-snapshot", "-module", modulePattern, "-out", outPath)
	return err
}

func gitOutput(dir string, args ...string) (string, error) {
	return command(dir, "git", args...)
}

func command(dir, name string, args ...string) (string, error) {
	out, err := commandBytes(dir, name, args...)
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func commandBytes(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return out.Bytes(), fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, out.String())
	}
	return out.Bytes(), nil
}

func extractTar(reader io.Reader, dest string) error {
	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if filepath.IsAbs(header.Name) {
			return fmt.Errorf("archive path is absolute: %s", header.Name)
		}
		cleanName := filepath.Clean(header.Name)
		target := filepath.Join(destAbs, cleanName)
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(destAbs, targetAbs)
		if err != nil {
			return err
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("archive path escapes destination: %s", header.Name)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetAbs, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
				return err
			}
			file, err := os.OpenFile(targetAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode)&0o777)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(file, tr)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			// Skip symlinks; Go source snapshots do not need them and skipping
			// avoids link traversal concerns in evidence generation.
			continue
		}
	}
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
