// +build sqlserver

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
)

func TestSQLServer(t *testing.T) {
	testLoader(
		t,
		"sqlserver",
		os.Getenv("SQLSERVER_CONN_STRING"),
		"testdata/schema/sqlserver.sql",
		DangerousSkipTestDatabaseCheck(),
	)
}

func TestDeprecatedMssql(t *testing.T) {
	testLoader(
		t,
		"mssql",
		os.Getenv("SQLSERVER_CONN_STRING"),
		"testdata/schema/sqlserver.sql",
		DangerousSkipTestDatabaseCheck(),
	)
}
