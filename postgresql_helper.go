package testfixtures

import (
	"database/sql"
	"fmt"
	"strings"
)

// PostgreSQLHelper is the PG helper for this package
type PostgreSQLHelper struct {
	// UseAlterConstraint If true, the contraint disabling will do
	// using ALTER CONTRAINT sintax, only allowed in PG >= 9.4.
	// If false, the constraint disabling will use DISABLE TRIGGER ALL,
	// which requires SUPERUSER privileges.
	UseAlterConstraint bool

	nonDeferrableConstraints []pgContraint
}

type pgContraint struct {
	tableName      string
	constraintName string
}

func (PostgreSQLHelper) paramType() int {
	return paramTypeDollar
}

func (h *PostgreSQLHelper) getTables(tx *sql.Tx) ([]string, error) {
	sql := `
SELECT table_name
FROM information_schema.tables
WHERE table_schema='public'
  AND table_type='BASE TABLE';
`
	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}

	var tables []string
	defer rows.Close()
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}
	return tables, nil
}

func (h *PostgreSQLHelper) getSequences(tx *sql.Tx) ([]string, error) {
	sql := "SELECT relname FROM pg_class WHERE relkind = 'S'"
	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sequences []string
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

func (PostgreSQLHelper) getNonDeferrableConstraints(tx *sql.Tx) ([]pgContraint, error) {
	sql := `
SELECT table_name, constraint_name
FROM information_schema.table_constraints
WHERE constraint_type = 'FOREIGN KEY'
  AND is_deferrable = 'NO'`
	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var constraints []pgContraint
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

func (h *PostgreSQLHelper) disableTriggers(tx *sql.Tx) error {
	tables, err := h.getTables(tx)
	if err != nil {
		return err
	}
	sql := ""

	for _, table := range tables {
		sql += fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER ALL;", table)
	}

	_, err = tx.Exec(sql)
	return err
}

func (h *PostgreSQLHelper) enableTriggers(tx *sql.Tx) error {
	tables, err := h.getTables(tx)
	if err != nil {
		return err
	}
	sql := ""

	for _, table := range tables {
		sql += fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER ALL;", table)
	}

	_, err = tx.Exec(sql)
	return err
}

func (h *PostgreSQLHelper) makeConstraintsDeferrable(tx *sql.Tx) error {
	sql := ""
	for _, constraint := range h.nonDeferrableConstraints {
		sql += fmt.Sprintf("ALTER TABLE %s ALTER CONSTRAINT %s DEFERRABLE;", constraint.tableName, constraint.constraintName)
	}
	_, err := tx.Exec(sql)
	return err
}

func (h *PostgreSQLHelper) undoMakeConstraintsDeferrable(tx *sql.Tx) error {
	sql := ""
	for _, constraint := range h.nonDeferrableConstraints {
		sql += fmt.Sprintf("ALTER TABLE %s ALTER CONSTRAINT %s NOT DEFERRABLE;", constraint.tableName, constraint.constraintName)
	}
	_, err := tx.Exec(sql)
	return err
}

func (h *PostgreSQLHelper) disableReferentialIntegrity(tx *sql.Tx) error {
	if h.UseAlterConstraint {
		return h.makeConstraintsDeferrable(tx)
	} else {
		return h.disableTriggers(tx)
	}
}

func (h *PostgreSQLHelper) enableReferentialIntegrity(tx *sql.Tx) error {
	if h.UseAlterConstraint {
		return h.undoMakeConstraintsDeferrable(tx)
	} else {
		return h.enableTriggers(tx)
	}
}

func (h *PostgreSQLHelper) resetSequences(tx *sql.Tx) error {
	sequences, err := h.getSequences(tx)
	if err != nil {
		return err
	}

	for _, sequence := range sequences {
		var max int
		table := strings.Replace(sequence, "_id_seq", "", 1)
		row := tx.QueryRow(fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", table))
		err = row.Scan(&max)
		if err != nil {
			return err
		}

		if max == 0 {
			max = 1
		}
		_, err = tx.Exec(fmt.Sprintf("SELECT SETVAL('%s', %d)", sequence, max))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *PostgreSQLHelper) beforeLoad(tx *sql.Tx) error {
	var err error
	h.nonDeferrableConstraints, err = h.getNonDeferrableConstraints(tx)
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	return err
}

func (h *PostgreSQLHelper) afterLoad(tx *sql.Tx) error {
	return h.resetSequences(tx)
}
