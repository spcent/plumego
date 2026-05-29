package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"guardus/internal/storage"
)

// ResponseTimeHistory returns the hourly average-response-time history as
// JSON {timestamps[], values[]}.
//
// guardus v1 omits the SVG chart endpoint (/chart.svg) — the SPA renders
// the same data client-side from this JSON.
func ResponseTimeHistory(store storage.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duration := PathParam(r, "duration")
		from, ok := chartFromForDuration(duration)
		if !ok {
			WriteErrorString(w, r, http.StatusBadRequest, "Durations supported: 30d, 7d, 24h")
			return
		}
		key, err := PathKey(r)
		if err != nil {
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		hourly, err := store.GetHourlyAverageResponseTimeByKey(key, from, time.Now())
		if err != nil {
			writeStorageError(w, r, err)
			return
		}
		if len(hourly) == 0 {
			WriteRawJSON(w, http.StatusOK, []byte(`{"timestamps":[],"values":[]}`))
			return
		}
		hourlyTimestamps := make([]int, 0, len(hourly))
		earliest := int64(0)
		for ts := range hourly {
			hourlyTimestamps = append(hourlyTimestamps, int(ts))
			if earliest == 0 || ts < earliest {
				earliest = ts
			}
		}
		for earliest > from.Unix() {
			earliest -= int64(time.Hour.Seconds())
			hourlyTimestamps = append(hourlyTimestamps, int(earliest))
		}
		sort.Ints(hourlyTimestamps)
		timestamps := make([]int64, 0, len(hourlyTimestamps))
		values := make([]int, 0, len(hourlyTimestamps))
		for _, ts := range hourlyTimestamps {
			t64 := int64(ts)
			timestamps = append(timestamps, t64*1000)
			values = append(values, hourly[t64])
		}
		body, err := json.Marshal(map[string]any{
			"timestamps": timestamps,
			"values":     values,
		})
		if err != nil {
			WriteErrorString(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		WriteRawJSON(w, http.StatusOK, body)
	})
}

// ResponseTimeChart renders an SVG line chart of hourly response times.
//
// To avoid pulling in github.com/wcharczuk/go-chart's drawing dependencies,
// guardus emits a minimal hand-rolled SVG. The data and axis layout match
// upstream gatus closely enough for the SPA's needs.
func ResponseTimeChart(store storage.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duration := PathParam(r, "duration")
		from, ok := chartFromForDuration(duration)
		if !ok {
			WriteErrorString(w, r, http.StatusBadRequest, "Durations supported: 30d, 7d, 24h")
			return
		}
		key, err := PathKey(r)
		if err != nil {
			WriteErrorString(w, r, http.StatusBadRequest, "invalid key encoding")
			return
		}
		hourly, err := store.GetHourlyAverageResponseTimeByKey(key, from, time.Now())
		if err != nil {
			writeStorageError(w, r, err)
			return
		}
		if len(hourly) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		WriteSVG(w, http.StatusOK, renderResponseTimeSVG(hourly, from))
	})
}

func chartFromForDuration(d string) (time.Time, bool) {
	now := time.Now().Truncate(time.Hour)
	switch d {
	case "30d":
		return now.Add(-30 * 24 * time.Hour), true
	case "7d":
		return now.Add(-7 * 24 * time.Hour), true
	case "24h":
		return now.Add(-24 * time.Hour), true
	}
	return time.Time{}, false
}

func renderResponseTimeSVG(hourly map[int64]int, from time.Time) []byte {
	const width, height = 1280, 300
	const pad = 40
	keys := make([]int64, 0, len(hourly))
	earliest := int64(0)
	maxV := 0
	for k, v := range hourly {
		keys = append(keys, k)
		if earliest == 0 || k < earliest {
			earliest = k
		}
		if v > maxV {
			maxV = v
		}
	}
	for earliest > from.Unix() {
		earliest -= int64(time.Hour.Seconds())
		keys = append(keys, earliest)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	if maxV == 0 {
		maxV = 1
	}
	yMax := float64(maxV) * 1.25

	plotW := float64(width - 2*pad)
	plotH := float64(height - 2*pad)
	stepX := plotW / float64(max(1, len(keys)-1))

	var path strings.Builder
	for i, k := range keys {
		v := hourly[k]
		x := float64(pad) + float64(i)*stepX
		y := float64(pad) + plotH - (float64(v)/yMax)*plotH
		if i == 0 {
			fmt.Fprintf(&path, "M %.1f %.1f", x, y)
		} else {
			fmt.Fprintf(&path, " L %.1f %.1f", x, y)
		}
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">
  <rect width="%d" height="%d" fill="none"/>
  <path d="%s" stroke="#3a8" stroke-width="1.5" fill="none"/>
  <g font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11" fill="#777">
    <text x="%d" y="%d">Average response time per hour</text>
  </g>
</svg>`, width, height, width, height, path.String(), pad, height-10)
	return []byte(svg)
}
