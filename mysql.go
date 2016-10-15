package testfixtures

import (
	"database/sql"
	"fmt"
)

// MySQL is the MySQL helper for this package
type MySQL struct {
	baseHelper
}

func (*MySQL) paramType() int {
	return paramTypeQuestion
}

func (*MySQL) quoteKeyword(str string) string {
	return fmt.Sprintf("`%s`", str)
}

func (*MySQL) databaseName(db *sql.DB) (dbName string) {
	db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	return
}

func (*MySQL) whileInsertOnTable(tx *sql.Tx, tableName string, fn func() error) error {
	return fn()
}

func (h *MySQL) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
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
