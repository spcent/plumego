package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func parsePatterns(s string) []string {
	if s == "" {
		return nil
	}

	patterns := strings.Split(s, ",")
	result := make([]string, 0, len(patterns))

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}

func parseCommandLine(input string) (string, []string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil, nil
	}

	var args []string
	var buf strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if buf.Len() > 0 {
			args = append(args, buf.String())
			buf.Reset()
		}
	}

	for _, r := range input {
		if escaped {
			buf.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingle {
			escaped = true
			continue
		}

		if r == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}

		if r == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if !inSingle && !inDouble && unicode.IsSpace(r) {
			flush()
			continue
		}

		buf.WriteRune(r)
	}

	if escaped {
		return "", nil, fmt.Errorf("unfinished escape sequence")
	}
	if inSingle || inDouble {
		return "", nil, fmt.Errorf("unterminated quote")
	}

	flush()
	if len(args) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}

	return args[0], args[1:], nil
}

// getExecutableDir returns the directory containing the plumego executable.
func getExecutableDir() string {
	ex, err := os.Executable()
	if err != nil {
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(ex)
}
