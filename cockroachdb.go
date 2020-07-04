package testfixtures

import (
	"database/sql"
	"fmt"
	"strings"
)

const fkName string = "FOREIGN KEY"

type cockroachDB struct {
	baseHelper

	skipResetSequences bool
	resetSequencesTo   int64

	constraints []crdbConstraint

	tables    []string
	sequences []string
}

type crdbConstraint struct {
	tableName      string
	constraintName string
	constraintType string
	details        string
	validated      bool
}

func (h *cockroachDB) init(db *sql.DB) error {
	var err error

	h.tables, err = h.tableNames(db)
	if err != nil {
		return err
	}

	h.sequences, err = h.getSequences(db)
	if err != nil {
		return err
	}

	h.constraints, err = h.getConstraints(db)
	if err != nil {
		return err
	}

	return nil
}

func (*cockroachDB) paramType() int {
	return paramTypeDollar
}

func (*cockroachDB) databaseName(q queryable) (string, error) {
	var dbName string
	err := q.QueryRow("SELECT current_database()").Scan(&dbName)
	return dbName, err
}

func (h *cockroachDB) tableNames(q queryable) ([]string, error) {
	var tables []string

	sql := `
	        SELECT pg_namespace.nspname || '.' || pg_class.relname
		FROM pg_class
		INNER JOIN pg_namespace ON pg_namespace.oid = pg_class.relnamespace
		WHERE pg_class.relkind = 'r'
		  AND pg_namespace.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND pg_namespace.nspname NOT LIKE 'pg_toast%'
		  AND pg_namespace.nspname NOT LIKE 'crdb_internal%'
		  AND pg_namespace.nspname NOT LIKE '\_timescaledb%';
	`
	rows, err := q.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (h *cockroachDB) getConstraints(q queryable) ([]crdbConstraint, error) {
	var constraints []crdbConstraint

	for _, table := range h.tables {
		sql := "SHOW CONSTRAINTS FROM %s;"
		rows, err := q.Query(fmt.Sprintf(sql, table))
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var constraint crdbConstraint
			if err = rows.Scan(&constraint.tableName,
				&constraint.constraintName,
				&constraint.constraintType,
				&constraint.details,
				&constraint.validated); err != nil {
				return nil, err
			}
			if constraint.constraintType == fkName {
				constraints = append(constraints, constraint)
			}
		}

		_ = rows.Close()
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}

	return constraints, nil
}

func (h *cockroachDB) getSequences(q queryable) ([]string, error) {
	const sql = `
		SELECT pg_namespace.nspname || '.' || pg_class.relname AS sequence_name
		FROM pg_class
		INNER JOIN pg_namespace ON pg_namespace.oid = pg_class.relnamespace
		WHERE pg_class.relkind = 'S'
		  AND pg_namespace.nspname NOT LIKE '\_timescaledb%'
	`

	rows, err := q.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sequences []string
	for rows.Next() {
		var sequence string
		if err = rows.Scan(&sequence); err != nil {
			return nil, err
		}
		sequences = append(sequences, sequence)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sequences, nil
}

func (h *cockroachDB) dropAndRecreateConstraints(db *sql.DB, loadFn loadFunction) (err error) {
	defer func() {
		// recreate constraints again after load
		var sql string
		for _, constraint := range h.constraints {
			sql += fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;",
				h.quoteKeyword(constraint.tableName),
				h.quoteKeyword(constraint.constraintName),
				constraint.details)
		}
		if _, err2 := db.Exec(sql); err2 != nil && err == nil {
			err = err2
		}
	}()

	var sql string
	for _, constraint := range h.constraints {
		sql += fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", h.quoteKeyword(constraint.tableName), h.quoteKeyword(constraint.constraintName))
	}
	if _, err := db.Exec(sql); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = loadFn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (h *cockroachDB) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	// ensure sequences being reset after load
	if !h.skipResetSequences {
		defer func() {
			if err2 := h.resetSequences(db); err2 != nil && err == nil {
				err = err2
			}
		}()
	}

	return h.dropAndRecreateConstraints(db, loadFn)

}

func (h *cockroachDB) resetSequences(db *sql.DB) error {
	resetSequencesTo := h.resetSequencesTo
	if resetSequencesTo == 0 {
		resetSequencesTo = 10000
	}

	for _, sequence := range h.sequences {
		_, err := db.Exec(fmt.Sprintf("SELECT SETVAL('%s', %d)", sequence, resetSequencesTo))
		if err != nil {
			return err
		}
	}
	return nil
}

func (*cockroachDB) quoteKeyword(s string) string {
	parts := strings.Split(s, ".")
	for i, p := range parts {
		parts[i] = fmt.Sprintf(`"%s"`, p)
	}
	return strings.Join(parts, ".")
}
