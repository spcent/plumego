package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/store/db"
)

func main() {
	// Create a Prometheus metrics collector
	collector := metrics.NewPrometheusCollector("dbmetrics_example")

	// Open SQLite database (in production, use PostgreSQL/MySQL)
	rawDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer rawDB.Close()

	// Configure connection pool settings
	rawDB.SetMaxOpenConns(10)
	rawDB.SetMaxIdleConns(5)
	rawDB.SetConnMaxLifetime(5 * time.Minute)
	rawDB.SetConnMaxIdleTime(2 * time.Minute)

	// Wrap database with metrics instrumentation
	instrumentedDB := db.NewInstrumentedDB(rawDB, collector, "sqlite3")

	// Create example schema
	ctx := context.Background()
	if err := setupSchema(ctx, instrumentedDB); err != nil {
		log.Fatalf("Failed to setup schema: %v", err)
	}

	// Start periodic connection pool monitoring
	go monitorConnectionPool(instrumentedDB)

	// Simulate database operations
	go simulateWorkload(instrumentedDB)

	// Expose metrics endpoint
	http.Handle("/metrics", collector.Handler())
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/stats", handleStats(instrumentedDB))

	addr := ":8080"
	fmt.Printf("Server started on %s\n", addr)
	fmt.Printf("Metrics: http://localhost%s/metrics\n", addr)
	fmt.Printf("Stats:   http://localhost%s/stats\n", addr)
	fmt.Printf("\nSample Prometheus queries:\n")
	fmt.Printf("  rate(db_query_total[1m])                  - Queries per second\n")
	fmt.Printf("  db_pool_in_use / db_pool_open_connections - Pool utilization\n")
	fmt.Printf("  histogram_quantile(0.95, rate(db_query_duration_bucket[5m])) - p95 latency\n")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func setupSchema(ctx context.Context, db db.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id)`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("schema creation failed: %w", err)
		}
	}

	return nil
}

func monitorConnectionPool(db *db.InstrumentedDB) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		db.RecordPoolStats(ctx)
	}
}

func simulateWorkload(db db.DB) {
	ctx := context.Background()
	userID := 0

	for {
		// Insert a new user every 2-5 seconds
		time.Sleep(time.Duration(2+time.Now().Unix()%3) * time.Second)

		userID++
		name := fmt.Sprintf("User%d", userID)
		email := fmt.Sprintf("user%d@example.com", userID)

		result, err := db.ExecContext(ctx,
			"INSERT INTO users (name, email) VALUES (?, ?)",
			name, email)
		if err != nil {
			log.Printf("Insert user failed: %v", err)
			continue
		}

		id, _ := result.LastInsertId()

		// Insert 2-3 posts for this user
		postCount := 2 + int(time.Now().Unix()%2)
		for i := 0; i < postCount; i++ {
			title := fmt.Sprintf("Post %d from %s", i+1, name)
			content := fmt.Sprintf("This is post content from %s at %s", name, time.Now().Format(time.RFC3339))

			_, err := db.ExecContext(ctx,
				"INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)",
				id, title, content)
			if err != nil {
				log.Printf("Insert post failed: %v", err)
			}
		}

		// Perform some queries
		// Query user by email
		row := db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE email = ?", email)
		var foundID int
		var foundName string
		if err := row.Scan(&foundID, &foundName); err != nil {
			log.Printf("Query user failed: %v", err)
		}

		// Query user's posts
		rows, err := db.QueryContext(ctx,
			"SELECT id, title FROM posts WHERE user_id = ? ORDER BY created_at DESC",
			id)
		if err != nil {
			log.Printf("Query posts failed: %v", err)
		} else {
			for rows.Next() {
				var postID int
				var title string
				rows.Scan(&postID, &title)
			}
			rows.Close()
		}

		// Occasionally run an aggregate query
		if userID%5 == 0 {
			var totalPosts int
			err := db.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM posts").Scan(&totalPosts)
			if err != nil {
				log.Printf("Aggregate query failed: %v", err)
			}
		}

		// Simulate a transaction
		if userID%10 == 0 {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				log.Printf("Begin transaction failed: %v", err)
				continue
			}

			_, err = tx.ExecContext(ctx,
				"UPDATE users SET name = ? WHERE id = ?",
				fmt.Sprintf("Updated%d", id), id)
			if err != nil {
				tx.Rollback()
				log.Printf("Transaction update failed: %v", err)
				continue
			}

			if err := tx.Commit(); err != nil {
				log.Printf("Transaction commit failed: %v", err)
			}
		}
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Database Metrics Example</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .links { margin: 20px 0; }
        .links a { display: block; margin: 10px 0; padding: 10px; background: #f0f0f0; text-decoration: none; color: #333; border-radius: 4px; }
        .links a:hover { background: #e0e0e0; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>Database Metrics Example</h1>
    <p>This example demonstrates database metrics tracking with plumego.</p>

    <div class="links">
        <a href="/metrics">📊 Prometheus Metrics</a>
        <a href="/stats">📈 Connection Pool Stats</a>
    </div>

    <h2>Features Demonstrated</h2>
    <ul>
        <li>Automatic query tracking (INSERT, SELECT, UPDATE)</li>
        <li>Transaction monitoring</li>
        <li>Connection pool health metrics</li>
        <li>Query duration tracking</li>
        <li>Error rate monitoring</li>
    </ul>

    <h2>Simulated Workload</h2>
    <p>The application is continuously:</p>
    <ul>
        <li>Creating users every 2-5 seconds</li>
        <li>Creating posts for each user</li>
        <li>Running SELECT queries</li>
        <li>Running periodic transactions</li>
        <li>Recording pool stats every 10 seconds</li>
    </ul>

    <h2>Sample Prometheus Queries</h2>
    <pre><code>
# Queries per second by operation
rate(db_query_total[1m])

# Average query duration
rate(db_query_duration_milliseconds_sum[5m]) / rate(db_query_duration_milliseconds_count[5m])

# Error rate
rate(db_query_errors_total[5m]) / rate(db_query_total[5m])

# Connection pool utilization
db_pool_in_use / db_pool_open_connections

# Connection wait time
rate(db_pool_wait_duration_sum[5m])
    </code></pre>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleStats(instrumentedDB *db.InstrumentedDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := instrumentedDB.Stats()

		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Database Stats</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        table { border-collapse: collapse; width: 100%%; max-width: 600px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:hover { background-color: #f5f5f5; }
        .good { color: green; }
        .warning { color: orange; }
        .bad { color: red; }
    </style>
    <meta http-equiv="refresh" content="5">
</head>
<body>
    <h1>Connection Pool Statistics</h1>
    <p>Auto-refreshes every 5 seconds</p>

    <table>
        <tr><th>Metric</th><th>Value</th></tr>
        <tr><td>Max Open Connections</td><td>%d</td></tr>
        <tr><td>Open Connections</td><td class="%s">%d</td></tr>
        <tr><td>In Use</td><td>%d</td></tr>
        <tr><td>Idle</td><td>%d</td></tr>
        <tr><td>Wait Count</td><td class="%s">%d</td></tr>
        <tr><td>Wait Duration</td><td>%s</td></tr>
        <tr><td>Max Idle Closed</td><td>%d</td></tr>
        <tr><td>Max Lifetime Closed</td><td>%d</td></tr>
        <tr><td>Max Idle Time Closed</td><td>%d</td></tr>
    </table>

    <p><a href="/">← Back</a></p>
</body>
</html>
`,
			stats.MaxOpenConnections,
			getUtilizationClass(stats.OpenConnections, stats.MaxOpenConnections),
			stats.OpenConnections,
			stats.InUse,
			stats.Idle,
			getWaitClass(stats.WaitCount),
			stats.WaitCount,
			stats.WaitDuration,
			stats.MaxIdleClosed,
			stats.MaxLifetimeClosed,
			stats.MaxIdleTimeClosed,
		)

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}

func getUtilizationClass(current, max int) string {
	if max == 0 {
		return ""
	}
	utilization := float64(current) / float64(max)
	if utilization < 0.7 {
		return "good"
	} else if utilization < 0.9 {
		return "warning"
	}
	return "bad"
}

func getWaitClass(waitCount int64) string {
	if waitCount == 0 {
		return "good"
	} else if waitCount < 100 {
		return "warning"
	}
	return "bad"
}
