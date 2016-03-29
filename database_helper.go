package testfixtures

import (
	"database/sql"
)

const (
	paramTypeDollar = iota + 1
	paramTypeQuestion
)

type loadFunction func(tx *sql.Tx) error

// DataBaseHelper is the generic interface for the database helper
type DataBaseHelper interface {
	disableReferentialIntegrity(*sql.DB, loadFunction) error
	paramType() int
}
