package frontend

import (
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
