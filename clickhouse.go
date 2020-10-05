package testfixtures

import (
	"database/sql"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
)

type clickhouse struct {
	baseHelper
	tables         []string
	tablesChecksum map[string]int64
}

func (h *clickhouse) init(db *sql.DB) error {
	var err error
	h.tables, err = h.tableNames(db)
	if err != nil {
		return err
	}

	return nil
}

func (h *clickhouse) cleanTable(tx *sql.Tx, tableName string) error {
	if _, err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s", tableName)); err != nil {
		return fmt.Errorf(`testfixtures: could not clean table "%s": %w`, tableName, err)
	}

	return nil
}

func (*clickhouse) paramType() int {
	return paramTypeQuestion
}

func (*clickhouse) quoteKeyword(str string) string {
	return fmt.Sprintf("`%s`", str)
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
		WHERE database = ?;
	`
	dbName, err := h.databaseName(q)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(query, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	defer tx.Rollback()

	err = loadFn(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (h *clickhouse) isTableModified(q queryable, tableName string) (bool, error) {
	checksum, err := h.getChecksum(q, tableName)
	if err != nil {
		return true, err
	}

	oldChecksum := h.tablesChecksum[tableName]

	return oldChecksum == 0 || checksum != oldChecksum, nil
}

func (h *clickhouse) afterLoad(q queryable) error {
	if h.tablesChecksum != nil {
		return nil
	}

	h.tablesChecksum = make(map[string]int64, len(h.tables))
	for _, t := range h.tables {
		checksum, err := h.getChecksum(q, t)
		if err != nil {
			return err
		}
		h.tablesChecksum[t] = checksum
	}
	return nil
}

func (h *clickhouse) getChecksum(q queryable, tableName string) (int64, error) {
	// This is an equivalent query to get the checksum of the content of the table
	// We divide by 2 because it returns an uint64 instead of an int64
	query := fmt.Sprintf("SELECT toInt64(groupBitXor(cityHash64(*)) / 2) FROM %s", h.quoteKeyword(tableName))
	var (
		checksum sql.NullInt64
	)

	if err := q.QueryRow(query).Scan(&checksum); err != nil {
		return 0, err
	}

	return checksum.Int64, nil
}

// splitter is a batchSplitter interface implementation. We need it for
// ClickHouseDB because clickhouse doesn't support multi-statements.
func (*clickhouse) splitter() []byte {
	return []byte(";\n")
}
