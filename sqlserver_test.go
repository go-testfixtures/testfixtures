//go:build sqlserver

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
)

func TestSQLServer(t *testing.T) {
	testSQLServer(t, "sqlserver")
}

func TestDeprecatedMssql(t *testing.T) {
	testSQLServer(t, "mssql")
}

func testSQLServer(t *testing.T, dialect string) {
	t.Helper()
	db := openDB(t, dialect, os.Getenv("SQLSERVER_CONN_STRING"))
	loadSchemaInBatchesBySplitter(t, db, "testdata/schema/sqlserver.sql", []byte("GO\n"))
	testLoader(
		t,
		db,
		dialect,
		DangerousSkipTestDatabaseCheck(),
	)
}
