package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spcent/plumego/store/db/rw"
	// Import your database driver here
	// For MySQL: _ "github.com/go-sql-driver/mysql"
	// For PostgreSQL: _ "github.com/lib/pq"
	// For SQLite: _ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("=== Plumego Read/Write Splitting Example ===")
	fmt.Println()

	// For this example, we'll use the same database for both primary and replicas
	// In production, these would be different database instances
	primaryDSN := "root:password@tcp(localhost:3306)/testdb?parseTime=true"

	// Open primary connection
	primary, err := sql.Open("mysql", primaryDSN)
	if err != nil {
		log.Fatalf("Failed to open primary: %v", err)
	}
	defer primary.Close()

	// Open replica connections (in this example, we use the same DSN)
	// In production, replicas would have different DSNs
	replica1, err := sql.Open("mysql", primaryDSN)
	if err != nil {
		log.Fatalf("Failed to open replica1: %v", err)
	}
	defer replica1.Close()

	replica2, err := sql.Open("mysql", primaryDSN)
	if err != nil {
		log.Fatalf("Failed to open replica2: %v", err)
	}
	defer replica2.Close()

	// Create read/write cluster with round-robin load balancing
	fmt.Println("1. Creating read/write cluster...")
	cluster, err := rw.New(rw.Config{
		Primary:  primary,
		Replicas: []*sql.DB{replica1, replica2},
		HealthCheck: rw.HealthCheckConfig{
			Enabled:           true,
			Interval:          30 * time.Second,
			Timeout:           5 * time.Second,
			FailureThreshold:  3,
			RecoveryThreshold: 2,
		},
		FallbackToPrimary: true,
	})
	if err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}
	defer cluster.Close()

	fmt.Println("✓ Cluster created with 1 primary + 2 replicas")
	fmt.Println()

	ctx := context.Background()

	// Example 1: Write operations (INSERT, UPDATE, DELETE) always go to primary
	fmt.Println("2. Executing WRITE operations (→ primary)...")
	_, err = cluster.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("Warning: CREATE TABLE failed: %v", err)
	}

	result, err := cluster.ExecContext(ctx,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"Alice", "alice@example.com",
	)
	if err != nil {
		log.Printf("Warning: INSERT failed: %v", err)
	} else {
		id, _ := result.LastInsertId()
		fmt.Printf("  ✓ Inserted user (id=%d)\n", id)
	}

	_, err = cluster.ExecContext(ctx,
		"UPDATE users SET name = ? WHERE email = ?",
		"Alice Smith", "alice@example.com",
	)
	if err != nil {
		log.Printf("Warning: UPDATE failed: %v", err)
	} else {
		fmt.Println("  ✓ Updated user")
	}

	// Example 2: Read operations (SELECT) go to replicas with load balancing
	fmt.Println("\n3. Executing READ operations (→ replicas, round-robin)...")
	for i := 0; i < 5; i++ {
		rows, err := cluster.QueryContext(ctx, "SELECT id, name, email FROM users")
		if err != nil {
			log.Printf("Warning: SELECT failed: %v", err)
			continue
		}

		count := 0
		for rows.Next() {
			var id int
			var name, email string
			if err := rows.Scan(&id, &name, &email); err != nil {
				log.Printf("Warning: Scan failed: %v", err)
				continue
			}
			count++
		}
		rows.Close()
		fmt.Printf("  ✓ Query %d: Read %d users (distributed to replica)\n", i+1, count)
	}

	// Example 3: Force primary for strong consistency
	fmt.Println("\n4. Forcing primary for strong consistency...")
	ctx = rw.ForcePrimary(context.Background())
	rows, err := cluster.QueryContext(ctx, "SELECT COUNT(*) as total FROM users")
	if err != nil {
		log.Printf("Warning: SELECT COUNT failed: %v", err)
	} else {
		if rows.Next() {
			var total int
			rows.Scan(&total)
			fmt.Printf("  ✓ Forced primary: Found %d users\n", total)
		}
		rows.Close()
	}

	// Example 4: Transaction (always uses primary)
	fmt.Println("\n5. Executing transaction (→ primary automatically)...")
	tx, err := cluster.BeginTx(context.Background(), nil)
	if err != nil {
		log.Printf("Warning: BeginTx failed: %v", err)
	} else {
		_, err = tx.ExecContext(context.Background(),
			"INSERT INTO users (name, email) VALUES (?, ?)",
			"Bob", "bob@example.com",
		)
		if err != nil {
			log.Printf("Warning: INSERT in tx failed: %v", err)
			tx.Rollback()
		} else {
			tx.Commit()
			fmt.Println("  ✓ Transaction committed")
		}
	}

	// Example 5: SELECT FOR UPDATE (uses primary)
	fmt.Println("\n6. Executing SELECT FOR UPDATE (→ primary automatically)...")
	rows, err = cluster.QueryContext(context.Background(),
		"SELECT id, name FROM users WHERE id = ? FOR UPDATE", 1)
	if err != nil {
		log.Printf("Warning: SELECT FOR UPDATE failed: %v", err)
	} else {
		if rows.Next() {
			var id int
			var name string
			rows.Scan(&id, &name)
			fmt.Printf("  ✓ Locked row: id=%d, name=%s\n", id, name)
		}
		rows.Close()
	}

	// Show cluster metrics
	fmt.Println("\n7. Cluster Metrics:")
	metrics := cluster.Metrics()
	fmt.Printf("  • Total queries:   %d\n", metrics.TotalQueries.Load())
	fmt.Printf("  • Primary queries: %d\n", metrics.PrimaryQueries.Load())
	fmt.Printf("  • Replica queries: %d\n", metrics.ReplicaQueries.Load())
	fmt.Printf("  • Fallback count:  %d\n", metrics.FallbackCount.Load())
	fmt.Printf("  • Health checks:   %d\n", metrics.HealthCheckCount.Load())

	// Show replica health
	fmt.Println("\n8. Replica Health Status:")
	health := cluster.ReplicaHealth()
	for i, isHealthy := range health {
		status := "✓ healthy"
		if !isHealthy {
			status = "✗ unhealthy"
		}
		fmt.Printf("  • Replica %d: %s\n", i, status)
	}

	// Show database stats
	fmt.Println("\n9. Connection Pool Stats:")
	stats := cluster.Stats()
	fmt.Printf("  Primary:  Open=%d, InUse=%d, Idle=%d\n",
		stats.Primary.OpenConnections,
		stats.Primary.InUse,
		stats.Primary.Idle,
	)
	for i, replicaStats := range stats.Replicas {
		fmt.Printf("  Replica %d: Open=%d, InUse=%d, Idle=%d\n",
			i,
			replicaStats.OpenConnections,
			replicaStats.InUse,
			replicaStats.Idle,
		)
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("  1. Automatic routing: writes → primary, reads → replicas")
	fmt.Println("  2. Load balancing: replicas selected using round-robin")
	fmt.Println("  3. Context control: ForcePrimary() for strong consistency")
	fmt.Println("  4. Transaction safety: all TX operations use primary")
	fmt.Println("  5. Health monitoring: automatic replica health checks")
	fmt.Println("  6. Metrics tracking: query counts and routing stats")
}
