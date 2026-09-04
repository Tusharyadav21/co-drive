package database

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"co-drive/config"
	"github.com/google/uuid"
)

// SetupTestDB initializes an isolated PostgreSQL database connection for tests.
// It creates a dedicated temporary schema (search_path) so each test runs in complete
// isolation, and registers a t.Cleanup hook to drop the schema and close the connection.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}

	url := cfg.Database.URL
	if url == "" {
		url = "postgres://localhost:5432/co-drive?sslmode=disable"
	}

	db, err := NewPostgres(url)
	if err != nil {
		t.Fatalf("failed to connect to postgres test database (%s): %v", url, err)
	}

	schemaName := fmt.Sprintf("test_%s", strings.ReplaceAll(uuid.NewString(), "-", "_"))
	if _, err := db.Exec(fmt.Sprintf("CREATE SCHEMA %s; SET search_path TO %s;", schemaName, schemaName)); err != nil {
		db.Close()
		t.Fatalf("failed to set up test schema %s: %v", schemaName, err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE;", schemaName))
		_ = db.Close()
	})

	if err := RunMigrations(db, "up"); err != nil {
		t.Fatalf("failed to run migrations in test schema %s: %v", schemaName, err)
	}

	return db
}
