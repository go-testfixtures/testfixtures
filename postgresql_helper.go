package testfixtures

import (
    "fmt"
    "database/sql"
)

type PostgreSQLHelper struct {

}

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

    tables := make([]string, 0)
    defer rows.Close()
    for rows.Next() {
        var table string
        rows.Scan(&table)
        tables = append(tables, table)
    }
    return tables, nil
}

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
