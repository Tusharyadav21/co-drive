package database

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// RunMigrations applies every embedded .<direction>.sql file that has not been
// applied yet, recording each one in schema_migrations. The migrations
// directory is the single source of truth for the schema — nothing else in the
// codebase may create or alter a table.
func RunMigrations(db *sql.DB, direction string) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`); err != nil {
		return fmt.Errorf("failed to create schema_migrations ledger: %w", err)
	}

	entries, err := MigrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read embedded migrations: %w", err)
	}

	var files []string
	suffix := "." + direction + ".sql"
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)
	if direction == "down" {
		// Reverse order for down migrations
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}

	for _, filename := range files {
		version := strings.TrimSuffix(filename, suffix)

		applied, err := isApplied(db, version)
		if err != nil {
			return err
		}
		// Up skips what is already applied; down only undoes what is.
		if (direction == "up") == applied {
			continue
		}

		content, err := MigrationsFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if direction == "up" {
			_, err = db.Exec(`INSERT INTO schema_migrations (version) VALUES ($1);`, version)
		} else {
			_, err = db.Exec(`DELETE FROM schema_migrations WHERE version = ($1);`, version)
		}
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}
	}

	return nil
}

func isApplied(db *sql.DB, version string) (bool, error) {
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM schema_migrations WHERE version = $1;`, version).Scan(&n); err != nil {
		return false, fmt.Errorf("failed to read schema_migrations: %w", err)
	}
	return n > 0, nil
}
