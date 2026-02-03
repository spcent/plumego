package commands

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/spcent/plumego/cmd/plumego/internal/migrate"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

type MigrateCmd struct{}

func (c *MigrateCmd) Name() string {
	return "migrate"
}

func (c *MigrateCmd) Short() string {
	return "Manage database migrations"
}

func (c *MigrateCmd) Long() string {
	return `Manage database migrations using SQL files.

Commands:
  up       Apply pending migrations
  down     Roll back applied migrations
  status   Show migration status
  create   Create new migration files

Examples:
  plumego migrate status --driver sqlite3 --db-url ./app.db
  plumego migrate up --driver mysql --db-url "user:pass@tcp(localhost:3306)/app"
  plumego migrate down --driver postgres --db-url "postgres://localhost/app" --steps 1
  plumego migrate create add_users_table`
}

func (c *MigrateCmd) Flags() []Flag {
	return []Flag{
		{Name: "dir", Default: "./migrations", Usage: "Migrations directory"},
		{Name: "db-url", Default: "", Usage: "Database connection string"},
		{Name: "driver", Default: "", Usage: "Database driver name"},
		{Name: "steps", Default: "0", Usage: "Number of migrations to apply/rollback (0 = all)"},
	}
}

func (c *MigrateCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)

	dir := fs.String("dir", "./migrations", "Migrations directory")
	dbURL := fs.String("db-url", "", "Database connection string")
	driver := fs.String("driver", "", "Database driver name")
	steps := fs.Int("steps", 0, "Number of migrations to apply/rollback (0 = all)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	out := ctx.Out

	subcommand := "status"
	if fs.NArg() > 0 {
		subcommand = fs.Arg(0)
	}

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		return out.Error(fmt.Sprintf("invalid directory: %v", err), 1)
	}

	switch subcommand {
	case "create":
		name := ""
		if fs.NArg() > 1 {
			name = fs.Arg(1)
		}
		if name == "" {
			return out.Error("migration name is required", 1)
		}

		migration, err := migrate.CreateMigrationFiles(absDir, name, time.Now())
		if err != nil {
			return out.Error(fmt.Sprintf("failed to create migration: %v", err), 1)
		}

		result := map[string]any{
			"version":   migration.Version,
			"name":      migration.Name,
			"up_path":   migration.UpPath,
			"down_path": migration.DownPath,
			"directory": absDir,
		}

		return out.Success("Migration files created", result)
	case "status", "up", "down":
		if *driver == "" || *dbURL == "" {
			return out.Error("driver and db-url are required", 1)
		}
		return c.runWithDatabase(out, subcommand, absDir, *driver, *dbURL, *steps)
	default:
		return out.Error(fmt.Sprintf("unknown subcommand: %s", subcommand), 1)
	}
}

func (c *MigrateCmd) runWithDatabase(out *output.Formatter, subcommand, dir, driver, dbURL string, steps int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open(driver, dbURL)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to open database: %v", err), 1)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return out.Error(fmt.Sprintf("failed to connect: %v", err), 1)
	}

	if err := migrate.EnsureSchemaTable(ctx, db); err != nil {
		return out.Error(fmt.Sprintf("failed to ensure schema table: %v", err), 1)
	}

	migrations, err := migrate.LoadMigrations(dir)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to load migrations: %v", err), 1)
	}

	applied, err := migrate.FetchApplied(ctx, db)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to fetch applied migrations: %v", err), 1)
	}

	switch subcommand {
	case "status":
		return c.reportStatus(out, migrations, applied)
	case "up":
		return c.applyUp(out, ctx, db, driver, migrations, applied, steps)
	case "down":
		return c.applyDown(out, ctx, db, driver, migrations, applied, steps)
	default:
		return out.Error(fmt.Sprintf("unknown subcommand: %s", subcommand), 1)
	}
}

func (c *MigrateCmd) reportStatus(out *output.Formatter, migrations []migrate.Migration, applied []migrate.AppliedMigration) error {
	appliedMap := make(map[string]migrate.AppliedMigration)
	for _, entry := range applied {
		appliedMap[entry.Version] = entry
	}

	var pending []map[string]any
	for _, migration := range migrations {
		if _, ok := appliedMap[migration.Version]; ok {
			continue
		}
		pending = append(pending, map[string]any{
			"version": migration.Version,
			"name":    migration.Name,
			"up_path": migration.UpPath,
		})
	}

	currentVersion := ""
	if len(applied) > 0 {
		currentVersion = applied[len(applied)-1].Version
	}

	result := map[string]any{
		"applied":         applied,
		"pending":         pending,
		"current_version": currentVersion,
		"total":           len(migrations),
	}

	return out.Success("Migration status", result)
}

func (c *MigrateCmd) applyUp(out *output.Formatter, ctx context.Context, db *sql.DB, driver string, migrations []migrate.Migration, applied []migrate.AppliedMigration, steps int) error {
	appliedMap := make(map[string]struct{})
	for _, entry := range applied {
		appliedMap[entry.Version] = struct{}{}
	}

	var pending []migrate.Migration
	for _, migration := range migrations {
		if _, ok := appliedMap[migration.Version]; ok {
			continue
		}
		pending = append(pending, migration)
	}

	if steps > 0 && len(pending) > steps {
		pending = pending[:steps]
	}

	if len(pending) == 0 {
		return out.Error("no migrations to apply", 2)
	}

	var appliedResults []map[string]any
	for _, migration := range pending {
		duration, err := migrate.ApplyUp(ctx, db, driver, migration, time.Now())
		if err != nil {
			return out.Error(fmt.Sprintf("failed to apply migration %s: %v", migration.Version, err), 1)
		}

		appliedResults = append(appliedResults, map[string]any{
			"version":     migration.Version,
			"name":        migration.Name,
			"duration_ms": duration.Milliseconds(),
		})
	}

	newApplied, err := migrate.FetchApplied(ctx, db)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to fetch applied migrations: %v", err), 1)
	}

	result := map[string]any{
		"command":         "up",
		"applied":         appliedResults,
		"current_version": latestVersion(newApplied),
		"pending":         pendingVersions(migrations, newApplied),
	}

	return out.Success("Migrations applied", result)
}

func (c *MigrateCmd) applyDown(out *output.Formatter, ctx context.Context, db *sql.DB, driver string, migrations []migrate.Migration, applied []migrate.AppliedMigration, steps int) error {
	if len(applied) == 0 {
		return out.Error("no migrations to roll back", 2)
	}

	migrationMap := make(map[string]migrate.Migration)
	for _, migration := range migrations {
		migrationMap[migration.Version] = migration
	}

	rollbackCount := len(applied)
	if steps > 0 && steps < rollbackCount {
		rollbackCount = steps
	}

	toRollback := applied[len(applied)-rollbackCount:]
	for i := len(toRollback)/2 - 1; i >= 0; i-- {
		opp := len(toRollback) - 1 - i
		toRollback[i], toRollback[opp] = toRollback[opp], toRollback[i]
	}

	var rolledBack []map[string]any
	for _, entry := range toRollback {
		migration, ok := migrationMap[entry.Version]
		if !ok {
			return out.Error(fmt.Sprintf("missing migration files for version %s", entry.Version), 1)
		}

		duration, err := migrate.ApplyDown(ctx, db, driver, migration)
		if err != nil {
			return out.Error(fmt.Sprintf("failed to roll back migration %s: %v", migration.Version, err), 1)
		}

		rolledBack = append(rolledBack, map[string]any{
			"version":     migration.Version,
			"name":        migration.Name,
			"duration_ms": duration.Milliseconds(),
		})
	}

	newApplied, err := migrate.FetchApplied(ctx, db)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to fetch applied migrations: %v", err), 1)
	}

	result := map[string]any{
		"command":         "down",
		"rolled_back":     rolledBack,
		"current_version": latestVersion(newApplied),
		"pending":         pendingVersions(migrations, newApplied),
	}

	return out.Success("Migrations rolled back", result)
}

func latestVersion(applied []migrate.AppliedMigration) string {
	if len(applied) == 0 {
		return ""
	}
	return applied[len(applied)-1].Version
}

func pendingVersions(migrations []migrate.Migration, applied []migrate.AppliedMigration) []string {
	appliedSet := make(map[string]struct{})
	for _, entry := range applied {
		appliedSet[entry.Version] = struct{}{}
	}

	var pending []string
	for _, migration := range migrations {
		if _, ok := appliedSet[migration.Version]; ok {
			continue
		}
		pending = append(pending, migration.Version)
	}

	sort.Strings(pending)
	return pending
}
