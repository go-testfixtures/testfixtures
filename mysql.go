package testfixtures

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-testfixtures/testfixtures/v3/shared"
)

type mySQL struct {
	baseHelper

	skipResetSequences                bool
	resetSequencesTo                  int64
	allowMultipleStatementsInOneQuery bool

	tables         []string
	tablesChecksum map[string]int64
}

func (h *mySQL) setParamType(_ int) {}

func (h *mySQL) init(db *sql.DB) error {
	var err error
	h.tables, err = h.tableNames(db)
	if err != nil {
		return err
	}

	return nil
}

func (*mySQL) paramType() int {
	return paramTypeQuestion
}

func (*mySQL) quoteKeyword(str string) string {
	return fmt.Sprintf("`%s`", str)
}

func (*mySQL) databaseName(q shared.Queryable) (string, error) {
	var dbName string
	err := q.QueryRow("SELECT DATABASE()").Scan(&dbName)
	return dbName, err
}

func (h *mySQL) tableNames(q shared.Queryable) ([]string, error) {
	const query = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE';
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

func (h *mySQL) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	if !h.skipResetSequences {
		defer func() {
			if err2 := h.resetSequences(db); err2 != nil && err == nil {
				err = err2
			}
		}()
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}

	err = loadFn(tx)
	_, err2 := tx.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	return tx.Commit()
}

func (h *mySQL) resetSequences(db *sql.DB) error {
	if len(h.tables) == 0 {
		return nil
	}

	resetSequencesTo := h.resetSequencesTo
	if resetSequencesTo == 0 {
		resetSequencesTo = 10000
	}

	if h.allowMultipleStatementsInOneQuery {
		return h.resetSequencesInOneQuery(db, resetSequencesTo)
	}
	return h.resetSequencesInMultipleQueries(db, resetSequencesTo)

}

func (h *mySQL) resetSequencesInOneQuery(db *sql.DB, resetSequencesTo int64) error {
	b := strings.Builder{}
	for _, t := range h.tables {
		b.WriteString(h.makeResetSequenceQuery(t, resetSequencesTo))
	}
	_, err := db.Exec(b.String())
	return err
}

func (h *mySQL) resetSequencesInMultipleQueries(db *sql.DB, resetSequencesTo int64) error {
	for _, t := range h.tables {
		_, err := db.Exec(h.makeResetSequenceQuery(t, resetSequencesTo))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *mySQL) makeResetSequenceQuery(tableName string, resetSequencesTo int64) string {
	return fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d;", h.quoteKeyword(tableName), resetSequencesTo)
}

func (h *mySQL) isTableModified(q shared.Queryable, tableName string) (bool, error) {
	oldChecksum, found := h.tablesChecksum[tableName]
	if !found {
		return true, nil
	}

	checksum, err := h.getChecksum(q, tableName)
	if err != nil {
		return true, err
	}
	return checksum != oldChecksum, nil
}

func (h *mySQL) computeTablesChecksum(q shared.Queryable) error {
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

func (h *mySQL) getChecksum(q shared.Queryable, tableName string) (int64, error) {
	query := fmt.Sprintf("CHECKSUM TABLE %s", h.quoteKeyword(tableName))
	var (
		table    string
		checksum sql.NullInt64
	)
	if err := q.QueryRow(query).Scan(&table, &checksum); err != nil {
		return 0, err
	}
	if !checksum.Valid {
		return 0, fmt.Errorf("testfixtures: table %s does not exist", tableName)
	}
	return checksum.Int64, nil
}
