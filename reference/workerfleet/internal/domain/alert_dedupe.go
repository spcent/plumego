package domain

import (
	"strconv"
	"strings"
	"time"
)

func dedupeKey(alertType AlertType, workerID WorkerID, taskID TaskID) string {
	switch {
	case workerID != "":
		return string(alertType) + ":" + string(workerID)
	case taskID != "":
		return string(alertType) + ":" + string(taskID)
	default:
		return string(alertType)
	}
}

func alertID(alertType AlertType, workerID WorkerID, taskID TaskID, now time.Time) string {
	return dedupeKey(alertType, workerID, taskID) + ":" + strconv.FormatInt(now.UnixNano(), 10)
}

func intString(value int) string {
	return strconv.Itoa(value)
}

func cloneAlertDetails(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func isOpenAlert(alert AlertRecord) bool {
	return alert.Status == AlertStatusFiring
}

func resolveAlert(alert AlertRecord, now time.Time) AlertRecord {
	return AlertRecord{
		AlertID:     alertID(alert.AlertType, alert.WorkerID, alert.TaskID, now),
		WorkerID:    alert.WorkerID,
		TaskID:      alert.TaskID,
		AlertType:   alert.AlertType,
		Status:      AlertStatusResolved,
		Severity:    alert.Severity,
		DedupeKey:   alert.DedupeKey,
		Message:     strings.TrimSpace(alert.Message + " resolved"),
		Details:     cloneAlertDetails(alert.Details),
		TriggeredAt: alert.TriggeredAt,
		ResolvedAt:  now,
	}
}
