package testfixtures

import (
	"database/sql"
)

type MockHelper struct {
	dbName string
}

func (*MockHelper) init(*sql.DB) error {
	return nil
}
func (*MockHelper) disableReferentialIntegrity(*sql.DB, loadFunction) error {
	return nil
}
func (*MockHelper) paramType() int {
	return 0
}
func (*MockHelper) tableNames(queryable) ([]string, error) {
	return nil, nil
}
func (*MockHelper) isTableModified(queryable, string) (bool, error) {
	return false, nil
}
func (*MockHelper) computeTablesChecksum(queryable) error {
	return nil
}
func (*MockHelper) quoteKeyword(string) string {
	return ""
}
func (*MockHelper) whileInsertOnTable(*sql.Tx, string, func() error) error {
	return nil
}
func (h *MockHelper) databaseName(queryable) (string, error) {
	return h.dbName, nil
}

func (h *MockHelper) cleanTableQuery(string) string {
	return ""
}

func (h *MockHelper) buildInsertSQL(queryable, string, []string, []string) (string, error) {
	return "", nil
}

// NewMockHelper returns MockHelper
func NewMockHelper(dbName string) *MockHelper {
	return &MockHelper{dbName: dbName}
}
