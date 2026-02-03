package devserver

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego"
)

type pprofProfile struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	SupportsSeconds bool   `json:"supports_seconds"`
	DefaultSeconds  int    `json:"default_seconds,omitempty"`
}

const (
	defaultCPUSeconds   = 10
	defaultTraceSeconds = 5
	maxCPUSeconds       = 120
	maxTraceSeconds     = 30
)

func pprofProfiles() []pprofProfile {
	return []pprofProfile{
		{ID: "cpu", Label: "CPU", SupportsSeconds: true, DefaultSeconds: defaultCPUSeconds},
		{ID: "heap", Label: "Heap"},
		{ID: "allocs", Label: "Allocs"},
		{ID: "goroutine", Label: "Goroutine"},
		{ID: "block", Label: "Block"},
		{ID: "mutex", Label: "Mutex"},
		{ID: "threadcreate", Label: "Thread Create"},
		{ID: "trace", Label: "Trace", SupportsSeconds: true, DefaultSeconds: defaultTraceSeconds},
	}
}

func parsePprofRequest(ctx *plumego.Context) (string, int, error) {
	profileType := strings.ToLower(strings.TrimSpace(ctx.Query.Get("type")))
	if profileType == "" {
		profileType = "cpu"
	}

	if !isSupportedProfile(profileType) {
		return "", 0, fmt.Errorf("unsupported profile type: %s", profileType)
	}

	seconds := 0
	if secondsRaw := strings.TrimSpace(ctx.Query.Get("seconds")); secondsRaw != "" {
		parsed, err := strconv.Atoi(secondsRaw)
		if err != nil {
			return "", 0, fmt.Errorf("invalid seconds")
		}
		seconds = parsed
	}

	return profileType, clampProfileSeconds(profileType, seconds), nil
}

func isSupportedProfile(profileType string) bool {
	for _, prof := range pprofProfiles() {
		if prof.ID == profileType {
			return true
		}
	}
	return false
}

func clampProfileSeconds(profileType string, seconds int) int {
	if profileType == "trace" {
		if seconds <= 0 {
			return defaultTraceSeconds
		}
		if seconds > maxTraceSeconds {
			return maxTraceSeconds
		}
		return seconds
	}

	if profileType == "cpu" {
		if seconds <= 0 {
			return defaultCPUSeconds
		}
		if seconds > maxCPUSeconds {
			return maxCPUSeconds
		}
		return seconds
	}

	return 0
}

func (a *Analyzer) FetchPprof(profileType string, seconds int) ([]byte, string, error) {
	path, query := pprofPath(profileType, seconds)
	url := a.appURL + path
	if query != "" {
		url += "?" + query
	}

	timeout := 30 * time.Second
	if profileType == "cpu" || profileType == "trace" {
		if seconds > 0 {
			timeout = time.Duration(seconds+5) * time.Second
		}
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch pprof: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("pprof endpoint returned status %d", resp.StatusCode)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read profile: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return payload, contentType, nil
}

func (a *Analyzer) ClearDevMetrics() error {
	metricsURL := a.appURL + "/_debug/metrics/clear"
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, metricsURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to clear metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metrics clear returned status %d", resp.StatusCode)
	}

	return nil
}

func pprofPath(profileType string, seconds int) (string, string) {
	base := "/_debug/pprof"

	switch profileType {
	case "cpu":
		return base + "/profile", fmt.Sprintf("seconds=%d", seconds)
	case "trace":
		return base + "/trace", fmt.Sprintf("seconds=%d", seconds)
	default:
		return base + "/" + profileType, ""
	}
}

func previewHex(payload []byte, max int) string {
	if len(payload) == 0 || max <= 0 {
		return ""
	}
	if len(payload) < max {
		max = len(payload)
	}

	var b strings.Builder
	for i := 0; i < max; i++ {
		if i > 0 {
			if i%16 == 0 {
				b.WriteString(" ")
			}
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%02x", payload[i])
	}
	return b.String()
}
