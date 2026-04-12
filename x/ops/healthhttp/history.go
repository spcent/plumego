package healthhttp

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/health"
)

// HealthHistoryHandler creates a handler that returns tracked health history.
func HealthHistoryHandler(tracker *Tracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireTracker(tracker, w, r) {
			return
		}

		_ = contract.WriteJSON(w, http.StatusOK, tracker.GetHealthHistory())
	})
}

// HealthHistoryExportHandler creates a handler that exports tracked health history in various formats.
func HealthHistoryExportHandler(tracker *Tracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireTracker(tracker, w, r) {
			return
		}

		query, err := parseHistoryQuery(r)
		if err != nil {
			contract.WriteError(w, r, contract.NewErrorBuilder().
				Status(http.StatusBadRequest).
				Code("INVALID_QUERY").
				Message(err.Error()).
				Build())
			return
		}

		format := strings.ToLower(r.URL.Query().Get("format"))
		if format == "" {
			format = "json"
		}

		result := tracker.QueryHealthHistory(query)

		switch format {
		case "csv":
			writeHistoryCSV(w, result.Entries)
		case "json":
			_ = contract.WriteJSON(w, http.StatusOK, result)
		default:
			contract.WriteError(w, r, contract.NewErrorBuilder().
				Status(http.StatusBadRequest).
				Code("INVALID_FORMAT").
				Message("supported formats: json, csv").
				Build())
		}
	})
}

// HealthHistoryStatsHandler returns statistics about tracked health history.
func HealthHistoryStatsHandler(tracker *Tracker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireTracker(tracker, w, r) {
			return
		}

		_ = contract.WriteJSON(w, http.StatusOK, tracker.GetHealthHistoryStats())
	})
}

func parseHistoryQuery(r *http.Request) (HealthHistoryQuery, error) {
	q := HealthHistoryQuery{}
	params := r.URL.Query()

	if s := params.Get("start_time"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			q.StartTime = &t
		}
	}

	if s := params.Get("end_time"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			q.EndTime = &t
		}
	}

	if s := params.Get("state"); s != "" {
		state := health.HealthState(s)
		if !isValidHealthState(state) {
			return q, &invalidParamError{param: "state", msg: "valid states: healthy, degraded, unhealthy"}
		}
		q.State = &state
	}

	if s := params.Get("component"); s != "" {
		q.Component = s
	}

	if s := params.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			q.Limit = n
		}
	}

	if s := params.Get("offset"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			q.Offset = n
		}
	}

	return q, nil
}

type invalidParamError struct {
	param string
	msg   string
}

func (e *invalidParamError) Error() string {
	return e.param + ": " + e.msg
}

func writeHistoryCSV(w http.ResponseWriter, entries []HealthHistoryEntry) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"Timestamp", "State", "Message", "Components", "Duration"})

	for _, entry := range entries {
		_ = writer.Write([]string{
			entry.Timestamp.Format(time.RFC3339),
			string(entry.State),
			entry.Message,
			strings.Join(entry.Components, ";"),
			entry.Duration.String(),
		})
	}

	writer.Flush()
	if writer.Error() != nil {
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=health_history.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

func isValidHealthState(state health.HealthState) bool {
	switch state {
	case health.StatusHealthy, health.StatusDegraded, health.StatusUnhealthy:
		return true
	default:
		return false
	}
}
