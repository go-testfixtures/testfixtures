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
	helper        Helper
	fixturesFiles []*fixtureFile
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
	dbnameRegexp = regexp.MustCompile("(?i)test")
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

// Driver informs Loader about which database driver you're using.
// Possible options are "postgresql", "mysql", "sqlite" and "mssql".
func Driver(driver string) func(*Loader) error {
	return func(l *Loader) error {
		h, err := helperForDriver(driver)
		if err != nil {
			return err
		}
		l.helper = h
		return nil
	}
}

func helperForDriver(driver string) (Helper, error) {
	switch driver {
	case "postgres":
		return &PostgreSQL{}, nil
	case "mysql":
		return &MySQL{}, nil
	case "sqlite3":
		return &SQLite{}, nil
	case "mssql":
		return &SQLServer{}, nil
	default:
		return nil, fmt.Errorf(`testfixtures: unrecognized driver "%s"`, driver)
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
		pgHelper, ok := l.helper.(*PostgreSQL)
		if !ok {
			return fmt.Errorf("testfixtures: UseAlterConstraint is only valid for PostgreSQL databases")
		}
		pgHelper.useAlterConstraint = true
		return nil
	}
}

// SkipResetSequences prevents the reset of the databases
// sequences after load fixtures time
//
// Only valid for PostgreSQL. Returns an error otherwise.
func SkipResetSequences() func(*Loader) error {
	return func(l *Loader) error {
		pgHelper, ok := l.helper.(*PostgreSQL)
		if !ok {
			return fmt.Errorf("testfixtures: SkipResetSequences is only valid for PostgreSQL databases")
		}
		pgHelper.skipResetSequences = true
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

// DetectTestDatabase returns nil if databaseName matches regexp
//     if err := fixtures.DetectTestDatabase(); err != nil {
//         log.Fatal(err)
//     }
func (l *Loader) DetectTestDatabase() error {
	dbName, err := l.helper.databaseName(l.db)
	if err != nil {
		return err
	}
	if !dbnameRegexp.MatchString(dbName) {
		return ErrNotTestDatabase
	}
	return nil
}

// Load wipes and after load all fixtures in the database.
//     if err := fixtures.Load(); err != nil {
//         log.Fatal(err)
//     }
func (l *Loader) Load() error {
	if !skipDatabaseNameCheck {
		if err := l.DetectTestDatabase(); err != nil {
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

func (l *Loader) buildInsertSQLs() error {
	for _, f := range l.fixturesFiles {
		var records interface{}
		if err := yaml.Unmarshal(f.content, &records); err != nil {
			return err
		}

		switch records := records.(type) {
		case []interface{}:
			for _, record := range records {
				recordMap, ok := record.(map[interface{}]interface{})
				if !ok {
					return ErrWrongCastNotAMap
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
					return ErrWrongCastNotAMap
				}

				sql, values, err := f.buildInsertSQL(l.helper, recordMap)
				if err != nil {
					return err
				}

				f.insertSQLs = append(f.insertSQLs, insertSQL{sql, values})
			}
		default:
			return ErrFileIsNotSliceOrMap
		}
	}

	return nil
}

func (f *fixtureFile) fileNameWithoutExtension() string {
	return strings.Replace(f.fileName, filepath.Ext(f.fileName), "", 1)
}

func (f *fixtureFile) delete(tx *sql.Tx, h Helper) error {
	_, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", h.quoteKeyword(f.fileNameWithoutExtension())))
	return err
}

func (f *fixtureFile) buildInsertSQL(h Helper, record map[interface{}]interface{}) (sqlStr string, values []interface{}, err error) {
	var (
		sqlColumns []string
		sqlValues  []string
		i          = 1
	)
	for key, value := range record {
		keyStr, ok := key.(string)
		if !ok {
			err = ErrKeyIsNotString
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
		return nil, err
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
				return nil, err
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
			return nil, err
		}
		fixtureFiles = append(fixtureFiles, fixture)
	}

	return fixtureFiles, nil
}
