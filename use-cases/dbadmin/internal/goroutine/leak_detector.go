package goroutine

import (
	"runtime"
	"sort"
	"strings"
	"time"
)

// Snapshot captures goroutine state at a point in time.
type Snapshot struct {
	Count      int
	Stacks     []string
	Timestamp  time.Time
}

// TakeSnapshot captures current goroutine stacks.
func TakeSnapshot() Snapshot {
	buf := make([]byte, 1<<20) // 1MB buffer
	n := runtime.Stack(buf, true)
	stacks := string(buf[:n])

	return Snapshot{
		Count:     runtime.NumGoroutine(),
		Stacks:    strings.Split(stacks, "\n\n"),
		Timestamp: time.Now(),
	}
}

// Diff compares two snapshots and returns stacks that appeared in 'after' but not in 'before'.
func Diff(before, after Snapshot) []string {
	beforeMap := make(map[string]bool)
	for _, stack := range before.Stacks {
		normalized := normalizeStack(stack)
		if normalized != "" {
			beforeMap[normalized] = true
		}
	}

	var leaked []string
	for _, stack := range after.Stacks {
		normalized := normalizeStack(stack)
		if normalized != "" && !beforeMap[normalized] {
			leaked = append(leaked, stack)
		}
	}

	return leaked
}

// normalizeStack removes variable parts (addresses, IDs) for comparison.
func normalizeStack(stack string) string {
	lines := strings.Split(stack, "\n")
	if len(lines) < 2 {
		return ""
	}
	// Skip goroutine header line
	return strings.Join(lines[1:], "\n")
}

// LeakDetector monitors for goroutine leaks over time.
type LeakDetector struct {
	baseline   Snapshot
	threshold  int
	checkInterval time.Duration
}

// NewLeakDetector creates a detector with baseline snapshot.
func NewLeakDetector(threshold int, checkInterval time.Duration) *LeakDetector {
	return &LeakDetector{
		baseline:      TakeSnapshot(),
		threshold:     threshold,
		checkInterval: checkInterval,
	}
}

// Check returns leaked goroutines if count exceeds threshold above baseline.
func (d *LeakDetector) Check() []string {
	current := TakeSnapshot()
	growth := current.Count - d.baseline.Count

	if growth <= d.threshold {
		return nil
	}

	return Diff(d.baseline, current)
}

// SummarizeStacks groups similar stacks and returns counts.
func SummarizeStacks(stacks []string) map[string]int {
	summary := make(map[string]int)
	for _, stack := range stacks {
		// Extract function name from first line
		lines := strings.Split(stack, "\n")
		if len(lines) > 0 {
			key := extractFunctionName(lines[0])
			summary[key]++
		}
	}
	return summary
}

// extractFunctionName parses goroutine header to get function name.
func extractFunctionName(header string) string {
	// Example: "goroutine 123 [running]:"
	// or "goroutine 456 [chan receive, 10 minutes]:"
	parts := strings.Split(header, " ")
	if len(parts) >= 3 {
		return parts[2] // Function name or state
	}
	return header
}

// TopLeakedFunctions returns the most common leaked functions.
func TopLeakedFunctions(stacks []string, limit int) []struct {
	Function string
	Count    int
} {
	summary := SummarizeStacks(stacks)

	type entry struct {
		Function string
		Count    int
	}

	var entries []entry
	for fn, count := range summary {
		entries = append(entries, entry{fn, count})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	var result []struct {
		Function string
		Count    int
	}
	for _, e := range entries {
		result = append(result, struct {
			Function string
			Count    int
		}{e.Function, e.Count})
	}

	return result
}
