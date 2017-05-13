package testfixtures

import (
	"database/sql"
	"fmt"
)

// MySQL is the MySQL helper for this package
type MySQL struct {
	baseHelper
}

func (*MySQL) paramType() int {
	return paramTypeQuestion
}

func (*MySQL) quoteKeyword(str string) string {
	return fmt.Sprintf("`%s`", str)
}

func (*MySQL) databaseName(db *sql.DB) (dbName string) {
	db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	return
}

func (h *MySQL) tableNames(db *sql.DB) ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema=?;
	`
	rows, err := db.Query(query, h.databaseName(db))
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

func (h *MySQL) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	// re-enable after load
	defer func() {
		if _, err2 := db.Exec("SET FOREIGN_KEY_CHECKS = 1"); err2 != nil && err == nil {
			err = err2
		}
	}()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}

	if err = loadFn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
