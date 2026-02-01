package metrics

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBaseMetricsCollectorMaxRecordsDefault(t *testing.T) {
	collector := NewBaseMetricsCollector()
	total := defaultMaxRecords + 5

	for i := 0; i < total; i++ {
		collector.Record(context.Background(), MetricRecord{
			Type:  MetricHTTPRequest,
			Name:  "http_request",
			Value: float64(i),
		})
	}

	records := collector.GetRecords()
	if len(records) != defaultMaxRecords {
		t.Fatalf("expected %d records, got %d", defaultMaxRecords, len(records))
	}

	if records[0].Value != 5 {
		t.Fatalf("expected oldest record to be 5, got %v", records[0].Value)
	}
	if records[len(records)-1].Value != float64(total-1) {
		t.Fatalf("expected newest record to be %d, got %v", total-1, records[len(records)-1].Value)
	}
}

func TestBaseMetricsCollectorMaxRecordsDisabled(t *testing.T) {
	collector := NewBaseMetricsCollector().WithMaxRecords(0)
	total := 250

	for i := 0; i < total; i++ {
		collector.Record(context.Background(), MetricRecord{
			Type:  MetricHTTPRequest,
			Name:  "http_request",
			Value: float64(i),
		})
	}

	records := collector.GetRecords()
	if len(records) != total {
		t.Fatalf("expected %d records, got %d", total, len(records))
	}
}

func TestBaseMetricsCollector_ObserveDB(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		driver    string
		query     string
		rows      int
		duration  time.Duration
		err       error
		wantType  MetricType
		wantName  string
	}{
		{
			name:      "query operation",
			operation: "query",
			driver:    "postgres",
			query:     "SELECT * FROM users",
			rows:      10,
			duration:  50 * time.Millisecond,
			err:       nil,
			wantType:  MetricDBQuery,
			wantName:  "db_query",
		},
		{
			name:      "exec operation",
			operation: "exec",
			driver:    "mysql",
			query:     "INSERT INTO users (name) VALUES (?)",
			rows:      1,
			duration:  30 * time.Millisecond,
			err:       nil,
			wantType:  MetricDBExec,
			wantName:  "db_exec",
		},
		{
			name:      "transaction operation",
			operation: "transaction",
			driver:    "sqlite3",
			query:     "BEGIN",
			rows:      0,
			duration:  5 * time.Millisecond,
			err:       nil,
			wantType:  MetricDBTransaction,
			wantName:  "db_transaction",
		},
		{
			name:      "ping operation",
			operation: "ping",
			driver:    "postgres",
			query:     "",
			rows:      0,
			duration:  2 * time.Millisecond,
			err:       nil,
			wantType:  MetricDBPing,
			wantName:  "db_ping",
		},
		{
			name:      "connect operation",
			operation: "connect",
			driver:    "postgres",
			query:     "",
			rows:      0,
			duration:  100 * time.Millisecond,
			err:       nil,
			wantType:  MetricDBConnect,
			wantName:  "db_connect",
		},
		{
			name:      "close operation",
			operation: "close",
			driver:    "postgres",
			query:     "",
			rows:      0,
			duration:  10 * time.Millisecond,
			err:       nil,
			wantType:  MetricDBClose,
			wantName:  "db_close",
		},
		{
			name:      "query with error",
			operation: "query",
			driver:    "postgres",
			query:     "SELECT * FROM invalid_table",
			rows:      0,
			duration:  20 * time.Millisecond,
			err:       errors.New("table does not exist"),
			wantType:  MetricDBQuery,
			wantName:  "db_query",
		},
		{
			name:      "long query truncation",
			operation: "query",
			driver:    "postgres",
			query:     "SELECT * FROM users WHERE name = 'very long name that should be truncated because it exceeds the maximum query length limit'",
			rows:      1,
			duration:  40 * time.Millisecond,
			err:       nil,
			wantType:  MetricDBQuery,
			wantName:  "db_query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewBaseMetricsCollector()
			ctx := context.Background()

			collector.ObserveDB(ctx, tt.operation, tt.driver, tt.query, tt.rows, tt.duration, tt.err)

			records := collector.GetRecords()
			if len(records) != 1 {
				t.Fatalf("expected 1 record, got %d", len(records))
			}

			record := records[0]

			// Verify record type
			if record.Type != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, record.Type)
			}

			// Verify record name
			if record.Name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, record.Name)
			}

			// Verify duration
			if record.Duration != tt.duration {
				t.Errorf("expected duration %v, got %v", tt.duration, record.Duration)
			}

			// Verify value (should be milliseconds)
			expectedValue := float64(tt.duration.Milliseconds())
			if record.Value != expectedValue {
				t.Errorf("expected value %f, got %f", expectedValue, record.Value)
			}

			// Verify error
			if (record.Error != nil) != (tt.err != nil) {
				t.Errorf("expected error %v, got %v", tt.err, record.Error)
			}

			// Verify labels
			if record.Labels == nil {
				t.Fatal("expected labels to be set")
			}

			if record.Labels[labelOperation] != tt.operation {
				t.Errorf("expected operation label %q, got %q", tt.operation, record.Labels[labelOperation])
			}

			if tt.driver != "" && record.Labels[labelDriver] != tt.driver {
				t.Errorf("expected driver label %q, got %q", tt.driver, record.Labels[labelDriver])
			}

			if tt.query != "" {
				queryLabel := record.Labels[labelQuery]
				if len(tt.query) > 100 {
					// Should be truncated
					if len(queryLabel) > 104 { // 100 + "..."
						t.Errorf("expected query to be truncated, got length %d", len(queryLabel))
					}
					if queryLabel[len(queryLabel)-3:] != "..." {
						t.Errorf("expected truncated query to end with '...', got %q", queryLabel)
					}
				} else {
					if queryLabel != tt.query {
						t.Errorf("expected query label %q, got %q", tt.query, queryLabel)
					}
				}
			}

			if tt.rows > 0 {
				rowsLabel := record.Labels[labelRows]
				expectedRows := "10"
				if tt.rows == 1 {
					expectedRows = "1"
				}
				if rowsLabel != expectedRows && tt.rows != 0 {
					// Only check if rows were actually set
					t.Errorf("expected rows label %q, got %q", expectedRows, rowsLabel)
				}
			}

			// Verify stats
			stats := collector.GetStats()
			if stats.TotalRecords != 1 {
				t.Errorf("expected 1 total record, got %d", stats.TotalRecords)
			}

			if tt.err != nil && stats.ErrorRecords != 1 {
				t.Errorf("expected 1 error record, got %d", stats.ErrorRecords)
			}

			if stats.TypeBreakdown[tt.wantType] != 1 {
				t.Errorf("expected 1 record of type %q, got %d", tt.wantType, stats.TypeBreakdown[tt.wantType])
			}
		})
	}
}

func TestBaseMetricsCollector_ObserveDB_DefaultOperation(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	// Test with unknown operation - should default to MetricDBQuery
	collector.ObserveDB(ctx, "unknown_operation", "postgres", "SELECT 1", 0, 10*time.Millisecond, nil)

	records := collector.GetRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.Type != MetricDBQuery {
		t.Errorf("expected default type MetricDBQuery, got %q", record.Type)
	}
}

func TestBaseMetricsCollector_ObserveDB_Multiple(t *testing.T) {
	collector := NewBaseMetricsCollector()
	ctx := context.Background()

	// Record multiple DB operations
	collector.ObserveDB(ctx, "query", "postgres", "SELECT * FROM users", 10, 50*time.Millisecond, nil)
	collector.ObserveDB(ctx, "exec", "postgres", "INSERT INTO users (name) VALUES (?)", 1, 30*time.Millisecond, nil)
	collector.ObserveDB(ctx, "ping", "postgres", "", 0, 5*time.Millisecond, nil)

	records := collector.GetRecords()
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	stats := collector.GetStats()
	if stats.TotalRecords != 3 {
		t.Errorf("expected 3 total records, got %d", stats.TotalRecords)
	}

	if stats.TypeBreakdown[MetricDBQuery] != 1 {
		t.Errorf("expected 1 query record, got %d", stats.TypeBreakdown[MetricDBQuery])
	}

	if stats.TypeBreakdown[MetricDBExec] != 1 {
		t.Errorf("expected 1 exec record, got %d", stats.TypeBreakdown[MetricDBExec])
	}

	if stats.TypeBreakdown[MetricDBPing] != 1 {
		t.Errorf("expected 1 ping record, got %d", stats.TypeBreakdown[MetricDBPing])
	}
}
