package testfixtures

import (
	"database/sql"
)

// MySQLHelper is the MySQL helper for this package
type MySQLHelper struct{}

func (MySQLHelper) paramType() int {
	return paramTypeQuestion
}

func (MySQLHelper) databaseName(db *sql.DB) (dbName string) {
	db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	return
}

func (h *MySQLHelper) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
	// re-enable after load
	defer db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		return err
	}

	err = loadFn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}
