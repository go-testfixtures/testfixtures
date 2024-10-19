package testfixtures

import (
	"database/sql"
	"fmt"
)

type clickhouse struct {
	baseHelper

	cleanTableFn func(string) string
}

func (h *clickhouse) init(_ *sql.DB) error {
	if h.cleanTableFn == nil {
		h.cleanTableFn = func(tableName string) string {
			return fmt.Sprintf("TRUNCATE TABLE %s", tableName)
		}
	}

	return nil
}

func (*clickhouse) paramType() int {
	return paramTypeDollar
}

func (*clickhouse) databaseName(q queryable) (string, error) {
	var dbName string
	err := q.QueryRow("SELECT DATABASE()").Scan(&dbName)
	return dbName, err
}

func (h *clickhouse) tableNames(q queryable) ([]string, error) {
	query := `
		SELECT name
		FROM system.tables
		WHERE database = $1;
	`
	dbName, err := h.databaseName(q)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(query, dbName)
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

func (h *clickhouse) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
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
// ClickHouseDB because clickhouse doesn't support multi-statements.
func (*clickhouse) splitter() []byte {
	return []byte(";\n")
}

func (h *clickhouse) cleanTableQuery(tableName string) string {
	if h.cleanTableFn == nil {
		return h.baseHelper.cleanTableQuery(tableName)
	}

	return h.cleanTableFn(tableName)
}
