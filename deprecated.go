package testfixtures

import (
	"database/sql"
)

type (
	// Deprecated: Use Helper instead
	DataBaseHelper Helper

	// Deprecated: Use PostgreSQL{} instead
	PostgreSQLHelper struct {
		PostgreSQL
		UseAlterConstraint bool
	}

	// Deprecated: Use MySQL{} instead
	MySQLHelper struct {
		MySQL
	}

	// Deprecated: Use SQLite{} instead
	SQLiteHelper struct {
		SQLite
	}

	// Deprecated: Use SQLServer{} instead
	SQLServerHelper struct {
		SQLServer
	}

	// Deprecated: Use Oracle{} instead
	OracleHelper struct {
		Oracle
	}
)

func (h *PostgreSQLHelper) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
	h.PostgreSQL.UseAlterConstraint = h.UseAlterConstraint
	return h.PostgreSQL.disableReferentialIntegrity(db, loadFn)
}

// LoadFixtureFiles load all specified fixtures files into database:
// 		LoadFixtureFiles(db, &PostgreSQL{},
// 			"fixtures/customers.yml", "fixtures/orders.yml")
//			// add as many files you want
//
// Deprecated: Use NewFiles() and Load() instead.
func LoadFixtureFiles(db *sql.DB, helper Helper, files ...string) error {
	c, err := NewFiles(db, helper, files...)
	if err != nil {
		return err
	}

	return c.Load()
}

// LoadFixtures loads all fixtures in a given folder into the database:
// 		LoadFixtures("myfixturesfolder", db, &PostgreSQL{})
//
// Deprecated: Use NewFolder() and Load() instead.
func LoadFixtures(folderName string, db *sql.DB, helper Helper) error {
	c, err := NewFolder(db, helper, folderName)
	if err != nil {
		return err
	}

	return c.Load()
}
