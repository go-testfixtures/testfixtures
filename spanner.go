package testfixtures

import (
	"database/sql"
	"fmt"

	_ "github.com/googleapis/go-sql-spanner"
)

type spanner struct {
	baseHelper

	cleanTableFn func(string) string
	constraints  []spannerConstraint
}

type spannerConstraint struct {
	constraintName   string
	referencingTable string
	foreignKeyColumn string
	referenceTable   string
	referenceColumn  string
}

func (h *spanner) init(db *sql.DB) error {
	var err error

	if h.cleanTableFn == nil {
		h.cleanTableFn = func(tableName string) string {
			return fmt.Sprintf("DELETE FROM %s WHERE true;", tableName)
		}
	}

	h.constraints, err = h.getConstraints(db)
	if err != nil {
		return err
	}

	return nil
}

func (*spanner) paramType() int {
	return paramTypeAtSign
}

func (*spanner) quoteKeyword(str string) string {
	return str
}

func (*spanner) databaseName(q queryable) (string, error) {
	return "testdb", nil
}

func (h *spanner) tableNames(q queryable) ([]string, error) {
	query := `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = '';
	`

	rows, err := q.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var tables []string
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func (h *spanner) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	return h.dropAndRecreateConstraints(db, loadFn)
}

// splitter is a batchSplitter interface implementation. We need it for
// spanner because spanner doesn't support multi-statements.
func (*spanner) splitter() []byte {
	return []byte(";\n")
}

func (h *spanner) cleanTableQuery(tableName string) string {
	if h.cleanTableFn == nil {
		return h.baseHelper.cleanTableQuery(tableName)
	}

	return h.cleanTableFn(tableName)
}

func (h *spanner) getConstraints(q queryable) ([]spannerConstraint, error) {
	var constraints []spannerConstraint

	const sql = `
		SELECT tc.CONSTRAINT_NAME, key.TABLE_NAME, key.COLUMN_NAME, ref.TABLE_NAME, ref.COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE key
			JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc ON key.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
			JOIN INFORMATION_SCHEMA.CONSTRAINT_COLUMN_USAGE ref ON ref.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
		WHERE tc.CONSTRAINT_TYPE = 'FOREIGN KEY';
		`
	rows, err := q.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var constraint spannerConstraint
		if err = rows.Scan(
			&constraint.constraintName,
			&constraint.referencingTable,
			&constraint.foreignKeyColumn,
			&constraint.referenceTable,
			&constraint.referenceColumn,
		); err != nil {
			return nil, err
		}

		constraints = append(constraints, constraint)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return constraints, nil
}

func (h *spanner) dropAndRecreateConstraints(db *sql.DB, loadFn loadFunction) (err error) {
	defer func() {
		// Re-create constraints again after load
		for _, constraint := range h.constraints {
			cmd := fmt.Sprintf(
				`ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)`,
				constraint.referencingTable,
				constraint.constraintName,
				constraint.foreignKeyColumn,
				constraint.referenceTable,
				constraint.referenceColumn,
			)

			if _, err2 := db.Exec(cmd); err2 != nil && err == nil {
				err = err2
			}
		}
	}()

	for _, constraint := range h.constraints {
		cmd := fmt.Sprintf(
			`ALTER TABLE %s DROP CONSTRAINT %s`,
			constraint.referencingTable,
			constraint.constraintName,
		)
		if _, err := db.Exec(cmd); err != nil {
			return err
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err = loadFn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
