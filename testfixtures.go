package testfixtures // import "github.com/go-testfixtures/testfixtures/v3"

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/go-testfixtures/testfixtures/v3/shared"
)

// Loader is the responsible to loading fixtures.
type Loader struct {
	db            *sql.DB
	helper        helper
	fixturesFiles []*fixtureFile

	skipCleanup             bool
	skipChecksumComputation bool
	skipTestDatabaseCheck   bool
	location                *time.Location

	template           bool
	templateFuncs      template.FuncMap
	templateLeftDelim  string
	templateRightDelim string
	templateOptions    []string
	templateData       interface{}

	fs fs.FS
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
	l := &Loader{
		templateLeftDelim:  "{{",
		templateRightDelim: "}}",
		templateOptions:    []string{"missingkey=zero"},
		fs:                 defaultFS{},
	}

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

// FS sets other fs.FS implementation
//
// Example embed.FS
// By default usage os package implementation
func FS(fs fs.FS) func(*Loader) error {
	return func(l *Loader) error {
		l.fs = fs
		return nil
	}
}

// Database sets an existing sql.DB instant to Loader.
func Database(db *sql.DB) func(*Loader) error {
	return func(l *Loader) error {
		l.db = db
		return nil
	}
}

type DialectOptions func(h helper) error

// WithCustomPlaceholder - allow to provide custom placeholder in queries
func WithCustomPlaceholder(placeholder ParamType) DialectOptions {
	return func(l helper) error {
		if err := placeholder.Valid(); err != nil {
			return err
		}

		l.setCustomParamType(placeholder)
		return nil
	}
}

// Dialect informs Loader about which database dialect you're using.
//
// Possible options are "postgresql", "timescaledb", "mysql", "mariadb",
// "sqlite", "sqlserver", "clickhouse", "spanner".
func Dialect(dialect string, opts ...DialectOptions) func(*Loader) error {
	return func(l *Loader) error {
		h, err := helperForDialect(dialect)
		if err != nil {
			return err
		}
		for _, opt := range opts {
			if err = opt(h); err != nil {
				return err
			}
		}
		l.helper = h

		return nil
	}
}

func helperForDialect(dialect string) (helper, error) {
	switch dialect {
	case "postgres", "postgresql", "timescaledb", "pgx":
		return &postgreSQL{}, nil
	case "mysql", "mariadb":
		return &mySQL{}, nil
	case "sqlite", "sqlite3":
		return &sqlite{}, nil
	case "mssql", "sqlserver":
		return &sqlserver{}, nil
	case "clickhouse":
		return &clickhouse{}, nil
	case "spanner":
		return &spanner{}, nil
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

// UseDropConstraint If true, the constraints will be dropped
// and recreated after loading fixtures. This is implemented mainly to support
// CockroachDB which does not support other methods.
// Only valid for PostgreSQL dialect. Returns an error otherwise.

func UseDropConstraint() func(*Loader) error {
	return func(l *Loader) error {
		pgHelper, ok := l.helper.(*postgreSQL)
		if !ok {
			return fmt.Errorf("testfixtures: UseDropConstraint is only valid for PostgreSQL databases")
		}
		pgHelper.useDropConstraint = true
		return nil
	}
}

// SkipResetSequences prevents Loader from reseting sequences after loading
// fixtures.
//
// Only valid for PostgreSQL and MySQL. Returns an error otherwise.
func SkipResetSequences() func(*Loader) error {
	return func(l *Loader) error {
		switch helper := l.helper.(type) {
		case *postgreSQL:
			helper.skipResetSequences = true
		case *mySQL:
			helper.skipResetSequences = true
		default:
			return fmt.Errorf("testfixtures: SkipResetSequences is valid for PostgreSQL and MySQL databases")
		}
		return nil
	}
}

// AllowMultipleStatementsInOneQuery is a performance tweak for running some setup statements in one query for better performance.
//
// Your database and connection must be configured to support it: https://github.com/go-sql-driver/mysql?tab=readme-ov-file#multistatements
//
// Only valid for MySQL as it is not enabled by default.
func AllowMultipleStatementsInOneQuery() func(*Loader) error {
	return func(l *Loader) error {
		switch helper := l.helper.(type) {
		case *mySQL:
			helper.allowMultipleStatementsInOneQuery = true
		default:
			return fmt.Errorf("testfixtures: AllowMultipleStatementsInOneQuery is valid for MySQL database")
		}
		return nil
	}
}

// ResetSequencesTo sets the value the sequences will be reset to.
//
// Defaults to 10000.
//
// Only valid for PostgreSQL and MySQL. Returns an error otherwise.
func ResetSequencesTo(value int64) func(*Loader) error {
	return func(l *Loader) error {
		switch helper := l.helper.(type) {
		case *postgreSQL:
			helper.resetSequencesTo = value
		case *mySQL:
			helper.resetSequencesTo = value
		default:
			return fmt.Errorf("testfixtures: ResetSequencesTo is only valid for PostgreSQL and MySQL databases")
		}
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

// DangerousSkipCleanupFixtureTables will make Loader not wipe data from fixture tables.
// This may lead to dirty data in the test database.
// Use with caution!
func DangerousSkipCleanupFixtureTables() func(*Loader) error {
	return func(l *Loader) error {
		l.skipCleanup = true
		return nil
	}
}

// SkipTableChecksumComputation will make Loader not compute table checksum at the end of the Loader.Load.
// This may improve performance. It may also slow down subsequent calls to Loader.Load, because checksums are used
// to not reload unchanged tables.
func SkipTableChecksumComputation() func(*Loader) error {
	return func(l *Loader) error {
		l.skipChecksumComputation = true
		return nil
	}
}

// Directory informs Loader to load YAML files from a given directory.
func Directory(dir string) func(*Loader) error {
	return func(l *Loader) error {
		_, ok := l.helper.(*spanner)
		if ok {
			return fmt.Errorf(shared.ErrorMessage_NotSupportedLoadingMethod, "Directory")
		}
		fixtures, err := l.fixturesFromDir(dir)
		if err != nil {
			return err
		}
		l.fixturesFiles = append(l.fixturesFiles, fixtures...)
		return nil
	}
}

// Files informs Loader to load a given set of YAML files.
func Files(files ...string) func(*Loader) error {
	return func(l *Loader) error {
		fixtures, err := l.fixturesFromFiles(files...)
		if err != nil {
			return err
		}
		l.fixturesFiles = append(l.fixturesFiles, fixtures...)
		return nil
	}
}

// Paths inform Loader to load a given set of YAML files and directories.
func Paths(paths ...string) func(*Loader) error {
	return func(l *Loader) error {
		_, ok := l.helper.(*spanner)
		if ok {
			return fmt.Errorf(shared.ErrorMessage_NotSupportedLoadingMethod, "Paths")
		}
		fixtures, err := l.fixturesFromPaths(paths...)
		if err != nil {
			return err
		}
		l.fixturesFiles = append(l.fixturesFiles, fixtures...)
		return nil
	}
}

// Files informs Loader to load a given set of YAML files as mutiple fixtures.
func FilesMultiTables(files ...string) func(*Loader) error {
	return func(l *Loader) error {
		fixtures, err := l.fixturesFromFilesMultiTables(files...)
		if err != nil {
			return err
		}
		l.fixturesFiles = append(l.fixturesFiles, fixtures...)
		return nil
	}
}

// Location makes Loader use the given location by default when parsing
// dates. If not given, by default it uses the value of time.Local.
func Location(location *time.Location) func(*Loader) error {
	return func(l *Loader) error {
		l.location = location
		return nil
	}
}

// Template makes loader process each YAML file as an template using the
// text/template package.
//
// For more information on how templates work in Go please read:
// https://golang.org/pkg/text/template/
//
// If not given the YAML files are parsed as is.
func Template() func(*Loader) error {
	return func(l *Loader) error {
		l.template = true
		return nil
	}
}

// TemplateFuncs allow choosing which functions will be available
// when processing templates.
//
// For more information see: https://golang.org/pkg/text/template/#Template.Funcs
func TemplateFuncs(funcs template.FuncMap) func(*Loader) error {
	return func(l *Loader) error {
		if !l.template {
			return fmt.Errorf(`testfixtures: the Template() options is required in order to use the TemplateFuns() option`)
		}

		l.templateFuncs = funcs
		return nil
	}
}

// TemplateDelims allow choosing which delimiters will be used for templating.
// This defaults to "{{" and "}}".
//
// For more information see https://golang.org/pkg/text/template/#Template.Delims
func TemplateDelims(left, right string) func(*Loader) error {
	return func(l *Loader) error {
		if !l.template {
			return fmt.Errorf(`testfixtures: the Template() options is required in order to use the TemplateDelims() option`)
		}

		l.templateLeftDelim = left
		l.templateRightDelim = right
		return nil
	}
}

// TemplateOptions allows you to specific which text/template options will
// be enabled when processing templates.
//
// This defaults to "missingkey=zero". Check the available options here:
// https://golang.org/pkg/text/template/#Template.Option
func TemplateOptions(options ...string) func(*Loader) error {
	return func(l *Loader) error {
		if !l.template {
			return fmt.Errorf(`testfixtures: the Template() options is required in order to use the TemplateOptions() option`)
		}

		l.templateOptions = options
		return nil
	}
}

// TemplateData allows you to specify which data will be available
// when processing templates. Data is accesible by prefixing it with a "."
// like {{.MyKey}}.
func TemplateData(data interface{}) func(*Loader) error {
	return func(l *Loader) error {
		if !l.template {
			return fmt.Errorf(`testfixtures: the Template() options is required in order to use the TemplateData() option`)
		}

		l.templateData = data
		return nil
	}
}

// ClickhouseUseDeleteFrom If true, then "DELETE FROM …" syntax will be used
// to clean data from Clickhouse.
// Default syntax is "TRUNCATE TABLE …"
//
// Only valid for Clickhouse. Returns an error otherwise.
func ClickhouseUseDeleteFrom() func(*Loader) error {
	return func(l *Loader) error {
		chHelper, ok := l.helper.(*clickhouse)
		if !ok {
			return fmt.Errorf("testfixtures: ClickhouseUseDeleteFrom is only valid for Clickhouse databases")
		}
		chHelper.cleanTableFn = func(tableName string) string {
			return fmt.Sprintf("DELETE FROM %s", tableName)
		}
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
//
//	if err := fixtures.Load(); err != nil {
//	        ...
//	}
func (l *Loader) Load() error {
	if !l.skipTestDatabaseCheck {
		if err := l.EnsureTestDatabase(); err != nil {
			return err
		}
	}

	err := l.helper.disableReferentialIntegrity(l.db, func(tx *sql.Tx) error {
		modifiedTables := make(map[string]bool, len(l.fixturesFiles))
		for _, file := range l.fixturesFiles {
			tableName := file.fileNameWithoutExtension()
			modified, err := l.helper.isTableModified(tx, tableName)
			if err != nil {
				return err
			}
			modifiedTables[tableName] = modified
		}

		// Delete existing table data for specified fixtures before populating the data. This helps avoid
		// DELETE CASCADE constraints when using the `UseAlterConstraint()` option.
		if !l.skipCleanup {
			for _, file := range l.fixturesFiles {
				modified := modifiedTables[file.fileNameWithoutExtension()]
				if !modified {
					continue
				}
				if err := file.delete(tx, l.helper); err != nil {
					return err
				}
			}
		}

		for _, file := range l.fixturesFiles {
			modified := modifiedTables[file.fileNameWithoutExtension()]
			if !modified {
				continue
			}
			err := l.helper.whileInsertOnTable(tx, file.fileNameWithoutExtension(), func() error {
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
	if !l.skipChecksumComputation {
		if err := l.helper.computeTablesChecksum(l.db); err != nil {
			return err
		}
	}
	return nil
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

func (l *Loader) buildInterfacesSlice(records interface{}) ([]interface{}, error) {
	switch records := records.(type) {
	case []interface{}:
		return records, nil
	case map[string]interface{}:
		var result []interface{}
		for _, record := range records {
			result = append(result, record)
		}
		return result, nil
	}

	return nil, fmt.Errorf("testfixtures: fixture is not a slice or map")
}

func (l *Loader) buildInsertSQLs() error {
	for _, f := range l.fixturesFiles {
		var records interface{}
		if err := yaml.Unmarshal(f.content, &records); err != nil {
			return fmt.Errorf("testfixtures: could not unmarshal YAML: %w", err)
		}

		result, err := l.buildInterfacesSlice(records)
		if err != nil {
			return err
		}

		f.insertSQLs = make([]insertSQL, 0, len(result))

		for _, record := range result {
			recordMap, ok := record.(map[string]interface{})
			if !ok {
				return fmt.Errorf("testfixtures: could not cast record: not a map[interface{}]interface{}")
			}

			sql, values, err := l.buildInsertSQL(f, recordMap)
			if err != nil {
				return err
			}

			f.insertSQLs = append(f.insertSQLs, insertSQL{sql, values})
		}
	}

	return nil
}

func (f *fixtureFile) fileNameWithoutExtension() string {
	return strings.Replace(f.fileName, filepath.Ext(f.fileName), "", 1)
}

func (f *fixtureFile) delete(tx *sql.Tx, h helper) error {
	deleteQuery := h.cleanTableQuery(h.quoteKeyword(f.fileNameWithoutExtension()))
	if _, err := tx.Exec(deleteQuery); err != nil {
		return fmt.Errorf(`testfixtures: could not clean table "%s": %w`, f.fileNameWithoutExtension(), err)
	}
	return nil
}

func (l *Loader) buildInsertSQL(f *fixtureFile, record map[string]interface{}) (sqlStr string, values []interface{}, err error) {
	var (
		sqlColumns = make([]string, 0, len(record))
		sqlValues  = make([]string, 0, len(record))
		i          = 1
	)
	for key, value := range record {
		sqlColumns = append(sqlColumns, l.helper.quoteKeyword(key))

		// if string, try convert to SQL or time
		// if map or array, convert to json
		switch v := value.(type) {
		case string:
			if strings.HasPrefix(v, "RAW=") {
				sqlValues = append(sqlValues, strings.TrimPrefix(v, "RAW="))
				continue
			}
			if b, err := l.tryHexStringToBytes(v); err == nil {
				value = b
			} else if t, err := l.tryStrToDate(v); err == nil {
				value = t
			}
		case []interface{}, map[string]interface{}:
			var bytes []byte
			bytes, err = json.Marshal(recursiveToJSON(v))
			if err != nil {
				return
			}
			value = string(bytes)
		}

		switch l.helper.paramType() {
		case ParamTypeDollar:
			sqlValues = append(sqlValues, fmt.Sprintf("$%d", i))
		case ParamTypeQuestion:
			sqlValues = append(sqlValues, "?")
		case ParamTypeAtSign:
			sqlValues = append(sqlValues, fmt.Sprintf("@p%d", i))
		}

		values = append(values, value)
		i++
	}

	sqlStr, err = l.helper.buildInsertSQL(
		l.db,
		l.helper.quoteKeyword(f.fileNameWithoutExtension()),
		sqlColumns, sqlValues,
	)
	return
}

func (l *Loader) fixturesFromDir(dir string) ([]*fixtureFile, error) {
	fileinfos, err := fs.ReadDir(l.fs, dir)
	if err != nil {
		return nil, fmt.Errorf(`testfixtures: could not stat directory "%s": %w`, dir, err)
	}

	files := make([]*fixtureFile, 0, len(fileinfos))

	for _, fileinfo := range fileinfos {
		fileExt := filepath.Ext(fileinfo.Name())
		if !fileinfo.IsDir() && (fileExt == ".yml" || fileExt == ".yaml") {
			fixture := &fixtureFile{
				path:     path.Join(dir, fileinfo.Name()),
				fileName: fileinfo.Name(),
			}
			fixture.content, err = fs.ReadFile(l.fs, fixture.path)
			if err != nil {
				return nil, fmt.Errorf(`testfixtures: could not read file "%s": %w`, fixture.path, err)
			}
			fixture.content, err = l.preProcessContent(fixture.fileName, fixture.content)
			if err != nil {
				return nil, err
			}
			files = append(files, fixture)
		}
	}
	return files, nil
}

func (l *Loader) fixturesFromFiles(fileNames ...string) ([]*fixtureFile, error) {
	var (
		fixtureFiles = make([]*fixtureFile, 0, len(fileNames))
		err          error
	)

	for _, f := range fileNames {
		fixture := &fixtureFile{
			path:     f,
			fileName: filepath.Base(f),
		}
		fixture.content, err = fs.ReadFile(l.fs, fixture.path)
		if err != nil {
			return nil, fmt.Errorf(`testfixtures: could not read file "%s": %w`, fixture.path, err)
		}
		fixture.content, err = l.preProcessContent(fixture.fileName, fixture.content)
		if err != nil {
			return nil, err
		}
		fixtureFiles = append(fixtureFiles, fixture)
	}

	return fixtureFiles, nil
}

func (l *Loader) fixturesFromPaths(paths ...string) ([]*fixtureFile, error) {
	fixtureExtractor := func(p string, isDir bool) ([]*fixtureFile, error) {
		if isDir {
			return l.fixturesFromDir(p)
		}

		return l.fixturesFromFiles(p)
	}

	var fixtureFiles []*fixtureFile

	for _, p := range paths {
		f, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf(`testfixtures: could not stat path "%s": %w`, p, err)
		}

		fixtures, err := fixtureExtractor(p, f.IsDir())
		if err != nil {
			return nil, err
		}

		fixtureFiles = append(fixtureFiles, fixtures...)
	}

	return fixtureFiles, nil
}

func (l *Loader) fixturesFromFilesMultiTables(fileNames ...string) ([]*fixtureFile, error) {
	fixtureFiles := make([]*fixtureFile, 0, len(fileNames))

	for _, f := range fileNames {
		var (
			content []byte
			err     error
		)

		content, err = fs.ReadFile(l.fs, f)
		if err != nil {
			return nil, fmt.Errorf(`testfixtures: could not read file "%s": %w`, f, err)
		}

		content, err = l.preProcessContent(f, content)
		if err != nil {
			return nil, err
		}

		var tablesMap yaml.MapSlice
		if err := yaml.UnmarshalWithOptions(content, &tablesMap, yaml.UseOrderedMap()); err != nil {
			return nil, fmt.Errorf("testfixtures: could not unmarshal YAML: %w", err)
		}

		for _, item := range tablesMap {
			table := item.Key.(string)
			records := item.Value
			result, err := l.buildInterfacesSlice(records)
			if err != nil {
				return nil, err
			}

			var content []byte
			if content, err = yaml.Marshal(result); err != nil {
				return nil, fmt.Errorf("testfixtures: could not marshal YAML: %w", err)
			}

			file := fmt.Sprintf("%s.yml", table)
			path := filepath.Join(filepath.Dir(f), file)
			fixtureFiles = append(fixtureFiles, &fixtureFile{
				path:     path,
				fileName: file,
				content:  content,
			})
		}
	}

	return fixtureFiles, nil
}

func (l *Loader) preProcessContent(name string, content []byte) ([]byte, error) {
	if !l.template {
		return content, nil
	}
	t := template.New("").
		Funcs(l.templateFuncs).
		Delims(l.templateLeftDelim, l.templateRightDelim).
		Option(l.templateOptions...)
	t, err := t.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf(`textfixtures: error on parsing template in %s: %w`, name, err)
	}

	var buffer bytes.Buffer
	if err := t.Execute(&buffer, l.templateData); err != nil {
		return nil, fmt.Errorf(`textfixtures: error on execute template in %s: %w`, name, err)
	}

	return buffer.Bytes(), nil
}
