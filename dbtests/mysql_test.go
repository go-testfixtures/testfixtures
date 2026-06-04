package dbtests

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-testfixtures/testfixtures/v3"
)

func TestMySQL(t *testing.T) {
	t.Parallel()
	connStr := createMySQLContainer(t)

	t.Run("Standard", func(t *testing.T) {
		db := openDB(t, "mysql", connStr)
		loadSchemaInBatchesBySplitter(t, db, "testdata/schema/mysql.sql", []byte(";\n"))
		testLoader(t, db, "mysql")
	})

	t.Run("WithMultipleStatementsSupport", func(t *testing.T) {
		db := openDB(t, "mysql", connStr+"?multiStatements=true")
		loadSchemaInOneQuery(t, db, "testdata/schema/mysql.sql")
		testLoader(t, db, "mysql", testfixtures.AllowMultipleStatementsInOneQuery())
	})
}
