package testfixtures

import (
	"cmp"
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-testfixtures/testfixtures/v3/shared"
)

type ParamType string

func (p ParamType) String() string {
	return string(p)
}

func (p ParamType) Valid() error {
	switch p {
	case ParamTypeDollar, ParamTypeQuestion, ParamTypeAtSign:
		return nil
	default:
		return fmt.Errorf("testfixtures: param type %s is not supported", p)
	}
}

const (
	ParamTypeDollar   ParamType = "$"
	ParamTypeQuestion ParamType = "?"
	ParamTypeAtSign   ParamType = "@"
)

type loadFunction func(tx *sql.Tx) error

type helper interface {
	init(*sql.DB) error
	disableReferentialIntegrity(*sql.DB, loadFunction) error
	paramType() ParamType
	getDefaultParamType() ParamType
	setCustomParamType(ParamType)
	databaseName(shared.Queryable) (string, error)
	tableNames(shared.Queryable) ([]string, error)
	isTableModified(shared.Queryable, string) (bool, error)
	computeTablesChecksum(shared.Queryable) error
	quoteKeyword(string) string
	whileInsertOnTable(*sql.Tx, string, func() error) error
	cleanTableQuery(string) string
	buildInsertSQL(q shared.Queryable, tableName string, columns, values []string) (string, error)
}

var (
	_ helper = &clickhouse{}
	_ helper = &spanner{}
	_ helper = &mySQL{}
	_ helper = &postgreSQL{}
	_ helper = &sqlite{}
	_ helper = &sqlserver{}
)

type baseHelper struct {
	customParamType ParamType
}

func (b *baseHelper) setCustomParamType(paramType ParamType) {
	b.customParamType = paramType
}

func (b *baseHelper) paramType() ParamType {
	return cmp.Or(b.customParamType, b.getDefaultParamType())
}

func (b *baseHelper) getDefaultParamType() ParamType {
	return ParamTypeDollar
}

// shared methods
func (baseHelper) init(_ *sql.DB) error {
	return nil
}

func (baseHelper) quoteKeyword(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}

func (baseHelper) whileInsertOnTable(_ *sql.Tx, _ string, fn func() error) error {
	return fn()
}

func (baseHelper) isTableModified(_ shared.Queryable, _ string) (bool, error) {
	return true, nil
}

func (baseHelper) computeTablesChecksum(_ shared.Queryable) error {
	return nil
}

func (baseHelper) cleanTableQuery(tableName string) string {
	return fmt.Sprintf("DELETE FROM %s", tableName)
}

func (h baseHelper) buildInsertSQL(_ shared.Queryable, tableName string, columns, values []string) (string, error) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(values, ", "),
	), nil
}
