package testfixtures

import (
	"database/sql"
)

// MySQLHelper is the MySQL helper for this package
type MySQLHelper struct {
}

func (MySQLHelper) paramType() int {
	return paramTypeQuestion
}

func (h *MySQLHelper) disableTriggers(tx *sql.Tx) error {
	_, err := tx.Exec("SET FOREIGN_KEY_CHECKS = 0")
	return err
}

func (h *MySQLHelper) enableTriggers(tx *sql.Tx) error {
	_, err := tx.Exec("SET FOREIGN_KEY_CHECKS = 1")
	return err
}

func (h *MySQLHelper) beforeLoad(db *sql.DB) error {
	return nil
}

func (h *MySQLHelper) afterLoad(db *sql.DB) error {
	return nil
}
