package dbtests

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

func TestPostgreSQL(t *testing.T) {
	t.Parallel()
	connStr := createPostgreSQLContainer(t)

	t.Run("Standard", func(t *testing.T) {
		testPostgreSQL(t, connStr)
	})

	t.Run("WithAlterConstraint", func(t *testing.T) {
		testPostgreSQL(t, connStr, testfixtures.UseAlterConstraint())
	})

	t.Run("WithDropConstraint", func(t *testing.T) {
		testPostgreSQL(t, connStr, testfixtures.UseDropConstraint())
	})

	t.Run("RestrictTableScanningBySearchPath", func(t *testing.T) {
		testPostgreSQLRestrictTableScanningBySearchPath(t, connStr)
	})
}

func testPostgreSQL(t *testing.T, connStr string, additionalOptions ...func(*testfixtures.Loader) error) {
	t.Helper()
	for _, dialect := range []string{"postgres", "pgx"} {
		t.Run(dialect, func(t *testing.T) {
			db := openDB(t, dialect, connStr)
			loadSchemaInOneQuery(t, db, "testdata/schema/postgresql.sql")
			testLoader(
				t,
				db,
				dialect,
				additionalOptions...,
			)
		})
	}
}

func testPostgreSQLRestrictTableScanningBySearchPath(t *testing.T, connStr string) {
	t.Helper()

	constraintModes := []struct {
		name   string
		option func(*testfixtures.Loader) error
	}{
		{name: "disable_triggers"},
		{name: "alter_constraints", option: testfixtures.UseAlterConstraint()},
		{name: "drop_constraints", option: testfixtures.UseDropConstraint()},
	}

	for _, dialect := range []string{"postgres", "pgx"} {
		t.Run(dialect, func(t *testing.T) {
			for _, constraintMode := range constraintModes {
				t.Run(constraintMode.name, func(t *testing.T) {
					testPostgreSQLRestrictTableScanningBySearchPathConstraintMode(t, dialect, connStr, constraintMode.option)
				})
			}
		})
	}
}

func testPostgreSQLRestrictTableScanningBySearchPathConstraintMode(
	t *testing.T,
	dialect string,
	connStr string,
	constraintOption func(*testfixtures.Loader) error,
) {
	t.Helper()

	suffix := strings.ReplaceAll(uuid.NewString(), "-", "_")
	schema := "testfixtures_search_path_" + suffix
	otherSchema := schema + "_other"

	adminDB := openDB(t, dialect, connStr)
	createSearchPathTestSchema(t, adminDB, schema)
	createSearchPathTestSchema(t, adminDB, otherSchema)
	t.Cleanup(func() {
		_, _ = adminDB.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS %q CASCADE`, schema))
		_, _ = adminDB.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS %q CASCADE`, otherSchema))
	})

	db := openDB(t, dialect, withSearchPath(t, connStr, schema))
	options := []func(*testfixtures.Loader) error{
		testfixtures.Database(db),
		testfixtures.Dialect(dialect),
		testfixtures.RestrictTableScanningBySearchPath(),
		testfixtures.Directory("testdata/fixtures_search_path"),
	}
	if constraintOption != nil {
		options = append(options, constraintOption)
	}

	loader, err := testfixtures.New(options...)
	if err != nil {
		t.Fatalf("failed to create loader: %v", err)
	}

	if _, err := adminDB.Exec(fmt.Sprintf(`DROP SCHEMA %q CASCADE`, otherSchema)); err != nil {
		t.Fatalf("failed to drop unrelated schema: %v", err)
	}
	if err := loader.Load(); err != nil {
		t.Fatalf("failed to load fixtures: %v", err)
	}

	var name string
	if err := db.QueryRow("SELECT name FROM widgets WHERE id = 1").Scan(&name); err != nil {
		t.Fatalf("failed to query loaded fixture: %v", err)
	}
	if name != "scoped widget" {
		t.Fatalf("unexpected fixture name: %q", name)
	}
}

func createSearchPathTestSchema(t *testing.T, db *sql.DB, schema string) {
	t.Helper()

	query := fmt.Sprintf(`
		CREATE SCHEMA %q;
		CREATE TABLE %q.parents (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE %q.widgets (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			parent_id BIGINT NOT NULL REFERENCES %q.parents(id),
			name TEXT NOT NULL
		);
	`, schema, schema, schema, schema)
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("failed to create search_path test schema: %v", err)
	}
}

func withSearchPath(t *testing.T, connStr, schema string) string {
	t.Helper()

	parsed, err := url.Parse(connStr)
	if err == nil && parsed.Scheme != "" {
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}

	return connStr + " search_path=" + schema
}
