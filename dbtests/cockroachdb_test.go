package dbtests

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

func TestCockroachDB(t *testing.T) {
	t.Parallel()
	connStr := createCockroachDBContainer(t)

	for _, dialect := range []string{"postgres", "pgx"} {
		t.Run(dialect, func(t *testing.T) {
			db := openDB(t, dialect, connStr)
			loadSchemaInOneQuery(t, db, "testdata/schema/cockroachdb.sql")
			testLoader(
				t,
				db,
				dialect,
				testfixtures.DangerousSkipTestDatabaseCheck(),
				testfixtures.UseDropConstraint(),
			)
		})
	}
}
