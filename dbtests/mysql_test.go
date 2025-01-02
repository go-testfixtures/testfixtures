//go:build mysql

package dbtests

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-testfixtures/testfixtures/v3"
)

func TestMySQL(t *testing.T) {
	db := openDB(t, "mysql", os.Getenv("MYSQL_CONN_STRING"))
	loadSchemaInBatchesBySplitter(t, db, "testdata/schema/mysql.sql", []byte(";\n"))
	testLoader(t, db, "mysql")
}

func TestMySQLWithMultipleStatementsSupport(t *testing.T) {
	db := openDB(t, "mysql", os.Getenv("MYSQL_CONN_STRING")+"?multiStatements=true")
	loadSchemaInOneQuery(t, db, "testdata/schema/mysql.sql")
	testLoader(t, db, "mysql", testfixtures.AllowMultipleStatementsInOneQuery())
}
