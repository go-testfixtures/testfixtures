// +build sqlite

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQLite(t *testing.T) {
	testTestFixtures(
		t,
		"sqlite3",
		os.Getenv("SQLITE_CONN_STRING"),
		"testdata/schema/sqlite.sql",
	)
}
