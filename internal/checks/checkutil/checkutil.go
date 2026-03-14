package checkutil

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const modulePath = "github.com/spcent/plumego"

var StableRoots = []string{
	"core",
	"router",
	"contract",
	"middleware",
	"security",
	"store",
	"health",
	"log",
	"metrics",
}

var allowedTopLevelDirs = []string{
	"cmd",
	"contract",
	"core",
	"docs",
	"health",
	"internal",
	"log",
	"metrics",
	"middleware",
	"reference",
	"router",
	"security",
	"specs",
	"store",
	"tasks",
	"x",
}

var requiredManifestKeys = []string{
	"name",
	"path",
	"layer",
	"status",
	"summary",
	"responsibilities",
	"non_goals",
	"allowed_imports",
	"forbidden_imports",
	"test_commands",
	"review_checklist",
	"agent_hints",
}

var manifestListLimits = map[string]int{
	"responsibilities":  7,
	"non_goals":         7,
	"allowed_imports":   12,
	"forbidden_imports": 20,
	"test_commands":     3,
	"review_checklist":  5,
	"agent_hints":       3,
}

var manifestEnums = map[string]map[string]struct{}{
	"layer": {
		"stable":    {},
		"extension": {},
		"tooling":   {},
		"reference": {},
	},
	"status": {
		"ga":           {},
		"beta":         {},
		"experimental": {},
	},
}

func AllowedTopLevelDirs() map[string]struct{} {
	out := make(map[string]struct{}, len(allowedTopLevelDirs))
	for _, dir := range allowedTopLevelDirs {
		out[dir] = struct{}{}
	}
	return out
}

func ReadBaseline(path string) (map[string]struct{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	out := map[string]struct{}{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func FindDisallowedImports(repoRoot string, baseline map[string]struct{}) ([]string, error) {
	blockedPrefixes := []string{
		modulePath + "/config",
		modulePath + "/net",
		modulePath + "/tenant",
		modulePath + "/utils",
		modulePath + "/validator",
		modulePath + "/x/",
	}
	var violations []string

	for _, root := range StableRoots {
		dir := filepath.Join(repoRoot, root)
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				return fmt.Errorf("parse imports %s: %w", path, err)
			}

			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			relPath = filepath.ToSlash(relPath)

			for _, imp := range node.Imports {
				importPath, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return fmt.Errorf("unquote import %s: %w", path, err)
				}
				if !isDisallowedImport(importPath, blockedPrefixes) {
					continue
				}

				key := relPath + "|" + importPath
				if _, ok := baseline[key]; ok {
					continue
				}
				violations = append(violations, key)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(violations)
	return violations, nil
}

func FindMissingModuleManifests(repoRoot string, baseline map[string]struct{}) ([]string, error) {
	var missing []string

	for _, root := range StableRoots {
		path := filepath.Join(repoRoot, root, "module.yaml")
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				missing = append(missing, filepath.ToSlash(filepath.Join(root, "module.yaml")))
				continue
			}
			return nil, err
		}
	}

	xDir := filepath.Join(repoRoot, "x")
	entries, err := os.ReadDir(xDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		relModule := filepath.ToSlash(filepath.Join("x", entry.Name()))
		modulePath := filepath.Join(xDir, entry.Name(), "module.yaml")
		if _, err := os.Stat(modulePath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return nil, err
		}

		if _, ok := baseline[relModule]; ok {
			continue
		}
		missing = append(missing, relModule)
	}

	sort.Strings(missing)
	return missing, nil
}

func FindUnexpectedTopLevelDirs(repoRoot string, allowed, baseline map[string]struct{}) ([]string, error) {
	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		return nil, err
	}

	var unexpected []string
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}
		if _, ok := allowed[name]; ok {
			continue
		}
		if _, ok := baseline[name]; ok {
			continue
		}
		unexpected = append(unexpected, name)
	}

	sort.Strings(unexpected)
	return unexpected, nil
}

func ValidateModuleManifests(repoRoot string) ([]string, error) {
	patterns := []string{
		filepath.Join(repoRoot, "*", "module.yaml"),
		filepath.Join(repoRoot, "x", "*", "module.yaml"),
	}
	var paths []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		paths = append(paths, matches...)
	}
	sort.Strings(paths)

	var violations []string
	for _, path := range paths {
		items, err := validateModuleManifest(repoRoot, path)
		if err != nil {
			return nil, err
		}
		violations = append(violations, items...)
	}

	sort.Strings(violations)
	return violations, nil
}

type manifestDoc struct {
	Seen       map[string]bool
	Scalars    map[string]string
	ListCounts map[string]int
	Lists      map[string][]string
}

func validateModuleManifest(repoRoot, path string) ([]string, error) {
	doc, err := parseManifest(path)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(repoRoot, filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	relPath = filepath.ToSlash(relPath)

	var violations []string
	for _, key := range requiredManifestKeys {
		if !doc.Seen[key] {
			violations = append(violations, fmt.Sprintf("%s: missing required key %q", filepath.ToSlash(path), key))
			continue
		}

		if _, isEnum := manifestEnums[key]; isEnum {
			continue
		}
		if maxCount, isList := manifestListLimits[key]; isList {
			if doc.ListCounts[key] == 0 {
				violations = append(violations, fmt.Sprintf("%s: required list %q must not be empty", filepath.ToSlash(path), key))
			}
			if doc.ListCounts[key] > maxCount {
				violations = append(violations, fmt.Sprintf("%s: list %q exceeds limit %d", filepath.ToSlash(path), key, maxCount))
			}
			continue
		}
		if strings.TrimSpace(doc.Scalars[key]) == "" {
			violations = append(violations, fmt.Sprintf("%s: required scalar %q must not be empty", filepath.ToSlash(path), key))
		}
	}

	for key, allowed := range manifestEnums {
		value := strings.TrimSpace(doc.Scalars[key])
		if value == "" {
			continue
		}
		if _, ok := allowed[value]; !ok {
			violations = append(violations, fmt.Sprintf("%s: invalid %s %q", filepath.ToSlash(path), key, value))
		}
	}

	if pathValue := strings.TrimSpace(doc.Scalars["path"]); pathValue != "" && pathValue != relPath {
		violations = append(violations, fmt.Sprintf("%s: path %q does not match directory %q", filepath.ToSlash(path), pathValue, relPath))
	}
	if nameValue := strings.TrimSpace(doc.Scalars["name"]); nameValue != "" && nameValue != relPath {
		violations = append(violations, fmt.Sprintf("%s: name %q should match module path %q", filepath.ToSlash(path), nameValue, relPath))
	}

	if strings.HasPrefix(relPath, "x/") {
		if layer := strings.TrimSpace(doc.Scalars["layer"]); layer != "" && layer != "extension" {
			violations = append(violations, fmt.Sprintf("%s: x/* modules must declare layer \"extension\"", filepath.ToSlash(path)))
		}
	} else if isStableRoot(relPath) {
		if layer := strings.TrimSpace(doc.Scalars["layer"]); layer != "" && layer != "stable" {
			violations = append(violations, fmt.Sprintf("%s: stable root modules must declare layer \"stable\"", filepath.ToSlash(path)))
		}
	}

	for _, docPath := range doc.Lists["doc_paths"] {
		targetPath := filepath.Join(repoRoot, filepath.FromSlash(docPath))
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				violations = append(violations, fmt.Sprintf("%s: doc_paths target %q does not exist", filepath.ToSlash(path), docPath))
				continue
			}
			return nil, err
		}
	}

	return violations, nil
}

func parseManifest(path string) (manifestDoc, error) {
	file, err := os.Open(path)
	if err != nil {
		return manifestDoc{}, err
	}
	defer file.Close()

	doc := manifestDoc{
		Seen:       map[string]bool{},
		Scalars:    map[string]string{},
		ListCounts: map[string]int{},
		Lists:      map[string][]string{},
	}

	currentKey := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			currentKey = strings.TrimSpace(parts[0])
			doc.Seen[currentKey] = true
			value := strings.TrimSpace(parts[1])
			if value != "" {
				doc.Scalars[currentKey] = strings.Trim(value, "\"'")
			}
			continue
		}

		if currentKey == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "- ") {
			doc.ListCounts[currentKey]++
			value := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
			value = strings.Trim(value, "\"'")
			if value != "" {
				doc.Lists[currentKey] = append(doc.Lists[currentKey], value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return manifestDoc{}, err
	}
	return doc, nil
}

func isDisallowedImport(importPath string, blockedPrefixes []string) bool {
	if importPath == modulePath {
		return true
	}
	for _, prefix := range blockedPrefixes {
		if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") || strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}

func isStableRoot(path string) bool {
	for _, root := range StableRoots {
		if path == root {
			return true
		}
	}
	return false
}
