package testfixtures

import (
	"database/sql"
	"fmt"
	"strings"
)

// PostgreSQLHelper is the PG helper for this package
type PostgreSQLHelper struct {
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

func (h *PostgreSQLHelper) getSequences(db *sql.DB) ([]string, error) {
	sql := "SELECT relname FROM pg_class WHERE relkind = 'S'"
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	var sequences []string
	defer rows.Close()
	for rows.Next() {
		var sequence string
		rows.Scan(&sequence)
		sequences = append(sequences, sequence)
	}
	return sequences, nil
}

func (h *PostgreSQLHelper) disableTriggers(tx *sql.Tx) error {
	tables, err := h.getTables(tx)
	if err != nil {
		return err
	}
	sql := ""

	for _, table := range tables {
		sql = sql + fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER ALL;", table)
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
		sql = sql + fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER ALL;", table)
	}

	_, err = tx.Exec(sql)
	return err
}

func (h *PostgreSQLHelper) resetSequences(db *sql.DB) error {
	sequences, err := h.getSequences(db)
	if err != nil {
		return err
	}

	for _, sequence := range sequences {
		var max int
		table := strings.Replace(sequence, "_id_seq", "", 1)
		row := db.QueryRow(fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", table))
		err = row.Scan(&max)
		if err != nil {
			return err
		}

		if max == 0 {
			max = 1
		}
		_, err = db.Exec(fmt.Sprintf("SELECT SETVAL('%s', %d)", sequence, max))
		if err != nil {
			return err
		}
	}
	return nil
}

func (PostgreSQLHelper) beforeLoad(db *sql.DB) error {
	return nil
}

func (h *PostgreSQLHelper) afterLoad(db *sql.DB) error {
	return h.resetSequences(db)
}
