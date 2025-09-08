package dbtests

import (
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-testfixtures/testfixtures/v3"
)

func TestSQLServer(t *testing.T) {
	t.Parallel()
	connStr := createSQLServerContainer(t)

	t.Run("SQLServer", func(t *testing.T) {
		testSQLServer(t, connStr, "sqlserver")
	})

	t.Run("DeprecatedMssql", func(t *testing.T) {
		testSQLServer(t, connStr, "mssql")
	})
}

func testSQLServer(t *testing.T, connStr string, dialect string) {
	t.Helper()
	db := openDB(t, dialect, connStr)
	loadSchemaInBatchesBySplitter(t, db, "testdata/schema/sqlserver.sql", []byte("GO\n"))
	testLoader(
		t,
		db,
		dialect,
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
}
