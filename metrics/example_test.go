package metrics_test

import (
	"context"
	"fmt"
	"time"

	"github.com/spcent/plumego/metrics"
)

func ExampleBaseMetricsCollector() {
	collector := metrics.NewBaseMetricsCollector()
	collector.ObserveHTTP(context.Background(), "GET", "/users", 200, 128, 25*time.Millisecond)

	stats := collector.GetStats()
	fmt.Printf("records=%d errors=%d\n", stats.TotalRecords, stats.ErrorRecords)

	// Output:
	// records=1 errors=0
}

func ExampleNewMultiCollector() {
	left := metrics.NewBaseMetricsCollector()
	right := metrics.NewBaseMetricsCollector()
	collector := metrics.NewMultiCollector(left, right)

	collector.ObserveHTTP(context.Background(), "POST", "/batch", 202, 64, 10*time.Millisecond)

	fmt.Printf("left=%d right=%d\n", len(left.GetRecords()), len(right.GetRecords()))

	// Output:
	// left=1 right=1
}
