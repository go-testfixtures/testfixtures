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
