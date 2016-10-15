package testfixtures

import (
	"database/sql"
	"fmt"
	"path/filepath"
)

// SQLite is the SQLite Helper for this package
type SQLite struct {
	baseHelper
}

func (*SQLite) paramType() int {
	return paramTypeQuestion
}

func (*SQLite) quoteKeyword(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}

func (*SQLite) databaseName(db *sql.DB) (dbName string) {
	var seq int
	var main string
	db.QueryRow("PRAGMA database_list").Scan(&seq, &main, &dbName)
	dbName = filepath.Base(dbName)
	return
}

func (*SQLite) whileInsertOnTable(tx *sql.Tx, tableName string, fn func() error) error {
	return fn()
}

func (*SQLite) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec("PRAGMA defer_foreign_keys = ON"); err != nil {
		return err
	}

	if err = loadFn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
