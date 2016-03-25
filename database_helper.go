package testfixtures

import (
	"database/sql"
)

const (
	paramTypeDollar = iota + 1
	paramTypeQuestion
)

// DataBaseHelper is the generic interface for the database helper
type DataBaseHelper interface {
	disableTriggers(tx *sql.Tx) error
	enableTriggers(tx *sql.Tx) error
	beforeLoad(db *sql.DB) error
	afterLoad(db *sql.DB) error
	paramType() int
}
