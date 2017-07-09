package testfixtures

import (
	"database/sql"
	"fmt"
)

const (
	paramTypeDollar = iota + 1
	paramTypeQuestion
	paramTypeColon
)

type loadFunction func(tx *sql.Tx) error

// Helper is the generic interface for the database helper
type Helper interface {
	init(*sql.DB) error
	disableReferentialIntegrity(*sql.DB, loadFunction) error
	paramType() int
	databaseName(*sql.DB) string
	tableNames(*sql.DB) ([]string, error)
	isTableModified(*sql.DB, string) (bool, error)
	tablesLoaded(*sql.DB) error
	quoteKeyword(string) string
	whileInsertOnTable(*sql.Tx, string, func() error) error
}

type baseHelper struct{}

func (*baseHelper) init(_ *sql.DB) error {
	return nil
}

func (*baseHelper) quoteKeyword(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}

func (*baseHelper) whileInsertOnTable(_ *sql.Tx, _ string, fn func() error) error {
	return fn()
}

func (*baseHelper) isTableModified(_ *sql.DB, _ string) (bool, error) {
	return true, nil
}

func (*baseHelper) tablesLoaded(_ *sql.DB) error {
	return nil
}
