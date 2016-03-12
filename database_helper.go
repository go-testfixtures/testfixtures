package testfixtures

import (
	"database/sql"
)

// DataBaseHelper is the generic interface for the database helper
type DataBaseHelper interface {
	DisableTriggers(db *sql.DB) error
	EnableTriggers(db *sql.DB) error
}
