package config

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// parseDotEnv reads a .env file and returns a key->value map.
// Blank lines and lines starting with '#' are ignored.
// Values may be optionally quoted with single or double quotes; quotes are stripped.
// No variable expansion is performed.
func parseDotEnv(r io.Reader) (map[string]string, error) {
	out := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = stripQuotes(val)
		// Drop any trailing inline comment for unquoted values.
		if !strings.HasPrefix(strings.TrimSpace(line[idx+1:]), "\"") &&
			!strings.HasPrefix(strings.TrimSpace(line[idx+1:]), "'") {
			if ci := strings.Index(val, " #"); ci >= 0 {
				val = strings.TrimSpace(val[:ci])
			}
		}
		out[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// stripQuotes removes matching surrounding single or double quotes.
func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// loadDotEnvFile parses the .env file at path. If the file is missing or unreadable,
// it returns an empty map (missing .env is not an error).
func loadDotEnvFile(path string) map[string]string {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	m, err := parseDotEnv(f)
	if err != nil {
		return nil
	}
	return m
}

// combineLookup returns a lookup function where real environment variables take
// precedence over .env file values. If neither has a value, returns ("", false).
func combineLookup(dotenv map[string]string, envLookup func(string) (string, bool)) func(string) (string, bool) {
	return func(key string) (string, bool) {
		if envLookup != nil {
			if v, ok := envLookup(key); ok {
				return v, true
			}
		}
		if dotenv != nil {
			if v, ok := dotenv[key]; ok {
				return v, true
			}
		}
		return "", false
	}
}
