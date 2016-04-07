package testfixtures

import (
	"database/sql"
	"errors"
	"regexp"
)

const (
	paramTypeDollar = iota + 1
	paramTypeQuestion
)

var (
	dbnameRegexp       = regexp.MustCompile("test")
	errNotTestDatabase = errors.New("Loading aborted because the database name does not contains \"test\"")
)

type loadFunction func(tx *sql.Tx) error

// DataBaseHelper is the generic interface for the database helper
type DataBaseHelper interface {
	disableReferentialIntegrity(*sql.DB, loadFunction) error
	paramType() int
	databaseName(*sql.DB) string
	quoteKeyword(string) string
}
