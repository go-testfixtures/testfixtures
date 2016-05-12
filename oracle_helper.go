package testfixtures

import (
	"database/sql"
	"fmt"
	"strings"
)

// OracleHelper is the Oracle database helper for this package
type OracleHelper struct{}

type oracleConstraint struct {
	tableName      string
	constraintName string
}

func (OracleHelper) paramType() int {
	return paramTypeColon
}

func (OracleHelper) quoteKeyword(str string) string {
	return fmt.Sprintf("\"%s\"", strings.ToUpper(str))
}

func (OracleHelper) databaseName(db *sql.DB) (dbName string) {
	db.QueryRow("SELECT user FROM DUAL").Scan(&dbName)
	return
}

func (OracleHelper) whileInsertOnTable(tx *sql.Tx, tableName string, fn func() error) error {
	return fn()
}

func (OracleHelper) getEnabledContraints(db *sql.DB) ([]oracleConstraint, error) {
	constraints := make([]oracleConstraint, 0)
	rows, err := db.Query(`
        SELECT table_name, constraint_name
        FROM user_constraints
        WHERE constraint_type = 'R'
          AND status = 'ENABLED'
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var constraint oracleConstraint
		rows.Scan(&constraint.tableName, &constraint.constraintName)
		constraints = append(constraints, constraint)
	}
	return constraints, nil
}

func (OracleHelper) getSequences(db *sql.DB) ([]string, error) {
	sequences := make([]string, 0)
	rows, err := db.Query("SELECT sequence_name FROM user_sequences")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var sequence string
		rows.Scan(&sequence)
		sequences = append(sequences, sequence)
	}
	return sequences, nil
}

func (h *OracleHelper) resetSequences(db *sql.DB) error {
	sequences, err := h.getSequences(db)
	if err != nil {
		return err
	}

	for _, sequence := range sequences {
		_, err := db.Exec(fmt.Sprintf("DROP SEQUENCE %s", h.quoteKeyword(sequence)))
		if err != nil {
			return err
		}
		_, err = db.Exec(fmt.Sprintf("CREATE SEQUENCE %s START WITH %d", h.quoteKeyword(sequence), resetSequencesTo))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *OracleHelper) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) error {
	constraints, err := h.getEnabledContraints(db)
	if err != nil {
		return err
	}

	// re-enable after load
	defer func() {
		for _, c := range constraints {
			db.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE CONSTRAINT %s", h.quoteKeyword(c.tableName), h.quoteKeyword(c.constraintName)))
		}
	}()

	// disable foreign keys
	for _, c := range constraints {
		_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s DISABLE CONSTRAINT %s", h.quoteKeyword(c.tableName), h.quoteKeyword(c.constraintName)))
		if err != nil {
			return err
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = loadFn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return h.resetSequences(db)
}
