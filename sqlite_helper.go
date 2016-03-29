package testfixtures

import (
	"database/sql"
)

// SQLiteHelper is the SQLite Helper for this package
type SQLiteHelper struct{}

func (SQLiteHelper) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
	tx, err := db.Begin()
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

func (SQLiteHelper) paramType() int {
	return paramTypeQuestion
}
