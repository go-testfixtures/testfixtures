package dbtests

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
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
