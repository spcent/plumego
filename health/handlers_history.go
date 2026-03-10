package health

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
)

// HealthHistoryHandler creates a handler that returns health check history.
func HealthHistoryHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		_ = contract.WriteJSON(w, http.StatusOK, manager.GetHealthHistory())
	})
}

// HealthHistoryExportHandler creates a handler that exports health check history in various formats.
func HealthHistoryExportHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		query, err := parseHistoryQuery(r)
		if err != nil {
			contract.WriteError(w, r, contract.APIError{
				Status:   http.StatusBadRequest,
				Code:     "INVALID_QUERY",
				Message:  err.Error(),
				Category: contract.CategoryForStatus(http.StatusBadRequest),
			})
			return
		}

		format := strings.ToLower(r.URL.Query().Get("format"))
		if format == "" {
			format = "json"
		}

		result := manager.QueryHealthHistory(query)

		switch format {
		case "csv":
			writeHistoryCSV(w, result.Entries)
		case "json":
			_ = contract.WriteJSON(w, http.StatusOK, result)
		default:
			contract.WriteError(w, r, contract.APIError{
				Status:   http.StatusBadRequest,
				Code:     "INVALID_FORMAT",
				Message:  "supported formats: json, csv",
				Category: contract.CategoryForStatus(http.StatusBadRequest),
			})
		}
	})
}

// HealthHistoryStatsHandler returns statistics about health history.
func HealthHistoryStatsHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireManager(manager, w, r) {
			return
		}

		_ = contract.WriteJSON(w, http.StatusOK, manager.GetHealthHistoryStats())
	})
}

// parseHistoryQuery parses query parameters from the request into a HealthHistoryQuery.
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
		state := HealthState(s)
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

// invalidParamError is a simple error type for query parameter validation.
type invalidParamError struct {
	param string
	msg   string
}

func (e *invalidParamError) Error() string {
	return e.param + ": " + e.msg
}

// writeHistoryCSV exports health history entries to CSV format.
func writeHistoryCSV(w http.ResponseWriter, entries []HealthHistoryEntry) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=health_history.csv")

	writer := csv.NewWriter(w)

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
	if err := writer.Error(); err != nil {
		http.Error(w, "failed to write CSV", http.StatusInternalServerError)
	}
}
