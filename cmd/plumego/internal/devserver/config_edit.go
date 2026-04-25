package devserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
)

const defaultConfigEditFile = ".env"

type ConfigEditEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ConfigEditResponse struct {
	Path    string            `json:"path"`
	Exists  bool              `json:"exists"`
	Entries []ConfigEditEntry `json:"entries"`
	Updated string            `json:"updated_at,omitempty"`
}

type ConfigEditRequest struct {
	Entries []ConfigEditEntry `json:"entries"`
	Restart bool              `json:"restart"`
}

type ConfigEditSaveResponse struct {
	Success   bool   `json:"success"`
	Path      string `json:"path"`
	Count     int    `json:"count"`
	Restarted bool   `json:"restarted,omitempty"`
}

func (d *Dashboard) handleConfigEditGet(w http.ResponseWriter, r *http.Request) {
	path, displayPath, err := d.resolveConfigEditPath()
	if err != nil {
		writeDevserverError(w, r, contract.TypeValidation, devserverCodeConfigEditPathInvalid, "config edit path is invalid")
		return
	}

	entries, exists, modTime, err := readEnvEntries(path)
	if err != nil {
		writeDevserverError(w, r, contract.TypeInternal, devserverCodeConfigEditReadFailed, "config edit file could not be read")
		return
	}

	payload := ConfigEditResponse{
		Path:    displayPath,
		Exists:  exists,
		Entries: entries,
	}
	if exists && !modTime.IsZero() {
		payload.Updated = modTime.Format(time.RFC3339)
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, payload, nil)
}

func (d *Dashboard) handleConfigEditSave(w http.ResponseWriter, r *http.Request) {
	path, displayPath, err := d.resolveConfigEditPath()
	if err != nil {
		writeDevserverError(w, r, contract.TypeValidation, devserverCodeConfigEditPathInvalid, "config edit path is invalid")
		return
	}

	var req ConfigEditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Category(contract.CategoryValidation).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Build())
		return
	}

	entries := normalizeConfigEntries(req.Entries)
	if err := writeEnvEntries(path, entries); err != nil {
		writeDevserverError(w, r, contract.TypeInternal, devserverCodeConfigEditWriteFailed, "config edit file could not be written")
		return
	}

	response := ConfigEditSaveResponse{
		Success: true,
		Path:    displayPath,
		Count:   len(entries),
	}

	if req.Restart {
		if err := d.Rebuild(r.Context()); err != nil {
			writeDevserverError(w, r, contract.TypeInternal, devserverCodeAppRebuildFailed, "application rebuild failed")
			return
		}
		response.Restarted = true
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, response, nil)
}

func (d *Dashboard) resolveConfigEditPath() (string, string, error) {
	envFile := defaultConfigEditFile

	if d.runner.IsRunning() {
		if snapshot, err := d.analyzer.GetAppSnapshot(); err == nil {
			if path := strings.TrimSpace(snapshot.EnvFile); path != "" {
				envFile = path
			}
		}
	}

	path := envFile
	if !filepath.IsAbs(path) {
		path = filepath.Join(d.projectDir, path)
	}
	path = filepath.Clean(path)

	projectDir := filepath.Clean(d.projectDir)
	if !isWithinDir(projectDir, path) {
		return "", "", fmt.Errorf("env file must be within project directory")
	}

	display := envFile
	if filepath.IsAbs(path) {
		if rel, err := filepath.Rel(projectDir, path); err == nil {
			display = rel
		} else {
			display = path
		}
	}

	return path, display, nil
}

func isWithinDir(dir, target string) bool {
	if dir == target {
		return true
	}
	prefix := dir + string(os.PathSeparator)
	return strings.HasPrefix(target, prefix)
}

func readEnvEntries(path string) ([]ConfigEditEntry, bool, time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, time.Time{}, nil
		}
		return nil, false, time.Time{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, true, time.Time{}, err
	}

	scanner := bufio.NewScanner(file)
	entries := make([]ConfigEditEntry, 0, 32)
	seen := make(map[string]int)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}

		if pos, ok := seen[key]; ok {
			entries[pos].Value = value
			continue
		}
		seen[key] = len(entries)
		entries = append(entries, ConfigEditEntry{Key: key, Value: value})
	}

	if err := scanner.Err(); err != nil {
		return nil, true, time.Time{}, err
	}

	return entries, true, info.ModTime(), nil
}

func normalizeConfigEntries(entries []ConfigEditEntry) []ConfigEditEntry {
	normalized := make([]ConfigEditEntry, 0, len(entries))
	seen := make(map[string]int)

	for _, entry := range entries {
		key := strings.TrimSpace(entry.Key)
		value := strings.TrimSpace(entry.Value)
		if key == "" {
			continue
		}
		if idx, ok := seen[key]; ok {
			normalized[idx].Value = value
			continue
		}
		seen[key] = len(normalized)
		normalized = append(normalized, ConfigEditEntry{Key: key, Value: value})
	}

	return normalized
}

func writeEnvEntries(path string, entries []ConfigEditEntry) error {
	builder := strings.Builder{}
	builder.WriteString("# Managed by plumego dev dashboard\n")
	builder.WriteString("# Changes require app restart to take effect\n")

	for _, entry := range entries {
		key := strings.TrimSpace(entry.Key)
		if key == "" {
			continue
		}
		value := strings.TrimSpace(entry.Value)
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(value)
		builder.WriteString("\n")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}
