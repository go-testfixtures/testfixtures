package testfixtures

import (
	"database/sql"
	"fmt"
	"path/filepath"
)

// SQLiteHelper is the SQLite Helper for this package
type SQLiteHelper struct{}

func (SQLiteHelper) paramType() int {
	return paramTypeQuestion
}

func (SQLiteHelper) quoteKeyword(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}

func (SQLiteHelper) databaseName(db *sql.DB) (dbName string) {
	var seq int
	var main string
	db.QueryRow("PRAGMA database_list").Scan(&seq, &main, &dbName)
	dbName = filepath.Base(dbName)
	return
}

func (SQLiteHelper) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("PRAGMA defer_foreign_keys = ON")
	if err != nil {
		return err
	}

	err = loadFn(tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}
