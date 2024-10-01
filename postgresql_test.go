//go:build postgresql

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

func TestPostgreSQL(t *testing.T) {
	testPostgreSQL(t)
}

func TestPostgreSQLWithAlterConstraint(t *testing.T) {
	testPostgreSQL(t, UseAlterConstraint())
}

func TestPostgreSQLWithDropConstraint(t *testing.T) {
	testPostgreSQL(t, UseDropConstraint())
}

func testPostgreSQL(t *testing.T, additionalOptions ...func(*Loader) error) {
	t.Helper()
	for _, dialect := range []string{"postgres", "pgx"} {
		db := openDB(t, dialect, os.Getenv("PG_CONN_STRING"))
		loadSchemaInOneQuery(t, db, "testdata/schema/postgresql.sql")
		testLoader(
			t,
			db,
			dialect,
			additionalOptions...,
		)
	}

}
