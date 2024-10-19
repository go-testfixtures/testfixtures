//go:build sqlite

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQLite(t *testing.T) {
	db := openDB(t, "sqlite3", os.Getenv("SQLITE_CONN_STRING"))
	loadSchemaInOneQuery(t, db, "testdata/schema/sqlite.sql")
	testLoader(t, db, "sqlite3")
}
