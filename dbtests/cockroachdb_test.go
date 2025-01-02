//go:build cockroachdb

package dbtests

import (
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

func TestCockroachDB(t *testing.T) {
	for _, dialect := range []string{"postgres", "pgx"} {
		db := openDB(t, dialect, os.Getenv("CRDB_CONN_STRING"))
		loadSchemaInOneQuery(t, db, "testdata/schema/cockroachdb.sql")
		testLoader(
			t,
			db,
			dialect,
			testfixtures.DangerousSkipTestDatabaseCheck(),
			testfixtures.UseDropConstraint(),
		)
	}
}
