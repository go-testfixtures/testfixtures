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
	disableReferentialIntegrity(tx *sql.Tx) error
	enableReferentialIntegrity(tx *sql.Tx) error
	beforeLoad(tx *sql.Tx) error
	afterLoad(tx *sql.Tx) error
	paramType() int
}
