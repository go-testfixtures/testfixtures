package testfixtures

import (
	"database/sql"
	"fmt"
	"strings"
)

// PostgreSQLHelper is the PG helper for this package
type PostgreSQLHelper struct {
}

// GetTables get all tables of the database
func (h *PostgreSQLHelper) GetTables(db *sql.DB) ([]string, error) {
	sql := `
SELECT table_name
FROM information_schema.tables
WHERE table_schema='public'
  AND table_type='BASE TABLE';
`
	rows, err := db.Query(sql)
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

// GetSequences get all sequences of the database
func (h *PostgreSQLHelper) GetSequences(db *sql.DB) ([]string, error) {
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

// DisableTriggers disable referential integrity triggers
func (h *PostgreSQLHelper) DisableTriggers(db *sql.DB) error {
	tables, err := h.GetTables(db)
	if err != nil {
		return err
	}
	sql := ""

	for _, table := range tables {
		sql = sql + fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER ALL;", table)
	}

	_, err = db.Exec(sql)
	return err
}

// EnableTriggers enable referential integrity triggers
func (h *PostgreSQLHelper) EnableTriggers(db *sql.DB) error {
	tables, err := h.GetTables(db)
	if err != nil {
		return err
	}
	sql := ""

	for _, table := range tables {
		sql = sql + fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER ALL;", table)
	}

	_, err = db.Exec(sql)
	return err
}

// ResetSequences resets the sequences of "id"s columns
// assumes the primery key is "id" and sequence is "<tablename>_id_seq",
// the default when using the SERIAL column type
func (h *PostgreSQLHelper) ResetSequences(db *sql.DB) error {
	sequences, err := h.GetSequences(db)
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
		_, err = db.Exec(fmt.Sprintf("SELECT SETVAL('%s', %d)", sequence, max))
		if err != nil {
			return err
		}
	}
	return nil
}

// BeforeLoad runs before the fixture load
// by now, does nothing
func (h *PostgreSQLHelper) BeforeLoad(db *sql.DB) error {
	return nil
}

// AfterLoad runs after the fixture load
func (h *PostgreSQLHelper) AfterLoad(db *sql.DB) error {
	return h.ResetSequences(db)
}
