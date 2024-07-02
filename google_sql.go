package testfixtures

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/googleapis/go-sql-spanner"
)

type googleSQL struct {
	baseHelper

	cleanTableFn func(string) string
	constraints  []spannerConstraint
}

type spannerConstraint struct {
	tableName      string
	constraintName string
	definition     string
}

func (h *googleSQL) init(_ *sql.DB) error {
	if h.cleanTableFn == nil {
		h.cleanTableFn = func(tableName string) string {
			return fmt.Sprintf("DELETE FROM %s WHERE true;", tableName)
		}
	}

	return nil
}

func (*googleSQL) paramType() int {
	return paramTypeAtSign
}

func (*googleSQL) quoteKeyword(str string) string {
	return fmt.Sprintf(`%s`, str)
}

func (*googleSQL) databaseName(q queryable) (string, error) {
	return "testdb", nil
}

func (h *googleSQL) tableNames(q queryable) ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = '';
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

func (h *googleSQL) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	err = loadFn(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// splitter is a batchSplitter interface implementation. We need it for
// spanner because spanner doesn't support multi-statements.
func (*googleSQL) splitter() []byte {
	return []byte(";\n")
}

func (h *googleSQL) cleanTableQuery(tableName string) string {
	if h.cleanTableFn == nil {
		return h.baseHelper.cleanTableQuery(tableName)
	}

	return h.cleanTableFn(tableName)
}

func (h *googleSQL) dropAndRecreateConstraints(db *sql.DB, loadFn loadFunction) (err error) {
	defer func() {
		// Re-create constraints again after load
		var b strings.Builder
		for _, constraint := range h.constraints {
			b.WriteString(fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s %s;",
				h.quoteKeyword(constraint.tableName),
				h.quoteKeyword(constraint.constraintName),
				constraint.definition,
			))
		}
		if _, err2 := db.Exec(b.String()); err2 != nil && err == nil {
			err = err2
		}
	}()

	var b strings.Builder
	for _, constraint := range h.constraints {
		b.WriteString(fmt.Sprintf(
			"ALTER TABLE %s DROP CONSTRAINT %s;",
			h.quoteKeyword(constraint.tableName),
			h.quoteKeyword(constraint.constraintName),
		))
	}
	if _, err := db.Exec(b.String()); err != nil {
		return err
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
