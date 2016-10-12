package testfixtures

import (
	"database/sql"
)

type (
	DataBaseHelper Helper

	PostgreSQLHelper struct {
		PostgreSQL
		UseAlterConstraint bool
	}
	MySQLHelper struct {
		MySQL
	}
	SQLiteHelper struct {
		SQLite
	}
	SQLServerHelper struct {
		SQLServer
	}
	OracleHelper struct {
		Oracle
	}
)

func (h *PostgreSQLHelper) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
	h.PostgreSQL.UseAlterConstraint = h.UseAlterConstraint
	return h.PostgreSQL.disableReferentialIntegrity(db, loadFn)
}

// LoadFixtureFiles load all specified fixtures files into database:
// 		LoadFixtureFiles(db, &PostgreSQLHelper{},
// 			"fixtures/customers.yml", "fixtures/orders.yml")
//			// add as many files you want
func LoadFixtureFiles(db *sql.DB, h Helper, files ...string) error {
	c, err := NewFiles(db, h, files...)
	if err != nil {
		return err
	}

	return c.Load()
}

// LoadFixtures loads all fixtures in a given folder into the database:
// 		LoadFixtures("myfixturesfolder", db, &PostgreSQLHelper{})
func LoadFixtures(folderName string, db *sql.DB, h Helper) error {
	c, err := NewFolder(db, h, folderName)
	if err != nil {
		return err
	}

	return c.Load()
}
