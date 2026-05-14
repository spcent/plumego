package frontend

import (
	"net/http"
	"path"
	"strconv"
	"strings"
)

type weightedToken struct {
	value   string
	quality float64
}

func parseWeightedToken(part string) (weightedToken, bool) {
	part = strings.TrimSpace(part)
	if part == "" {
		return weightedToken{}, false
	}

	pieces := strings.Split(part, ";")
	value := strings.ToLower(strings.TrimSpace(pieces[0]))
	if value == "" {
		return weightedToken{}, false
	}

	q := 1.0
	for _, param := range pieces[1:] {
		key, value, ok := strings.Cut(strings.TrimSpace(param), "=")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "q") {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil || parsed < 0 || parsed > 1 {
			return weightedToken{}, false
		}
		q = parsed
	}
	return weightedToken{value: value, quality: q}, true
}

func shouldFallbackToIndex(r *http.Request, filePath string) bool {
	if path.Ext(filePath) != "" {
		return false
	}
	accept := strings.TrimSpace(r.Header.Get("Accept"))
	if accept == "" {
		return true
	}
	return acceptsHTML(accept)
}

func acceptsHTML(header string) bool {
	for _, part := range strings.Split(header, ",") {
		token, ok := parseWeightedToken(part)
		if !ok || token.quality <= 0 {
			continue
		}
		switch token.value {
		case "text/html", "application/xhtml+xml", "*/*", "text/*":
			return true
		}
	}
	return false
}
