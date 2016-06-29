package testfixtures

import (
	"database/sql"
	"fmt"
)

// PostgreSQLHelper is the PG helper for this package
type PostgreSQLHelper struct {
	// UseAlterConstraint If true, the contraint disabling will do
	// using ALTER CONTRAINT sintax, only allowed in PG >= 9.4.
	// If false, the constraint disabling will use DISABLE TRIGGER ALL,
	// which requires SUPERUSER privileges.
	UseAlterConstraint bool
}

type pgContraint struct {
	tableName      string
	constraintName string
}

func (*PostgreSQLHelper) paramType() int {
	return paramTypeDollar
}

func (*PostgreSQLHelper) quoteKeyword(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}

func (*PostgreSQLHelper) databaseName(db *sql.DB) (dbName string) {
	db.QueryRow("SELECT current_database()").Scan(&dbName)
	return
}

func (*PostgreSQLHelper) whileInsertOnTable(tx *sql.Tx, tableName string, fn func() error) error {
	return fn()
}

func (h *PostgreSQLHelper) getTables(db *sql.DB) ([]string, error) {
	var tables []string

	sql := `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE';
`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}
	return tables, nil
}

func (h *PostgreSQLHelper) getSequences(db *sql.DB) ([]string, error) {
	var sequences []string

	sql := "SELECT relname FROM pg_class WHERE relkind = 'S'"
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var sequence string
		err = rows.Scan(&sequence)
		if err != nil {
			return nil, err
		}
		sequences = append(sequences, sequence)
	}
	return sequences, nil
}

func (*PostgreSQLHelper) getNonDeferrableConstraints(db *sql.DB) ([]pgContraint, error) {
	var constraints []pgContraint

	sql := `
SELECT table_name, constraint_name
FROM information_schema.table_constraints
WHERE constraint_type = 'FOREIGN KEY'
  AND is_deferrable = 'NO'`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var constraint pgContraint
		err = rows.Scan(&constraint.tableName, &constraint.constraintName)
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	return constraints, nil
}

func (h *PostgreSQLHelper) disableTriggers(db *sql.DB, loadFn loadFunction) error {
	tables, err := h.getTables(db)
	if err != nil {
		return err
	}

	defer func() {
		// re-enable triggers after load
		var sql string
		for _, table := range tables {
			sql += fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER ALL;", h.quoteKeyword(table))
		}
		db.Exec(sql)
	}()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	var sql string
	for _, table := range tables {
		sql += fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER ALL;", h.quoteKeyword(table))
	}
	_, err = tx.Exec(sql)
	if err != nil {
		return err
	}

	err = loadFn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

func (h *PostgreSQLHelper) makeConstraintsDeferrable(db *sql.DB, loadFn loadFunction) error {
	nonDeferrableConstraints, err := h.getNonDeferrableConstraints(db)
	if err != nil {
		return err
	}

	defer func() {
		// ensure constraint being not deferrable again after load
		var sql string
		for _, constraint := range nonDeferrableConstraints {
			sql += fmt.Sprintf("ALTER TABLE %s ALTER CONSTRAINT %s NOT DEFERRABLE;", h.quoteKeyword(constraint.tableName), h.quoteKeyword(constraint.constraintName))
		}
		db.Exec(sql)
	}()

	var sql string
	for _, constraint := range nonDeferrableConstraints {
		sql += fmt.Sprintf("ALTER TABLE %s ALTER CONSTRAINT %s DEFERRABLE;", h.quoteKeyword(constraint.tableName), h.quoteKeyword(constraint.constraintName))
	}
	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		return nil
	}

	err = loadFn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}

func (h *PostgreSQLHelper) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
	// ensure sequences being reset after load
	defer h.resetSequences(db)

	if h.UseAlterConstraint {
		return h.makeConstraintsDeferrable(db, loadFn)
	} else {
		return h.disableTriggers(db, loadFn)
	}
}

func (h *PostgreSQLHelper) resetSequences(db *sql.DB) error {
	sequences, err := h.getSequences(db)
	if err != nil {
		return err
	}

	for _, sequence := range sequences {
		_, err = db.Exec(fmt.Sprintf("SELECT SETVAL('%s', %d)", sequence, resetSequencesTo))
		if err != nil {
			return err
		}
	}
	return nil
}
