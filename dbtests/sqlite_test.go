package dbtests

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/mattn/go-sqlite3"
)

func TestSQLite(t *testing.T) {
	t.Parallel()

	connStr := createSQLite(t)
	db := openDB(t, "sqlite3", connStr)
	loadSchemaInOneQuery(t, db, "testdata/schema/sqlite.sql")
	testLoader(t, db, "sqlite3", testfixtures.DangerousSkipTestDatabaseCheck())
}
