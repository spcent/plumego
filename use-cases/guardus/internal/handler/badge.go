package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"guardus/internal/config"
	"guardus/internal/domain/endpoint/ui"
	"guardus/internal/storage"
	"guardus/internal/storage/common"
	"guardus/internal/storage/common/paging"
)

const (
	badgeColorHexAwesome  = "#40cc11"
	badgeColorHexGreat    = "#94cc11"
	badgeColorHexGood     = "#ccd311"
	badgeColorHexPassable = "#ccb311"
	badgeColorHexBad      = "#cc8111"
	badgeColorHexVeryBad  = "#c7130a"
)

const (
	HealthStatusUp      = "up"
	HealthStatusDown    = "down"
	HealthStatusUnknown = "?"
)

var badgeColors = []string{
	badgeColorHexAwesome,
	badgeColorHexGreat,
	badgeColorHexGood,
	badgeColorHexPassable,
	badgeColorHexBad,
}

// durationFromParam returns the lookback window for the given :duration
// path param, or false if the value is unknown. The "1h" case is offset by
// -2h because hourly metrics are written at hour boundaries.
func durationFromParam(d string) (time.Time, bool) {
	now := time.Now()
	switch d {
	case "30d":
		return now.Add(-30 * 24 * time.Hour), true
	case "7d":
		return now.Add(-7 * 24 * time.Hour), true
	case "24h":
		return now.Add(-24 * time.Hour), true
	case "1h":
		return now.Add(-2 * time.Hour), true
	}
	return time.Time{}, false
}

// UptimeBadge returns an SVG uptime badge for the configured duration.
func UptimeBadge(store storage.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duration := PathParam(r, "duration")
		from, ok := durationFromParam(duration)
		if !ok {
			WriteErrorString(w, r, http.StatusBadRequest, "Durations supported: 30d, 7d, 24h, 1h")
			return
		}
		key, err := PathKey(r)
		if err != nil {
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		uptime, err := store.GetUptimeByKey(key, from, time.Now())
		if err != nil {
			writeStorageError(w, r, err)
			return
		}
		WriteSVG(w, http.StatusOK, generateUptimeBadgeSVG(duration, uptime))
	})
}

// ResponseTimeBadge returns an SVG average-response-time badge.
func ResponseTimeBadge(cfg *config.Config, store storage.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duration := PathParam(r, "duration")
		from, ok := durationFromParam(duration)
		if !ok {
			WriteErrorString(w, r, http.StatusBadRequest, "Durations supported: 30d, 7d, 24h, 1h")
			return
		}
		key, err := PathKey(r)
		if err != nil {
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		avg, err := store.GetAverageResponseTimeByKey(key, from, time.Now())
		if err != nil {
			writeStorageError(w, r, err)
			return
		}
		WriteSVG(w, http.StatusOK, generateResponseTimeBadgeSVG(duration, avg, key, cfg))
	})
}

// HealthBadge returns an SVG up/down/? health badge for the latest result.
func HealthBadge(store storage.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, err := PathKey(r)
		if err != nil {
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		status, err := store.GetEndpointStatusByKey(key, paging.NewEndpointStatusParams().WithResults(1, 1))
		if err != nil {
			writeStorageError(w, r, err)
			return
		}
		health := HealthStatusUnknown
		if status != nil && len(status.Results) > 0 {
			if status.Results[0].Success {
				health = HealthStatusUp
			} else {
				health = HealthStatusDown
			}
		}
		WriteSVG(w, http.StatusOK, generateHealthBadgeSVG(health))
	})
}

// HealthBadgeShields returns the shields.io endpoint JSON for an endpoint's
// current health.
func HealthBadgeShields(store storage.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, err := PathKey(r)
		if err != nil {
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		status, err := store.GetEndpointStatusByKey(key, paging.NewEndpointStatusParams().WithResults(1, 1))
		if err != nil {
			writeStorageError(w, r, err)
			return
		}
		health := HealthStatusUnknown
		if status != nil && len(status.Results) > 0 {
			if status.Results[0].Success {
				health = HealthStatusUp
			} else {
				health = HealthStatusDown
			}
		}
		body, err := json.Marshal(map[string]any{
			"schemaVersion": 1,
			"label":         "guardus",
			"message":       health,
			"color":         shieldsColorFromHealth(health),
		})
		if err != nil {
			WriteErrorString(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Expires", "0")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
}

func writeStorageError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, common.ErrEndpointNotFound):
		WriteErrorString(w, r, http.StatusNotFound, err.Error())
	case errors.Is(err, common.ErrInvalidTimeRange):
		WriteErrorString(w, r, http.StatusBadRequest, err.Error())
	default:
		WriteErrorString(w, r, http.StatusInternalServerError, err.Error())
	}
}

func generateUptimeBadgeSVG(duration string, uptime float64) []byte {
	var labelWidth, valueWidthAdjustment int
	switch duration {
	case "30d", "24h":
		labelWidth = 70
	default:
		labelWidth = 65
	}
	color := uptimeBadgeColor(uptime)
	sanitized := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", uptime*100), "0"), ".") + "%"
	if strings.Contains(sanitized, ".") {
		valueWidthAdjustment = -10
	}
	valueWidth := (len(sanitized) * 11) + valueWidthAdjustment
	return composeBadgeSVG("uptime "+duration, sanitized, labelWidth, valueWidth, color)
}

func generateResponseTimeBadgeSVG(duration string, avg int, key string, cfg *config.Config) []byte {
	labelWidth := 110
	if duration == "7d" || duration == "1h" {
		labelWidth = 105
	}
	color := responseTimeBadgeColor(avg, key, cfg)
	sanitized := strconv.Itoa(avg) + "ms"
	valueWidth := len(sanitized) * 11
	return composeBadgeSVG("response time "+duration, sanitized, labelWidth, valueWidth, color)
}

func generateHealthBadgeSVG(health string) []byte {
	valueWidth := 10
	switch health {
	case HealthStatusUp:
		valueWidth = 28
	case HealthStatusDown:
		valueWidth = 44
	}
	return composeBadgeSVG("health", health, 48, valueWidth, healthBadgeColor(health))
}

func composeBadgeSVG(label, value string, labelWidth, valueWidth int, color string) []byte {
	width := labelWidth + valueWidth
	labelX := labelWidth / 2
	valueX := labelWidth + (valueWidth / 2)
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20">
  <linearGradient id="b" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a">
    <rect width="%d" height="20" rx="3" fill="#fff"/>
  </mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h%dv20H0z"/>
    <path fill="%s" d="M%d 0h%dv20H%dz"/>
    <path fill="url(#b)" d="M0 0h%dv20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>`, width, width, labelWidth, color, labelWidth, valueWidth, labelWidth, width,
		labelX, label, labelX, label,
		valueX, value, valueX, value,
	)
	return []byte(svg)
}

func uptimeBadgeColor(uptime float64) string {
	switch {
	case uptime >= 0.975:
		return badgeColorHexAwesome
	case uptime >= 0.95:
		return badgeColorHexGreat
	case uptime >= 0.9:
		return badgeColorHexGood
	case uptime >= 0.8:
		return badgeColorHexPassable
	case uptime >= 0.65:
		return badgeColorHexBad
	default:
		return badgeColorHexVeryBad
	}
}

func responseTimeBadgeColor(rt int, key string, cfg *config.Config) string {
	thresholds := ui.GetDefaultConfig().Badge.ResponseTime.Thresholds
	if ep := cfg.GetEndpointByKey(key); ep != nil && ep.UIConfig != nil && ep.UIConfig.Badge != nil && ep.UIConfig.Badge.ResponseTime != nil {
		thresholds = ep.UIConfig.Badge.ResponseTime.Thresholds
	}
	for i := range 5 {
		if rt <= thresholds[i] {
			return badgeColors[i]
		}
	}
	return badgeColorHexVeryBad
}

func healthBadgeColor(health string) string {
	switch health {
	case HealthStatusUp:
		return badgeColorHexAwesome
	case HealthStatusDown:
		return badgeColorHexVeryBad
	default:
		return badgeColorHexPassable
	}
}

func shieldsColorFromHealth(health string) string {
	switch health {
	case HealthStatusUp:
		return "brightgreen"
	case HealthStatusDown:
		return "red"
	default:
		return "yellow"
	}
}
