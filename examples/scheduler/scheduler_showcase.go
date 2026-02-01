package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spcent/plumego/scheduler"
)

// This example showcases all the enhanced scheduler features:
// 1. Backpressure strategies
// 2. Batch operations (pause/resume/cancel by group/tags)
// 3. Job queries with filtering and pagination
// 4. Job dependencies (DAG)
// 5. Cron expression enhancements (@descriptors, @every, timezones)

func main() {
	// ========================================
	// Feature 1: Backpressure Strategy
	// ========================================
	fmt.Println("=== Feature 1: Backpressure Strategy ===")

	// Configure backpressure to block with timeout instead of dropping
	backpressureConfig := scheduler.BackpressureConfig{
		Policy:  scheduler.BackpressureBlockTimeout,
		Timeout: 100 * time.Millisecond,
		OnBackpressure: func(jobID scheduler.JobID) {
			log.Printf("BACKPRESSURE: Job %s dropped due to queue full\n", jobID)
		},
	}

	sched := scheduler.New(
		scheduler.WithWorkers(2),
		scheduler.WithQueueSize(5),
		scheduler.WithBackpressure(backpressureConfig),
	)
	sched.Start()
	defer sched.Stop(context.Background())

	// ========================================
	// Feature 2: Cron Expression Enhancements
	// ========================================
	fmt.Println("\n=== Feature 2: Cron Expression Enhancements ===")

	// 2.1: Use @descriptors for common schedules
	sched.AddCron("hourly-task", "@hourly", func(ctx context.Context) error {
		fmt.Println("Hourly task executed")
		return nil
	}, scheduler.WithGroup("scheduled"), scheduler.WithTags("periodic"))

	sched.AddCron("daily-cleanup", "@daily", func(ctx context.Context) error {
		fmt.Println("Daily cleanup executed")
		return nil
	}, scheduler.WithGroup("maintenance"))

	// 2.2: Use @every for interval-based scheduling
	sched.AddCron("health-check", "@every 30s", func(ctx context.Context) error {
		fmt.Println("Health check executed")
		return nil
	}, scheduler.WithGroup("monitoring"), scheduler.WithTags("health"))

	sched.AddCron("metrics-collector", "@every 5m", func(ctx context.Context) error {
		fmt.Println("Collecting metrics...")
		return nil
	}, scheduler.WithGroup("monitoring"), scheduler.WithTags("metrics", "periodic"))

	// 2.3: Use timezone-aware scheduling
	tokyo := time.FixedZone("JST", 9*3600)
	sched.AddCronWithLocation("tokyo-report", "0 9 * * *", func(ctx context.Context) error {
		fmt.Println("Tokyo morning report (9 AM JST)")
		return nil
	}, tokyo, scheduler.WithGroup("reports"), scheduler.WithTags("international"))

	newYork, _ := time.LoadLocation("America/New_York")
	sched.AddCronWithLocation("ny-market-open", "30 9 * * 1-5", func(ctx context.Context) error {
		fmt.Println("New York market opened (9:30 AM EST)")
		return nil
	}, newYork, scheduler.WithGroup("trading"))

	// ========================================
	// Feature 3: Job Dependencies (DAG)
	// ========================================
	fmt.Println("\n=== Feature 3: Job Dependencies ===")

	// Create a data pipeline with dependencies
	// Pipeline: fetch-data -> process-data -> generate-report

	sched.Delay("fetch-data", 100*time.Millisecond, func(ctx context.Context) error {
		fmt.Println("Step 1: Fetching data...")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Step 1: Data fetched successfully")
		return nil
	}, scheduler.WithGroup("pipeline"), scheduler.WithTags("data"))

	sched.Delay("process-data", 100*time.Millisecond, func(ctx context.Context) error {
		fmt.Println("Step 2: Processing data...")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Step 2: Data processed successfully")
		return nil
	},
		scheduler.WithGroup("pipeline"),
		scheduler.WithTags("data"),
		scheduler.WithDependsOn(scheduler.DependencyFailureCancel, "fetch-data"),
	)

	sched.Delay("generate-report", 100*time.Millisecond, func(ctx context.Context) error {
		fmt.Println("Step 3: Generating report...")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Step 3: Report generated successfully")
		return nil
	},
		scheduler.WithGroup("pipeline"),
		scheduler.WithTags("reporting"),
		scheduler.WithDependsOn(scheduler.DependencyFailureCancel, "fetch-data", "process-data"),
	)

	// Wait for pipeline to complete
	time.Sleep(500 * time.Millisecond)

	// ========================================
	// Feature 4: Batch Operations
	// ========================================
	fmt.Println("\n=== Feature 4: Batch Operations ===")

	// Pause all monitoring jobs
	paused := sched.PauseByGroup("monitoring")
	fmt.Printf("Paused %d monitoring jobs\n", paused)

	// Query to verify
	result := sched.QueryJobs(scheduler.JobQuery{
		Group:  "monitoring",
		Paused: boolPtr(true),
	})
	fmt.Printf("Confirmed: %d monitoring jobs are paused\n", result.Total)

	// Resume jobs with specific tags
	resumed := sched.ResumeByTags("health")
	fmt.Printf("Resumed %d health-check jobs\n", resumed)

	// Cancel entire pipeline group
	canceled := sched.CancelByGroup("pipeline")
	fmt.Printf("Canceled %d pipeline jobs\n", canceled)

	// ========================================
	// Feature 5: Advanced Job Queries
	// ========================================
	fmt.Println("\n=== Feature 5: Advanced Job Queries ===")

	// Query 1: Find all periodic tasks
	periodicJobs := sched.QueryJobs(scheduler.JobQuery{
		Tags:      []string{"periodic"},
		OrderBy:   "id",
		Ascending: true,
	})
	fmt.Printf("\nPeriodic jobs (%d total):\n", periodicJobs.Total)
	for _, job := range periodicJobs.Jobs {
		fmt.Printf("  - %s (Group: %s, Tags: %v)\n", job.ID, job.Group, job.Tags)
	}

	// Query 2: Find running jobs with pagination
	running := true
	runningJobs := sched.QueryJobs(scheduler.JobQuery{
		Running:   &running,
		Limit:     5,
		Offset:    0,
		OrderBy:   "next_run",
		Ascending: true,
	})
	fmt.Printf("\nRunning jobs (showing %d of %d):\n", len(runningJobs.Jobs), runningJobs.Total)
	for _, job := range runningJobs.Jobs {
		fmt.Printf("  - %s (Next run: %v)\n", job.ID, job.NextRun)
	}

	// Query 3: Complex filter - monitoring group, not paused, with metrics tag
	notPaused := false
	complexQuery := sched.QueryJobs(scheduler.JobQuery{
		Group:     "monitoring",
		Tags:      []string{"metrics"},
		Paused:    &notPaused,
		OrderBy:   "id",
		Ascending: true,
	})
	fmt.Printf("\nActive monitoring jobs with metrics tag: %d\n", complexQuery.Total)

	// ========================================
	// Feature 6: Complete Example Pipeline
	// ========================================
	fmt.Println("\n=== Feature 6: Complete Real-World Example ===")

	// Create a sophisticated data processing pipeline
	// with retries, timeouts, and proper error handling

	// Step 1: API data fetcher with retry
	sched.AddCron("api-fetcher", "@every 2m", func(ctx context.Context) error {
		fmt.Println("Fetching data from API...")
		// Simulate API call
		return nil
	},
		scheduler.WithGroup("production"),
		scheduler.WithTags("api", "critical"),
		scheduler.WithTimeout(30*time.Second),
		scheduler.WithRetryPolicy(scheduler.RetryExponential(3, time.Second, 10*time.Second)),
		scheduler.WithOverlapPolicy(scheduler.SkipIfRunning),
	)

	// Step 2: Data validator (depends on fetcher)
	sched.AddCron("data-validator", "@every 2m", func(ctx context.Context) error {
		fmt.Println("Validating data...")
		return nil
	},
		scheduler.WithGroup("production"),
		scheduler.WithTags("validation"),
		scheduler.WithTimeout(10*time.Second),
		scheduler.WithOverlapPolicy(scheduler.SerialQueue),
	)

	// Step 3: Alert system (runs independently)
	sched.AddCron("alert-checker", "@every 1m", func(ctx context.Context) error {
		fmt.Println("Checking for alerts...")
		return nil
	},
		scheduler.WithGroup("alerting"),
		scheduler.WithTags("critical", "monitoring"),
		scheduler.WithTimeout(5*time.Second),
	)

	// ========================================
	// Statistics and Monitoring
	// ========================================
	time.Sleep(1 * time.Second)

	fmt.Println("\n=== Scheduler Statistics ===")
	stats := sched.Stats()
	fmt.Printf("Queued: %d, Success: %d, Failure: %d\n", stats.Queued, stats.Success, stats.Failure)
	fmt.Printf("Dropped: %d, Retry: %d, Timeout: %d\n", stats.Dropped, stats.Retry, stats.Timeout)

	health := sched.Health()
	fmt.Printf("\nHealth: Queued=%d, Workers=%d, QueueSize=%d\n",
		health.Queued, health.Workers, health.QueueSize)

	// List all jobs summary
	allJobs := sched.List()
	fmt.Printf("\nTotal active jobs: %d\n", len(allJobs))

	// Keep running for demonstration
	fmt.Println("\n=== Scheduler running... (Press Ctrl+C to stop) ===")
	time.Sleep(5 * time.Second)
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}
