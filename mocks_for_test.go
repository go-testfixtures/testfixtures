package testfixtures

import (
	"database/sql"

	"github.com/go-testfixtures/testfixtures/v3/shared"
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
func (*MockHelper) tableNames(shared.Queryable) ([]string, error) {
	return nil, nil
}
func (*MockHelper) isTableModified(shared.Queryable, string) (bool, error) {
	return false, nil
}
func (*MockHelper) computeTablesChecksum(shared.Queryable) error {
	return nil
}
func (*MockHelper) quoteKeyword(string) string {
	return ""
}
func (*MockHelper) whileInsertOnTable(*sql.Tx, string, func() error) error {
	return nil
}
func (h *MockHelper) databaseName(shared.Queryable) (string, error) {
	return h.dbName, nil
}

func (h *MockHelper) cleanTableQuery(string) string {
	return ""
}

func (h *MockHelper) buildInsertSQL(shared.Queryable, string, []string, []string) (string, error) {
	return "", nil
}

// NewMockHelper returns MockHelper
func NewMockHelper(dbName string) *MockHelper {
	return &MockHelper{dbName: dbName}
}
