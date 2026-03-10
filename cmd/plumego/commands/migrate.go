package commands

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"time"

	"github.com/spcent/plumego/cmd/plumego/internal/migrate"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

type MigrateCmd struct{}

func (c *MigrateCmd) Name() string  { return "migrate" }
func (c *MigrateCmd) Short() string { return "Manage database migrations" }

func (c *MigrateCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dir := fs.String("dir", "./migrations", "Migrations directory")
	dbURL := fs.String("db-url", "", "Database connection string")
	driver := fs.String("driver", "", "Database driver name")
	steps := fs.Int("steps", 0, "Number of migrations to apply/rollback (0 = all)")

	positionals, err := parseInterspersedFlags(fs, args)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	out := ctx.Out

	subcommand := "status"
	if len(positionals) > 0 {
		subcommand = positionals[0]
		positionals = positionals[1:]
	}

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		return out.Error(fmt.Sprintf("invalid directory: %v", err), 1)
	}

	switch subcommand {
	case "create":
		name := ""
		if len(positionals) > 0 {
			name = positionals[0]
		}
		if len(positionals) > 1 {
			return out.Error(fmt.Sprintf("unexpected arguments: %v", positionals[1:]), 1)
		}
		if name == "" {
			return out.Error("migration name is required", 1)
		}

		migration, err := migrate.CreateMigrationFiles(absDir, name, time.Now())
		if err != nil {
			return out.Error(fmt.Sprintf("failed to create migration: %v", err), 1)
		}

		return out.Success("Migration files created", map[string]any{
			"version":   migration.Version,
			"name":      migration.Name,
			"up_path":   migration.UpPath,
			"down_path": migration.DownPath,
			"directory": absDir,
		})
	case "status", "up", "down":
		if len(positionals) > 0 {
			return out.Error(fmt.Sprintf("unexpected arguments: %v", positionals), 1)
		}
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

func appliedVersionSet(applied []migrate.AppliedMigration) map[string]struct{} {
	set := make(map[string]struct{}, len(applied))
	for _, entry := range applied {
		set[entry.Version] = struct{}{}
	}
	return set
}

func pendingMigrations(migrations []migrate.Migration, applied []migrate.AppliedMigration) []migrate.Migration {
	set := appliedVersionSet(applied)
	var pending []migrate.Migration
	for _, m := range migrations {
		if _, ok := set[m.Version]; !ok {
			pending = append(pending, m)
		}
	}
	return pending
}

func (c *MigrateCmd) reportStatus(out *output.Formatter, migrations []migrate.Migration, applied []migrate.AppliedMigration) error {
	pending := pendingMigrations(migrations, applied)

	var pendingInfo []map[string]any
	for _, m := range pending {
		pendingInfo = append(pendingInfo, map[string]any{
			"version": m.Version,
			"name":    m.Name,
			"up_path": m.UpPath,
		})
	}

	return out.Success("Migration status", map[string]any{
		"applied":         applied,
		"pending":         pendingInfo,
		"current_version": latestVersion(applied),
		"total":           len(migrations),
	})
}

func (c *MigrateCmd) applyUp(out *output.Formatter, ctx context.Context, db *sql.DB, driver string, migrations []migrate.Migration, applied []migrate.AppliedMigration, steps int) error {
	pending := pendingMigrations(migrations, applied)

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

	return out.Success("Migrations applied", map[string]any{
		"command":         "up",
		"applied":         appliedResults,
		"current_version": latestVersion(newApplied),
		"pending":         pendingVersions(migrations, newApplied),
	})
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

	return out.Success("Migrations rolled back", map[string]any{
		"command":         "down",
		"rolled_back":     rolledBack,
		"current_version": latestVersion(newApplied),
		"pending":         pendingVersions(migrations, newApplied),
	})
}

func latestVersion(applied []migrate.AppliedMigration) string {
	if len(applied) == 0 {
		return ""
	}
	return applied[len(applied)-1].Version
}

func pendingVersions(migrations []migrate.Migration, applied []migrate.AppliedMigration) []string {
	pending := pendingMigrations(migrations, applied)
	versions := make([]string, len(pending))
	for i, m := range pending {
		versions[i] = m.Version
	}
	sort.Strings(versions)
	return versions
}
