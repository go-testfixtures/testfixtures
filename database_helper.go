package testfixtures

import (
    "database/sql"
)

type DataBaseHelper interface {
    DisableTriggers(db *sql.DB) error
    EnableTriggers(db *sql.DB) error
}
