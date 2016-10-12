package testfixtures

import (
	"database/sql"
	"errors"
	"regexp"
)

const (
	paramTypeDollar = iota + 1
	paramTypeQuestion
	paramTypeColon
)

var (
	dbnameRegexp       = regexp.MustCompile("(?i)test")
	errNotTestDatabase = errors.New("Loading aborted because the database name does not contains \"test\"")
)

type loadFunction func(tx *sql.Tx) error

// Helper is the generic interface for the database helper
type Helper interface {
	disableReferentialIntegrity(*sql.DB, loadFunction) error
	paramType() int
	databaseName(*sql.DB) string
	quoteKeyword(string) string
	whileInsertOnTable(*sql.Tx, string, func() error) error
}
