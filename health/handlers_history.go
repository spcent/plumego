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
		if manager == nil {
			sendErrorResponse(w, r, http.StatusServiceUnavailable, "HEALTH_MANAGER_UNAVAILABLE",
				"Health manager is not configured", "")
			return
		}

		history := manager.GetHealthHistory()

		_ = contract.WriteJSON(w, http.StatusOK, history)
	})
}

// HealthHistoryExportHandler creates a handler that exports health check history in various formats.
func HealthHistoryExportHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if manager == nil {
			sendErrorResponse(w, r, http.StatusServiceUnavailable, "HEALTH_MANAGER_UNAVAILABLE",
				"Health manager is not configured", "")
			return
		}

		// Parse query parameters for filtering and format
		query := HealthHistoryQuery{}

		// Parse time range
		if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
			if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				query.StartTime = &startTime
			}
		}

		if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
			if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
				query.EndTime = &endTime
			}
		}

		// Parse state filter
		if stateStr := r.URL.Query().Get("state"); stateStr != "" {
			state := HealthState(stateStr)
			query.State = &state
		}

		// Parse component filter
		if component := r.URL.Query().Get("component"); component != "" {
			query.Component = component
		}

		// Parse limit
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				query.Limit = limit
			}
		}

		// Parse offset
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil {
				query.Offset = offset
			}
		}

		// Get format parameter
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "json"
		}

		// Query history
		result := manager.QueryHealthHistory(query)

		switch strings.ToLower(format) {
		case "csv":
			exportHistoryToCSV(w, result.Entries)
		case "json":
			_ = contract.WriteJSON(w, http.StatusOK, result)
		default:
			sendErrorResponse(w, r, http.StatusBadRequest, "INVALID_FORMAT",
				"Supported formats: json, csv", "")
			return
		}
	})
}

// exportHistoryToCSV exports health history entries to CSV format.
func exportHistoryToCSV(w http.ResponseWriter, entries []HealthHistoryEntry) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=health_history.csv")

	writer := csv.NewWriter(w)

	// Write header
	header := []string{"Timestamp", "State", "Message", "Components", "Duration"}
	_ = writer.Write(header)

	// Write data
	for _, entry := range entries {
		record := []string{
			entry.Timestamp.Format(time.RFC3339),
			string(entry.State),
			entry.Message,
			strings.Join(entry.Components, ";"),
			entry.Duration.String(),
		}
		_ = writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		http.Error(w, "Failed to write CSV", http.StatusInternalServerError)
	}
}

// HealthHistoryStatsHandler returns statistics about health history.
func HealthHistoryStatsHandler(manager HealthManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if manager == nil {
			sendErrorResponse(w, r, http.StatusServiceUnavailable, "HEALTH_MANAGER_UNAVAILABLE",
				"Health manager is not configured", "")
			return
		}

		stats := manager.GetHealthHistoryStats()

		_ = contract.WriteJSON(w, http.StatusOK, stats)
	})
}
