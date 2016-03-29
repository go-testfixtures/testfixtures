// +build sqlite

package testfixtures

func init() {
	databases = append(databases, databaseTest{
		"sqlite3",
		"SQLITE_CONN_STRING",
		"test_schema/sqlite.sql",
		&SQLiteHelper{},
	})
}
