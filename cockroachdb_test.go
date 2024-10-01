//go:build cockroachdb

package testfixtures

import (
	"os"
	"testing"

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
			DangerousSkipTestDatabaseCheck(),
			UseDropConstraint(),
		)
	}
}
