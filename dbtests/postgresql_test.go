//go:build postgresql

package dbtests

import (
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

func TestPostgreSQL(t *testing.T) {
	testPostgreSQL(t)
}

func TestPostgreSQLWithAlterConstraint(t *testing.T) {
	testPostgreSQL(t, testfixtures.UseAlterConstraint())
}

func TestPostgreSQLWithDropConstraint(t *testing.T) {
	testPostgreSQL(t, testfixtures.UseDropConstraint())
}

func testPostgreSQL(t *testing.T, additionalOptions ...func(*testfixtures.Loader) error) {
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
