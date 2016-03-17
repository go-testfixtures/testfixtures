package testfixtures

import (
	"database/sql"
)

// DataBaseHelper is the generic interface for the database helper
type DataBaseHelper interface {
	DisableTriggers(db *sql.DB) error
	EnableTriggers(db *sql.DB) error
	BeforeLoad(db *sql.DB) error
	AfterLoad(db *sql.DB) error
}
