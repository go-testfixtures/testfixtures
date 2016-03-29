// +build sqlite

package testfixtures

import (
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	databases = append(databases, databaseTest{
		"sqlite3",
		"SQLITE_CONN_STRING",
		"test_schema/sqlite.sql",
		&SQLiteHelper{},
	})
}
