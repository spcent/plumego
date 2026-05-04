package frontend

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// acceptsToken reports whether the client's Accept-Encoding header accepts the
// named encoding token. It compares tokens case-insensitively and rejects tokens
// with a zero quality factor.
func acceptsToken(r *http.Request, token string) bool {
	return acceptedEncodings(r.Header.Get("Accept-Encoding")).quality(token) > 0
}

type encodingPreference struct {
	headerPresent bool
	explicit      map[string]float64
	wildcard      float64
}

func acceptedEncodings(header string) encodingPreference {
	pref := encodingPreference{wildcard: -1}
	if header == "" {
		return pref
	}
	pref.headerPresent = true
	for _, part := range strings.Split(header, ",") {
		token, q, ok := parseEncodingToken(part)
		if !ok {
			continue
		}
		if token == "*" {
			pref.wildcard = q
			continue
		}
		if pref.explicit == nil {
			pref.explicit = make(map[string]float64)
		}
		pref.explicit[token] = q
	}
	return pref
}

func (p encodingPreference) quality(token string) float64 {
	token = strings.ToLower(strings.TrimSpace(token))
	if token == "" {
		return 0
	}
	if q, ok := p.explicit[token]; ok {
		return q
	}
	if p.wildcard >= 0 {
		return p.wildcard
	}
	return 0
}

func (p encodingPreference) identityAcceptable() bool {
	if !p.headerPresent {
		return true
	}
	if q, ok := p.explicit["identity"]; ok {
		return q > 0
	}
	if p.wildcard >= 0 {
		return p.wildcard > 0
	}
	return true
}

func parseEncodingToken(part string) (string, float64, bool) {
	part = strings.TrimSpace(part)
	if part == "" {
		return "", 0, false
	}

	pieces := strings.Split(part, ";")
	token := strings.ToLower(strings.TrimSpace(pieces[0]))
	if token == "" {
		return "", 0, false
	}

	q := 1.0
	for _, param := range pieces[1:] {
		key, value, ok := strings.Cut(strings.TrimSpace(param), "=")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "q") {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return "", 0, false
		}
		if parsed < 0 || parsed > 1 {
			return "", 0, false
		}
		q = parsed
	}
	return token, q, true
}

// tryPrecompressed attempts to serve a pre-compressed version of the file.
// Returns (file, stat, encoding) if successful, or (nil, nil, "") if not found.
func (h *handler) tryPrecompressed(r *http.Request, filePath string) (http.File, os.FileInfo, string) {
	if !h.cfg.EnablePrecompressed {
		return nil, nil, ""
	}

	if h.variants.known {
		variants := h.variants.variants[filePath]
		if !variants.any() {
			return nil, nil, ""
		}
		for _, encoding := range preferredPrecompressedEncodings(r) {
			if !variants.has(encoding) {
				continue
			}
			if f, stat := h.tryOpenFile(filePath + precompressedSuffix(encoding)); f != nil {
				return f, stat, encoding
			}
		}
		return nil, nil, ""
	}

	for _, encoding := range preferredPrecompressedEncodings(r) {
		if f, stat := h.tryOpenFile(filePath + precompressedSuffix(encoding)); f != nil {
			return f, stat, encoding
		}
	}

	return nil, nil, ""
}

func preferredPrecompressedEncodings(r *http.Request) []string {
	pref := acceptedEncodings(r.Header.Get("Accept-Encoding"))
	brQ := pref.quality("br")
	gzipQ := pref.quality("gzip")
	switch {
	case brQ <= 0 && gzipQ <= 0:
		return nil
	case brQ >= gzipQ:
		out := []string{}
		if brQ > 0 {
			out = append(out, "br")
		}
		if gzipQ > 0 {
			out = append(out, "gzip")
		}
		return out
	default:
		out := []string{}
		if gzipQ > 0 {
			out = append(out, "gzip")
		}
		if brQ > 0 {
			out = append(out, "br")
		}
		return out
	}
}

func identityEncodingAcceptable(r *http.Request) bool {
	return acceptedEncodings(r.Header.Get("Accept-Encoding")).identityAcceptable()
}

type precompressedVariants struct {
	br   bool
	gzip bool
}

func (v precompressedVariants) has(encoding string) bool {
	switch encoding {
	case "br":
		return v.br
	case "gzip":
		return v.gzip
	default:
		return false
	}
}

func (v precompressedVariants) any() bool {
	return v.br || v.gzip
}

type precompressedVariantPlan struct {
	known    bool
	variants map[string]precompressedVariants
}

func newPrecompressedVariantPlan(fsys http.FileSystem, enabled bool) precompressedVariantPlan {
	if !enabled {
		return precompressedVariantPlan{}
	}
	root, ok := fsys.(localDirFS)
	if !ok {
		return precompressedVariantPlan{}
	}

	plan := precompressedVariantPlan{
		known:    true,
		variants: make(map[string]precompressedVariants),
	}
	_ = filepath.WalkDir(string(root), func(name string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		encoding, ok := precompressedEncodingForPath(name)
		if !ok {
			return nil
		}
		originalName := strings.TrimSuffix(name, precompressedSuffix(encoding))
		info, err := os.Stat(originalName)
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(string(root), originalName)
		if err != nil {
			return nil
		}
		filePath, ok := cleanAssetPath(filepath.ToSlash(rel))
		if !ok {
			return nil
		}
		variants := plan.variants[filePath]
		switch encoding {
		case "br":
			variants.br = true
		case "gzip":
			variants.gzip = true
		}
		plan.variants[filePath] = variants
		return nil
	})
	if len(plan.variants) == 0 {
		plan.variants = nil
	}
	return plan
}

func precompressedEncodingForPath(filePath string) (string, bool) {
	switch {
	case strings.HasSuffix(filePath, ".br"):
		return "br", true
	case strings.HasSuffix(filePath, ".gz"):
		return "gzip", true
	default:
		return "", false
	}
}

func precompressedSuffix(encoding string) string {
	if encoding == "gzip" {
		return ".gz"
	}
	return "." + encoding
}

func (h *handler) hasPrecompressedVariant(filePath string) bool {
	if !h.cfg.EnablePrecompressed {
		return false
	}
	if h.variants.known {
		return h.variants.variants[filePath].any()
	}
	if f, _ := h.tryOpenFile(filePath + ".br"); f != nil {
		f.Close()
		return true
	}
	if f, _ := h.tryOpenFile(filePath + ".gz"); f != nil {
		f.Close()
		return true
	}
	return false
}

// tryOpenFile attempts to open a file and return it with its stat.
func (h *handler) tryOpenFile(filePath string) (http.File, os.FileInfo) {
	f, err := h.fs.Open(filePath)
	if err != nil {
		return nil, nil
	}
	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		f.Close()
		return nil, nil
	}
	return f, stat
}
