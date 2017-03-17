package testfixtures

import (
	"database/sql"
	"path/filepath"
)

// SQLite is the SQLite Helper for this package
type SQLite struct {
	baseHelper
}

func (*SQLite) paramType() int {
	return paramTypeQuestion
}

func (*SQLite) databaseName(db *sql.DB) (dbName string) {
	var seq int
	var main string
	db.QueryRow("PRAGMA database_list").Scan(&seq, &main, &dbName)
	dbName = filepath.Base(dbName)
	return
}

func (*SQLite) tableNames(db *sql.DB) ([]string, error) {
	query := `
		SELECT name
		FROM sqlite_master
		WHERE type='table';
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
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
