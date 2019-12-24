package testfixtures

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// Loader is the responsible to loading fixtures.
type Loader struct {
	db            *sql.DB
	helper        helper
	fixturesFiles []*fixtureFile

	skipTestDatabaseCheck bool
}

type fixtureFile struct {
	path       string
	fileName   string
	content    []byte
	insertSQLs []insertSQL
}

type insertSQL struct {
	sql    string
	params []interface{}
}

var (
	testDatabaseRegexp = regexp.MustCompile("(?i)test")

	errDatabaseIsRequired = fmt.Errorf("testfixtures: database is required")
	errDialectIsRequired  = fmt.Errorf("testfixtures: dialect is required")
)

// New instantiates a new Loader instance. The "Database" and "Driver"
// options are required.
func New(options ...func(*Loader) error) (*Loader, error) {
	l := &Loader{}

	for _, option := range options {
		if err := option(l); err != nil {
			return nil, err
		}
	}

	if l.db == nil {
		return nil, errDatabaseIsRequired
	}
	if l.helper == nil {
		return nil, errDialectIsRequired
	}

	if err := l.helper.init(l.db); err != nil {
		return nil, err
	}
	if err := l.buildInsertSQLs(); err != nil {
		return nil, err
	}

	return l, nil
}

// Database sets an existing sql.DB instant to Loader.
func Database(db *sql.DB) func(*Loader) error {
	return func(l *Loader) error {
		l.db = db
		return nil
	}
}

// Dialect informs Loader about which database dialect you're using.
//
// Possible options are "postgresql", "timescaledb", "mysql", "mariadb",
// "sqlite" and "sqlserver".
func Dialect(dialect string) func(*Loader) error {
	return func(l *Loader) error {
		h, err := helperForDialect(dialect)
		if err != nil {
			return err
		}
		l.helper = h
		return nil
	}
}

func helperForDialect(dialect string) (helper, error) {
	switch dialect {
	case "postgres", "postgresql", "timescaledb":
		return &postgreSQL{}, nil
	case "mysql", "mariadb":
		return &mySQL{}, nil
	case "sqlite", "sqlite3":
		return &sqlite{}, nil
	case "mssql", "sqlserver":
		return &sqlserver{}, nil
	default:
		return nil, fmt.Errorf(`testfixtures: unrecognized dialect "%s"`, dialect)
	}
}

// UseAlterConstraint If true, the contraint disabling will do
// using ALTER CONTRAINT sintax, only allowed in PG >= 9.4.
// If false, the constraint disabling will use DISABLE TRIGGER ALL,
// which requires SUPERUSER privileges.
//
// Only valid for PostgreSQL. Returns an error otherwise.
func UseAlterConstraint() func(*Loader) error {
	return func(l *Loader) error {
		pgHelper, ok := l.helper.(*postgreSQL)
		if !ok {
			return fmt.Errorf("testfixtures: UseAlterConstraint is only valid for PostgreSQL databases")
		}
		pgHelper.useAlterConstraint = true
		return nil
	}
}

// SkipResetSequences prevents Loader from reseting sequences after loading
// fixtures.
//
// Only valid for PostgreSQL. Returns an error otherwise.
func SkipResetSequences() func(*Loader) error {
	return func(l *Loader) error {
		pgHelper, ok := l.helper.(*postgreSQL)
		if !ok {
			return fmt.Errorf("testfixtures: SkipResetSequences is only valid for PostgreSQL databases")
		}
		pgHelper.skipResetSequences = true
		return nil
	}
}

// ResetSequencesTo sets the value the sequences will be reset to.
// Defaults to 10000.
//
// Only valid for PostgreSQL. Returns an error otherwise.
func ResetSequencesTo(value int64) func(*Loader) error {
	return func(l *Loader) error {
		pgHelper, ok := l.helper.(*postgreSQL)
		if !ok {
			return fmt.Errorf("testfixtures: ResetSequencesTo is only valid for PostgreSQL databases")
		}
		pgHelper.resetSequencesTo = value
		return nil
	}
}

// DangerousSkipTestDatabaseCheck will make Loader not check if the database
// name contains "test". Use with caution!
func DangerousSkipTestDatabaseCheck() func(*Loader) error {
	return func(l *Loader) error {
		l.skipTestDatabaseCheck = true
		return nil
	}
}

// Directory informs Loader to load YAML files from a given directory.
func Directory(dir string) func(*Loader) error {
	return func(l *Loader) error {
		fixtures, err := fixturesFromDir(dir)
		if err != nil {
			return err
		}
		l.fixturesFiles = fixtures
		return nil
	}
}

// Files informs Loader to load a given set of YAML files.
func Files(files ...string) func(*Loader) error {
	return func(l *Loader) error {
		fixtures, err := fixturesFromFiles(files...)
		if err != nil {
			return err
		}
		l.fixturesFiles = fixtures
		return nil
	}
}

// EnsureTestDatabase returns an error if the database name does not contains
// "test".
func (l *Loader) EnsureTestDatabase() error {
	dbName, err := l.helper.databaseName(l.db)
	if err != nil {
		return err
	}
	if !testDatabaseRegexp.MatchString(dbName) {
		return fmt.Errorf(`testfixtures: database "%s" does not appear to be a test database`, dbName)
	}
	return nil
}

// Load wipes and after load all fixtures in the database.
//     if err := fixtures.Load(); err != nil {
//         log.Fatal(err)
//     }
func (l *Loader) Load() error {
	if !l.skipTestDatabaseCheck {
		if err := l.EnsureTestDatabase(); err != nil {
			return err
		}
	}

	err := l.helper.disableReferentialIntegrity(l.db, func(tx *sql.Tx) error {
		for _, file := range l.fixturesFiles {
			modified, err := l.helper.isTableModified(tx, file.fileNameWithoutExtension())
			if err != nil {
				return err
			}
			if !modified {
				continue
			}
			if err := file.delete(tx, l.helper); err != nil {
				return err
			}

			err = l.helper.whileInsertOnTable(tx, file.fileNameWithoutExtension(), func() error {
				for j, i := range file.insertSQLs {
					if _, err := tx.Exec(i.sql, i.params...); err != nil {
						return &InsertError{
							Err:    err,
							File:   file.fileName,
							Index:  j,
							SQL:    i.sql,
							Params: i.params,
						}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return l.helper.afterLoad(l.db)
}

// InsertError will be returned if any error happens on database while
// inserting the record.
type InsertError struct {
	Err    error
	File   string
	Index  int
	SQL    string
	Params []interface{}
}

func (e *InsertError) Error() string {
	return fmt.Sprintf(
		"testfixtures: error inserting record: %v, on file: %s, index: %d, sql: %s, params: %v",
		e.Err,
		e.File,
		e.Index,
		e.SQL,
		e.Params,
	)
}

func (l *Loader) buildInsertSQLs() error {
	for _, f := range l.fixturesFiles {
		var records interface{}
		if err := yaml.Unmarshal(f.content, &records); err != nil {
			return fmt.Errorf("testfixtures: could not unmarshal YAML: %w", err)
		}

		switch records := records.(type) {
		case []interface{}:
			for _, record := range records {
				recordMap, ok := record.(map[interface{}]interface{})
				if !ok {
					return fmt.Errorf("testfixtures: could not cast record: not a map[interface{}]interface{}")
				}

				sql, values, err := f.buildInsertSQL(l.helper, recordMap)
				if err != nil {
					return err
				}

				f.insertSQLs = append(f.insertSQLs, insertSQL{sql, values})
			}
		case map[interface{}]interface{}:
			for _, record := range records {
				recordMap, ok := record.(map[interface{}]interface{})
				if !ok {
					return fmt.Errorf("testfixtures: could not cast record: not a map[interface{}]interface{}")
				}

				sql, values, err := f.buildInsertSQL(l.helper, recordMap)
				if err != nil {
					return err
				}

				f.insertSQLs = append(f.insertSQLs, insertSQL{sql, values})
			}
		default:
			return fmt.Errorf("testfixtures: fixture is not a slice or map")
		}
	}

	return nil
}

func (f *fixtureFile) fileNameWithoutExtension() string {
	return strings.Replace(f.fileName, filepath.Ext(f.fileName), "", 1)
}

func (f *fixtureFile) delete(tx *sql.Tx, h helper) error {
	if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", h.quoteKeyword(f.fileNameWithoutExtension()))); err != nil {
		return fmt.Errorf(`testfixtures: could not clean table "%s": %w`, f.fileNameWithoutExtension(), err)
	}
	return nil
}

func (f *fixtureFile) buildInsertSQL(h helper, record map[interface{}]interface{}) (sqlStr string, values []interface{}, err error) {
	var (
		sqlColumns []string
		sqlValues  []string
		i          = 1
	)
	for key, value := range record {
		keyStr, ok := key.(string)
		if !ok {
			err = fmt.Errorf("testfixtures: record map key is not a string")
			return
		}

		sqlColumns = append(sqlColumns, h.quoteKeyword(keyStr))

		// if string, try convert to SQL or time
		// if map or array, convert to json
		switch v := value.(type) {
		case string:
			if strings.HasPrefix(v, "RAW=") {
				sqlValues = append(sqlValues, strings.TrimPrefix(v, "RAW="))
				continue
			}

			if t, err := tryStrToDate(v); err == nil {
				value = t
			}
		case []interface{}, map[interface{}]interface{}:
			value = recursiveToJSON(v)
		}

		switch h.paramType() {
		case paramTypeDollar:
			sqlValues = append(sqlValues, fmt.Sprintf("$%d", i))
		case paramTypeQuestion:
			sqlValues = append(sqlValues, "?")
		}

		values = append(values, value)
		i++
	}

	sqlStr = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		h.quoteKeyword(f.fileNameWithoutExtension()),
		strings.Join(sqlColumns, ", "),
		strings.Join(sqlValues, ", "),
	)
	return
}

func fixturesFromDir(dir string) ([]*fixtureFile, error) {
	var files []*fixtureFile
	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf(`testfixtures: could not stat directory "%s": %w`, dir, err)
	}

	for _, fileinfo := range fileinfos {
		fileExt := filepath.Ext(fileinfo.Name())
		if !fileinfo.IsDir() && (fileExt == ".yml" || fileExt == ".yaml") {
			fixture := &fixtureFile{
				path:     path.Join(dir, fileinfo.Name()),
				fileName: fileinfo.Name(),
			}
			fixture.content, err = ioutil.ReadFile(fixture.path)
			if err != nil {
				return nil, fmt.Errorf(`testfixtures: could not read file "%s": %w`, fixture.path, err)
			}
			files = append(files, fixture)
		}
	}
	return files, nil
}

func fixturesFromFiles(fileNames ...string) ([]*fixtureFile, error) {
	var (
		fixtureFiles []*fixtureFile
		err          error
	)

	for _, f := range fileNames {
		fixture := &fixtureFile{
			path:     f,
			fileName: filepath.Base(f),
		}
		fixture.content, err = ioutil.ReadFile(fixture.path)
		if err != nil {
			return nil, fmt.Errorf(`testfixtures: could not read file "%s": %w`, fixture.path, err)
		}
		fixtureFiles = append(fixtureFiles, fixture)
	}

	return fixtureFiles, nil
}
