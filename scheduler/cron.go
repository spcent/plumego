package scheduler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CronSpec represents a parsed cron expression (minute, hour, dom, month, dow).
type CronSpec struct {
	seconds []int
	minutes []int
	hours   []int
	dom     []int
	months  []int
	dow     []int
}

// ParseCronSpec parses a cron expression with 5 (minute) or 6 (second) fields.
// Fields: [second] minute hour day-of-month month day-of-week
func ParseCronSpec(expr string) (CronSpec, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 && len(fields) != 6 {
		return CronSpec{}, fmt.Errorf("invalid cron spec: expected 5 or 6 fields, got %d", len(fields))
	}

	var seconds []int
	offset := 0
	if len(fields) == 6 {
		parsed, err := parseCronField(fields[0], 0, 59)
		if err != nil {
			return CronSpec{}, fmt.Errorf("second: %w", err)
		}
		seconds = parsed
		offset = 1
	} else {
		seconds = []int{0}
	}

	minutes, err := parseCronField(fields[offset], 0, 59)
	if err != nil {
		return CronSpec{}, fmt.Errorf("minute: %w", err)
	}
	hours, err := parseCronField(fields[offset+1], 0, 23)
	if err != nil {
		return CronSpec{}, fmt.Errorf("hour: %w", err)
	}
	dom, err := parseCronField(fields[offset+2], 1, 31)
	if err != nil {
		return CronSpec{}, fmt.Errorf("day-of-month: %w", err)
	}
	months, err := parseCronField(fields[offset+3], 1, 12)
	if err != nil {
		return CronSpec{}, fmt.Errorf("month: %w", err)
	}
	dow, err := parseCronField(fields[offset+4], 0, 6)
	if err != nil {
		return CronSpec{}, fmt.Errorf("day-of-week: %w", err)
	}

	return CronSpec{
		seconds: seconds,
		minutes: minutes,
		hours:   hours,
		dom:     dom,
		months:  months,
		dow:     dow,
	}, nil
}

func parseCronField(field string, min, max int) ([]int, error) {
	if field == "*" {
		return buildRange(min, max, 1), nil
	}

	parts := strings.Split(field, ",")
	values := map[int]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty field")
		}

		step := 1
		if strings.Contains(part, "/") {
			s := strings.Split(part, "/")
			if len(s) != 2 {
				return nil, fmt.Errorf("invalid step syntax")
			}
			part = s[0]
			parsed, err := strconv.Atoi(s[1])
			if err != nil || parsed <= 0 {
				return nil, fmt.Errorf("invalid step")
			}
			step = parsed
		}

		var start, end int
		if part == "" || part == "*" {
			start = min
			end = max
		} else if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid range")
			}
			var err error
			start, err = strconv.Atoi(bounds[0])
			if err != nil {
				return nil, fmt.Errorf("invalid range start")
			}
			end, err = strconv.Atoi(bounds[1])
			if err != nil {
				return nil, fmt.Errorf("invalid range end")
			}
		} else {
			val, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid value")
			}
			start = val
			end = val
		}

		if start < min || end > max || start > end {
			return nil, fmt.Errorf("value out of bounds")
		}

		for v := start; v <= end; v += step {
			values[v] = struct{}{}
		}
	}

	result := make([]int, 0, len(values))
	for v := range values {
		result = append(result, v)
	}
	sort.Ints(result)
	return result, nil
}

func buildRange(min, max, step int) []int {
	result := make([]int, 0, ((max-min)/step)+1)
	for v := min; v <= max; v += step {
		result = append(result, v)
	}
	return result
}

// Next returns the next time after the given time that matches the spec.
func (c CronSpec) Next(after time.Time) time.Time {
	next := after.Truncate(time.Second).Add(time.Second)
	for i := 0; i < 366*24*60*60; i++ {
		if !containsInt(c.months, int(next.Month())) {
			next = time.Date(next.Year(), next.Month()+1, 1, 0, 0, 0, 0, next.Location())
			continue
		}
		if !containsInt(c.dom, next.Day()) {
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()).AddDate(0, 0, 1)
			continue
		}
		if !containsInt(c.dow, int(next.Weekday())) {
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location()).AddDate(0, 0, 1)
			continue
		}

		hour, carry := nextAllowed(next.Hour(), c.hours)
		if carry {
			next = time.Date(next.Year(), next.Month(), next.Day()+1, c.hours[0], 0, 0, 0, next.Location())
			continue
		}
		if hour != next.Hour() {
			next = time.Date(next.Year(), next.Month(), next.Day(), hour, 0, 0, 0, next.Location())
			continue
		}

		minute, carry := nextAllowed(next.Minute(), c.minutes)
		if carry {
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour()+1, c.minutes[0], 0, 0, next.Location())
			continue
		}
		if minute != next.Minute() {
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), minute, 0, 0, next.Location())
			continue
		}

		second, carry := nextAllowed(next.Second(), c.seconds)
		if carry {
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute()+1, c.seconds[0], 0, next.Location())
			continue
		}
		if second != next.Second() {
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute(), second, 0, next.Location())
			continue
		}

		return next
	}
	return time.Time{}
}

func containsInt(list []int, value int) bool {
	idx := sort.SearchInts(list, value)
	return idx < len(list) && list[idx] == value
}

func nextAllowed(current int, allowed []int) (int, bool) {
	if len(allowed) == 0 {
		return current, false
	}
	idx := sort.SearchInts(allowed, current)
	if idx < len(allowed) && allowed[idx] == current {
		return current, false
	}
	if idx < len(allowed) {
		return allowed[idx], false
	}
	return allowed[0], true
}
