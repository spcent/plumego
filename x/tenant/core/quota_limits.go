package tenant

import "time"

// QuotaWindow defines the time window for quota limits.
type QuotaWindow string

const (
	QuotaWindowMinute QuotaWindow = "minute"
	QuotaWindowHour   QuotaWindow = "hour"
	QuotaWindowDay    QuotaWindow = "day"
	QuotaWindowMonth  QuotaWindow = "month"
)

// QuotaLimit defines quota limits for a specific window.
// Zero values mean unlimited for the dimension.
type QuotaLimit struct {
	Window   QuotaWindow
	Requests int
	Tokens   int
}

func normalizeQuotaLimits(cfg QuotaConfig) []QuotaLimit {
	if len(cfg.Limits) > 0 {
		limits := make([]QuotaLimit, 0, len(cfg.Limits))
		for _, limit := range cfg.Limits {
			if !isValidQuotaWindow(limit.Window) {
				continue
			}
			limits = append(limits, limit)
		}
		return limits
	}

	if cfg.RequestsPerMinute > 0 || cfg.TokensPerMinute > 0 {
		return []QuotaLimit{{
			Window:   QuotaWindowMinute,
			Requests: cfg.RequestsPerMinute,
			Tokens:   cfg.TokensPerMinute,
		}}
	}

	return nil
}

func isValidQuotaWindow(window QuotaWindow) bool {
	switch window {
	case QuotaWindowMinute, QuotaWindowHour, QuotaWindowDay, QuotaWindowMonth:
		return true
	default:
		return false
	}
}

func quotaWindowStart(now time.Time, window QuotaWindow) time.Time {
	now = now.UTC()
	switch window {
	case QuotaWindowHour:
		return now.Truncate(time.Hour)
	case QuotaWindowDay:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case QuotaWindowMonth:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	case QuotaWindowMinute:
		fallthrough
	default:
		return now.Truncate(time.Minute)
	}
}

func quotaWindowEnd(start time.Time, window QuotaWindow) time.Time {
	switch window {
	case QuotaWindowHour:
		return start.Add(time.Hour)
	case QuotaWindowDay:
		return start.Add(24 * time.Hour)
	case QuotaWindowMonth:
		return start.AddDate(0, 1, 0)
	case QuotaWindowMinute:
		fallthrough
	default:
		return start.Add(time.Minute)
	}
}
