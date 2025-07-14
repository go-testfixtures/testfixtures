package testfixtures

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-testfixtures/testfixtures/v3/shared"
	"golang.org/x/sync/errgroup"
)

type spanner struct {
	baseHelper

	cleanTableFn          func(string) string
	constraints           map[string][]shared.SpannerConstraint
	tablesWithJSONColumns map[string]map[string]bool
}

func (h *spanner) init(db *sql.DB) error {
	if h.cleanTableFn == nil {
		h.cleanTableFn = func(tableName string) string {
			return fmt.Sprintf("DELETE FROM %s WHERE true;", tableName)
		}
	}

	var grp errgroup.Group
	grp.Go(func() error {
		var err error
		h.constraints, err = shared.GetConstraints(db)
		return err
	})
	grp.Go(func() error {
		var err error
		h.tablesWithJSONColumns, err = h.buildTableJSONColumns(db)
		return err
	})
	if err := grp.Wait(); err != nil {
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

func (*spanner) databaseName(q shared.Queryable) (string, error) {
	return "", errors.New("could not determine database name. Please skip the test database check")
}

func (h *spanner) tableNames(q shared.Queryable) ([]string, error) {
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

func (h *spanner) cleanTableQuery(tableName string) string {
	if h.cleanTableFn == nil {
		return h.baseHelper.cleanTableQuery(tableName)
	}

	return h.cleanTableFn(tableName)
}

func (h *spanner) dropAndRecreateConstraints(db *sql.DB, loadFn loadFunction) (err error) {
	defer func() {
		// Re-create constraints again after load
		for key := range h.constraints {
			var lengthConstraints = len(h.constraints[key])
			var orderedConstraints = make([]shared.SpannerConstraint, lengthConstraints)

			for _, constraint := range h.constraints[key] {
				orderedConstraints[constraint.Position-1] = constraint
			}

			var columnName = orderedConstraints[0].ColumnName
			for i := 1; i < lengthConstraints; i++ {
				columnName = strings.Join([]string{columnName, orderedConstraints[i].ColumnName}, ", ")
			}

			var referencedColumn = orderedConstraints[0].ReferencedColumn
			for i := 1; i < lengthConstraints; i++ {
				referencedColumn = strings.Join([]string{referencedColumn, orderedConstraints[i].ReferencedColumn}, ", ")
			}

			cmd := fmt.Sprintf(
				`ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)`,
				orderedConstraints[0].TableName,
				orderedConstraints[0].ConstraintName,
				columnName,
				orderedConstraints[0].ReferencedTable,
				referencedColumn,
			)

			if _, err2 := db.Exec(cmd); err2 != nil && err == nil {
				err = err2
			}
		}
	}()

	for key := range h.constraints {
		constraints := h.constraints[key]
		cmd := fmt.Sprintf(
			`ALTER TABLE %s DROP CONSTRAINT %s`,
			constraints[0].TableName,
			constraints[0].ConstraintName,
		)
		if _, err := db.Exec(cmd); err != nil {
			fmt.Println("error dropping constraint", err)
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

func (h *spanner) buildTableJSONColumns(q shared.Queryable) (map[string]map[string]bool, error) {
	const query = `
		SELECT table_name, column_name, spanner_type
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE table_schema = ''
		  AND spanner_type = 'JSON'
	`

	rows, err := q.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	tablesWithJSONColumns := make(map[string]map[string]bool)
	for rows.Next() {
		var tableName, columnName, spannerType string
		if err = rows.Scan(&tableName, &columnName, &spannerType); err != nil {
			return nil, err
		}

		if tablesWithJSONColumns[tableName] == nil {
			tablesWithJSONColumns[tableName] = make(map[string]bool)
		}
		tablesWithJSONColumns[tableName][columnName] = true
	}

	return tablesWithJSONColumns, rows.Err()
}

func (h *spanner) buildInsertSQL(q shared.Queryable, tableName string, columns, values []string) (string, error) {
	if jsonColumns, tableExists := h.tablesWithJSONColumns[tableName]; tableExists {
		for i, column := range columns {
			if jsonColumns[column] {
				values[i] = fmt.Sprintf("PARSE_JSON(%s)", values[i])
			}
		}

		return fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(values, ", "),
		), nil
	}

	return h.baseHelper.buildInsertSQL(q, tableName, columns, values)
}
