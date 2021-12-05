package testfixtures

import (
	"database/sql"
	"fmt"
)

type clickhouse struct {
	baseHelper

	tables         []string
	tablesChecksum map[string]int64
}

func (c *clickhouse) init(db *sql.DB) error {
	var err error

	if c.tables, err = c.tableNames(db); err != nil {
		return err
	}

	return nil
}

func (c *clickhouse) tableNames(q queryable) ([]string, error) {
	const query = `
		SELECT name
		FROM system.tables
		WHERE database = ?
	`

	dbName, err := c.databaseName(q)
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

func (*clickhouse) databaseName(q queryable) (string, error) {
	var dbName string

	err := q.QueryRow("SELECT DATABASE()").Scan(&dbName)

	return dbName, err
}

func (*clickhouse) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
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

func (*clickhouse) paramType() int {
	return paramTypeQuestion
}

func (c *clickhouse) isTableModified(q queryable, tableName string) (bool, error) {
	checksum, err := c.getChecksum(q, tableName)
	if err != nil {
		return true, err
	}

	oldChecksum := c.tablesChecksum[tableName]

	return oldChecksum == 0 || checksum != oldChecksum, nil
}

func (c *clickhouse) afterLoad(q queryable) error {
	if c.tablesChecksum != nil {
		return nil
	}

	c.tablesChecksum = make(map[string]int64, len(c.tables))

	for _, t := range c.tables {
		checksum, err := c.getChecksum(q, t)
		if err != nil {
			return err
		}
		c.tablesChecksum[t] = checksum
	}

	return nil
}

func (c *clickhouse) getChecksum(q queryable, tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT toInt64(groupBitXor(cityHash64(*)) / 2) FROM %s", c.quoteKeyword(tableName))
	var checksum sql.NullInt64

	if err := q.QueryRow(query).Scan(&checksum); err != nil {
		return 0, err
	}

	if !checksum.Valid {
		return 0, fmt.Errorf("testfixtures: table %s does not exist", tableName)
	}

	return checksum.Int64, nil
}

// splitter is a batchSplitter interface implementation.
func (*clickhouse) splitter() []byte {
	return []byte(";\n")
}

func (*clickhouse) cleanTableQuery(tableName string) string {
	return fmt.Sprintf("TRUNCATE TABLE %s", tableName)
}

func clickhouseWileInsertOnTableFn(db *sql.DB, file *fixtureFile) func() error {
	return func() error {
		for j, i := range file.insertSQLs {
			tx, err := db.Begin()
			if err != nil {
				return err
			}

			stmt, err := tx.Prepare(i.sql)
			if err != nil {
				return err
			}

			if _, err := stmt.Exec(i.params...); err != nil {
				if err = tx.Rollback(); err != nil {
					return err
				}

				return &InsertError{
					Err:    err,
					File:   file.fileName,
					Index:  j,
					SQL:    i.sql,
					Params: i.params,
				}
			}

			if err = tx.Commit(); err != nil {
				return err
			}
		}

		return nil
	}
}
