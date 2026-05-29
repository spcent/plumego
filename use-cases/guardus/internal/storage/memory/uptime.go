package memory

import (
	"time"

	"guardus/internal/domain/endpoint"
)

const (
	uptimeCleanUpThreshold = 32 * 24
	uptimeRetention        = 30 * 24 * time.Hour
)

// processUptimeAfterResult updates hourly uptime counters and prunes entries
// older than uptimeRetention once we accumulate more than uptimeCleanUpThreshold.
func processUptimeAfterResult(uptime *endpoint.Uptime, result *endpoint.Result) {
	if uptime.HourlyStatistics == nil {
		uptime.HourlyStatistics = make(map[int64]*endpoint.HourlyUptimeStatistics)
	}
	unixTimestampFlooredAtHour := result.Timestamp.Truncate(time.Hour).Unix()
	hourlyStats := uptime.HourlyStatistics[unixTimestampFlooredAtHour]
	if hourlyStats == nil {
		hourlyStats = &endpoint.HourlyUptimeStatistics{}
		uptime.HourlyStatistics[unixTimestampFlooredAtHour] = hourlyStats
	}
	if result.Success {
		hourlyStats.SuccessfulExecutions++
	}
	hourlyStats.TotalExecutions++
	hourlyStats.TotalExecutionsResponseTime += uint64(result.Duration.Milliseconds())
	if len(uptime.HourlyStatistics) > uptimeCleanUpThreshold {
		threshold := time.Now().Add(-(uptimeRetention + time.Hour)).Unix()
		for ts := range uptime.HourlyStatistics {
			if threshold > ts {
				delete(uptime.HourlyStatistics, ts)
			}
		}
	}
}
