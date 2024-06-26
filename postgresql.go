package testfixtures

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type postgreSQL struct {
	baseHelper

	useAlterConstraint bool
	useDropConstraint  bool
	skipResetSequences bool
	resetSequencesTo   int64

	tables                   []string
	sequences                []string
	nonDeferrableConstraints []pgConstraint
	constraints              []pgConstraint
	tablesChecksum           map[string]string

	version                      int
	tablesHasIdentityColumnMutex sync.Mutex
	tablesHasIdentityColumn      map[string]bool
}

type pgConstraint struct {
	tableName      string
	constraintName string
	definition     string
}

func (h *postgreSQL) init(db *sql.DB) error {
	var err error

	h.tables, err = h.tableNames(db)
	if err != nil {
		return err
	}

	h.sequences, err = h.getSequences(db)
	if err != nil {
		return err
	}

	h.nonDeferrableConstraints, err = h.getNonDeferrableConstraints(db)
	if err != nil {
		return err
	}

	h.constraints, err = h.getConstraints(db)
	if err != nil {
		return err
	}

	h.version, err = h.getMajorVersion(db)
	if err != nil {
		return err
	}

	h.tablesHasIdentityColumn = make(map[string]bool)

	return nil
}

func (*postgreSQL) paramType() int {
	return paramTypeDollar
}

func (*postgreSQL) databaseName(q queryable) (string, error) {
	var dbName string
	err := q.QueryRow("SELECT current_database()").Scan(&dbName)
	return dbName, err
}

func (h *postgreSQL) tableNames(q queryable) ([]string, error) {
	var tables []string

	const sql = `
	        SELECT pg_namespace.nspname || '.' || pg_class.relname
		FROM pg_class
		INNER JOIN pg_namespace ON pg_namespace.oid = pg_class.relnamespace
		WHERE pg_class.relkind = 'r'
		  AND pg_namespace.nspname NOT IN ('pg_catalog', 'information_schema', 'crdb_internal')
		  AND pg_namespace.nspname NOT LIKE 'pg_toast%'
		  AND pg_namespace.nspname NOT LIKE '\_timescaledb%';
	`
	rows, err := q.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (h *postgreSQL) getSequences(q queryable) ([]string, error) {
	const sql = `
		SELECT pg_namespace.nspname || '.' || pg_class.relname AS sequence_name
		FROM pg_class
		INNER JOIN pg_namespace ON pg_namespace.oid = pg_class.relnamespace
		WHERE pg_class.relkind = 'S'
		  AND pg_namespace.nspname NOT LIKE '\_timescaledb%'
	`

	rows, err := q.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sequences []string
	for rows.Next() {
		var sequence string
		if err = rows.Scan(&sequence); err != nil {
			return nil, err
		}
		sequences = append(sequences, sequence)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sequences, nil
}

func (*postgreSQL) getNonDeferrableConstraints(q queryable) ([]pgConstraint, error) {
	var constraints []pgConstraint

	const sql = `
		SELECT table_schema || '.' || table_name, constraint_name
		FROM information_schema.table_constraints
		WHERE constraint_type = 'FOREIGN KEY'
		  AND is_deferrable = 'NO'
		  AND table_schema <> 'crdb_internal'
		  AND table_schema NOT LIKE '\_timescaledb%'
  	`
	rows, err := q.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var constraint pgConstraint
		if err = rows.Scan(&constraint.tableName, &constraint.constraintName); err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return constraints, nil
}

func (h *postgreSQL) getConstraints(q queryable) ([]pgConstraint, error) {
	var constraints []pgConstraint

	const sql = `
		SELECT conrelid::regclass AS table_from, conname, pg_get_constraintdef(pg_constraint.oid)
		FROM pg_constraint
		INNER JOIN pg_namespace ON pg_namespace.oid = pg_constraint.connamespace
		WHERE contype = 'f'
		  AND pg_namespace.nspname NOT IN ('pg_catalog', 'information_schema', 'crdb_internal')
		  AND pg_namespace.nspname NOT LIKE 'pg_toast%'
		  AND pg_namespace.nspname NOT LIKE '\_timescaledb%';
		`
	rows, err := q.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var constraint pgConstraint
		if err = rows.Scan(
			&constraint.tableName,
			&constraint.constraintName,
			&constraint.definition,
		); err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return constraints, nil
}

func (h *postgreSQL) dropAndRecreateConstraints(db *sql.DB, loadFn loadFunction) (err error) {
	defer func() {
		// Re-create constraints again after load
		var b strings.Builder
		for _, constraint := range h.constraints {
			b.WriteString(fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s %s;",
				h.quoteKeyword(constraint.tableName),
				h.quoteKeyword(constraint.constraintName),
				constraint.definition,
			))
		}
		if _, err2 := db.Exec(b.String()); err2 != nil && err == nil {
			err = err2
		}
	}()

	var b strings.Builder
	for _, constraint := range h.constraints {
		b.WriteString(fmt.Sprintf(
			"ALTER TABLE %s DROP CONSTRAINT %s;",
			h.quoteKeyword(constraint.tableName),
			h.quoteKeyword(constraint.constraintName),
		))
	}
	if _, err := db.Exec(b.String()); err != nil {
		return err
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

func (h *postgreSQL) disableTriggers(db *sql.DB, loadFn loadFunction) (err error) {
	defer func() {
		var b strings.Builder
		for _, table := range h.tables {
			b.WriteString(fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER ALL;", h.quoteKeyword(table)))
		}
		if _, err2 := db.Exec(b.String()); err2 != nil && err == nil {
			err = err2
		}
	}()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	var b strings.Builder
	for _, table := range h.tables {
		b.WriteString(fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER ALL;", h.quoteKeyword(table)))
	}
	if _, err = tx.Exec(b.String()); err != nil {
		return err
	}

	if err = loadFn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (h *postgreSQL) makeConstraintsDeferrable(db *sql.DB, loadFn loadFunction) (err error) {
	defer func() {
		// ensure constraint being not deferrable again after load
		var b strings.Builder
		for _, constraint := range h.nonDeferrableConstraints {
			b.WriteString(fmt.Sprintf("ALTER TABLE %s ALTER CONSTRAINT %s NOT DEFERRABLE;", h.quoteKeyword(constraint.tableName), h.quoteKeyword(constraint.constraintName)))
		}
		if _, err2 := db.Exec(b.String()); err2 != nil && err == nil {
			err = err2
		}
	}()

	var b strings.Builder
	for _, constraint := range h.nonDeferrableConstraints {
		b.WriteString(fmt.Sprintf("ALTER TABLE %s ALTER CONSTRAINT %s DEFERRABLE;", h.quoteKeyword(constraint.tableName), h.quoteKeyword(constraint.constraintName)))
	}
	if _, err := db.Exec(b.String()); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED"); err != nil {
		return err
	}

	if err = loadFn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (h *postgreSQL) disableReferentialIntegrity(db *sql.DB, loadFn loadFunction) (err error) {
	// ensure sequences being reset after load
	if !h.skipResetSequences {
		defer func() {
			if err2 := h.resetSequences(db); err2 != nil && err == nil {
				err = err2
			}
		}()
	}

	if h.useDropConstraint {
		return h.dropAndRecreateConstraints(db, loadFn)
	}
	if h.useAlterConstraint {
		return h.makeConstraintsDeferrable(db, loadFn)
	}
	return h.disableTriggers(db, loadFn)
}

func (h *postgreSQL) resetSequences(db *sql.DB) error {
	resetSequencesTo := h.resetSequencesTo
	if resetSequencesTo == 0 {
		resetSequencesTo = 10000
	}

	for _, sequence := range h.sequences {
		_, err := db.Exec(fmt.Sprintf("SELECT SETVAL('%s', %d)", sequence, resetSequencesTo))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *postgreSQL) isTableModified(q queryable, tableName string) (bool, error) {
	checksum, err := h.getChecksum(q, tableName)
	if err != nil {
		return false, err
	}

	oldChecksum := h.tablesChecksum[tableName]

	return oldChecksum == "" || checksum != oldChecksum, nil
}

func (h *postgreSQL) computeTablesChecksum(q queryable) error {
	if h.tablesChecksum != nil {
		return nil
	}

	h.tablesChecksum = make(map[string]string, len(h.tables))
	for _, t := range h.tables {
		checksum, err := h.getChecksum(q, t)
		if err != nil {
			return err
		}
		h.tablesChecksum[t] = checksum
	}
	return nil
}

func (h *postgreSQL) getChecksum(q queryable, tableName string) (string, error) {
	sqlStr := fmt.Sprintf(`
			SELECT md5(CAST((json_agg(t.*)) AS TEXT))
			FROM %s AS t
		`,
		h.quoteKeyword(tableName),
	)

	var checksum sql.NullString
	if err := q.QueryRow(sqlStr).Scan(&checksum); err != nil {
		return "", err
	}
	return checksum.String, nil
}

func (*postgreSQL) quoteKeyword(s string) string {
	parts := strings.Split(s, ".")
	for i, p := range parts {
		parts[i] = fmt.Sprintf(`"%s"`, p)
	}
	return strings.Join(parts, ".")
}

func (h *postgreSQL) buildInsertSQL(q queryable, tableName string, columns, values []string) (string, error) {
	if h.version >= 10 {
		ok, err := h.tableHasIdentityColumn(q, tableName)
		if err != nil {
			return "", err
		}
		if ok {
			return fmt.Sprintf(
				"INSERT INTO %s (%s) OVERRIDING SYSTEM VALUE VALUES (%s)",
				tableName,
				strings.Join(columns, ", "),
				strings.Join(values, ", "),
			), nil
		}
	}

	return h.baseHelper.buildInsertSQL(q, tableName, columns, values)
}

func (h *postgreSQL) tableHasIdentityColumn(q queryable, tableName string) (bool, error) {
	defer h.tablesHasIdentityColumnMutex.Unlock()
	h.tablesHasIdentityColumnMutex.Lock()

	hasIdentityColumn, exists := h.tablesHasIdentityColumn[tableName]
	if exists {
		return hasIdentityColumn, nil
	}

	parts := strings.Split(tableName, ".")
	tableName = parts[0][1 : len(parts[0])-1]
	if len(parts) > 1 {
		tableName = parts[1][1 : len(parts[1])-1]
	}

	query := `
		SELECT COUNT(*) AS count
		FROM information_schema.columns
		WHERE table_name = $1
		  AND is_identity = 'YES'
	`
	var count int
	if err := q.QueryRow(query, tableName).Scan(&count); err != nil {
		return false, err
	}

	h.tablesHasIdentityColumn[tableName] = count > 0
	return h.tablesHasIdentityColumn[tableName], nil
}

func (h *postgreSQL) getMajorVersion(q queryable) (int, error) {
	var version string
	err := q.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		return 0, err
	}

	return h.parseMajorVersion(version)
}

func (*postgreSQL) parseMajorVersion(version string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	versionNumbers := re.FindAllString(version, -1)
	if len(versionNumbers) > 0 {
		majorVersion, err := strconv.Atoi(versionNumbers[0])
		if err != nil {
			return 0, err
		}
		return majorVersion, nil
	}

	return 0, fmt.Errorf("testfixtures: could not parse major version from: %s", version)
}
