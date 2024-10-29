package testfixtures

import (
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

//go:embed testdata
var fixtures embed.FS //nolint:unused

func TestFixtureFile(t *testing.T) {
	f := &fixtureFile{fileName: "posts.yml"}
	file := f.fileNameWithoutExtension()
	if file != "posts" {
		t.Errorf("Should be 'posts', but returned %s", file)
	}
}

func TestRequiredOptions(t *testing.T) {
	t.Run("DatabaseIsRequired", func(t *testing.T) {
		_, err := New()
		if !errors.Is(err, errDatabaseIsRequired) {
			t.Error("should return an error if database if not given")
		}
	})

	t.Run("DialectIsRequired", func(t *testing.T) {
		_, err := New(Database(&sql.DB{}))
		if !errors.Is(err, errDialectIsRequired) {
			t.Error("should return an error if dialect if not given")
		}
	})
}

func openDB(t *testing.T, dialect, connStr string) *sql.DB { //nolint:unused
	t.Helper()
	db, err := sql.Open(dialect, connStr)
	if err != nil {
		t.Errorf("failed to open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Ping(); err != nil {
		t.Errorf("failed to connect to database: %v", err)
	}
	return db
}

func loadSchemaInOneQuery(t *testing.T, db *sql.DB, schemaFilePath string) { //nolint:unused
	t.Helper()
	schema, err := os.ReadFile(schemaFilePath)
	if err != nil {
		t.Errorf("cannot read schema file: %v", err)
		return
	}
	loadSchemaInBatches(t, db, [][]byte{schema})
}

func loadSchemaInBatchesBySplitter(t *testing.T, db *sql.DB, schemaFilePath string, splitter []byte) { //nolint:unused
	t.Helper()
	schema, err := os.ReadFile(schemaFilePath)
	if err != nil {
		t.Errorf("cannot read schema file: %v", err)
		return
	}
	batches := bytes.Split(schema, splitter)
	loadSchemaInBatches(t, db, batches)
}

func loadSchemaInBatches(t *testing.T, db *sql.DB, batches [][]byte) { //nolint:unused
	t.Helper()
	for _, b := range batches {
		if len(b) == 0 {
			continue
		}
		if _, err := db.Exec(string(b)); err != nil {
			t.Errorf("cannot load schema: %v", err)
			return
		}
	}
}

func testLoader(t *testing.T, db *sql.DB, dialect string, additionalOptions ...func(*Loader) error) { //nolint:unused
	t.Run("LoadFromDirectory", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Directory("testdata/fixtures"),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		// Call load again to test against a database with existing data.
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromDirectory with SkipTableChecksumComputation", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Directory("testdata/fixtures"),
				SkipTableChecksumComputation(),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		// Call load again to test against a database with existing data.
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromDirectory-Multiple", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Directory("testdata/fixtures_dirs/fixtures1"),
				Directory("testdata/fixtures_dirs/fixtures2"),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromFiles", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Files(
					"testdata/fixtures/posts.yml",
					"testdata/fixtures/comments.yml",
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/posts_tags.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/assets.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromFiles-Multiple", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Files(
					"testdata/fixtures/posts.yml",
					"testdata/fixtures/comments.yml",
				),
				Files(
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/posts_tags.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/assets.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromFiles-MultiTables", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				FilesMultiTables(
					"testdata/fixtures_multi_tables/posts_comments.yml",
					"testdata/fixtures_multi_tables/tags.yml",
					"testdata/fixtures_multi_tables/users.yml",
					"testdata/fixtures_multi_tables/posts_tags.yml",
					"testdata/fixtures_multi_tables/assets.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromFiles-MultiTablesWithFS", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				FS(fixtures),
				FilesMultiTables(
					"testdata/fixtures_multi_tables/posts_comments.yml",
					"testdata/fixtures_multi_tables/tags.yml",
					"testdata/fixtures_multi_tables/users.yml",
					"testdata/fixtures_multi_tables/posts_tags.yml",
					"testdata/fixtures_multi_tables/assets.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromDirectoryAndFiles", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Directory("testdata/fixtures_dirs/fixtures1"),
				Files(
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/users.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromDirectoryAndFilesWithFS", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				FS(fixtures),
				Directory("testdata/fixtures_dirs/fixtures1"),
				Files(
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/users.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromPaths", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Paths(
					"testdata/fixtures_dirs/fixtures1",
					"testdata/fixtures_dirs/fixtures2/tags.yml",
					"testdata/fixtures_dirs/fixtures2/users.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromPathsWithFS", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				FS(fixtures),
				Paths(
					"testdata/fixtures_dirs/fixtures1",
					"testdata/fixtures_dirs/fixtures2/tags.yml",
					"testdata/fixtures_dirs/fixtures2/users.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromPaths-OnlyFiles", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Paths(
					"testdata/fixtures/posts.yml",
					"testdata/fixtures/comments.yml",
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/posts_tags.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/assets.yml",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("LoadFromPaths-OnlyDirs", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Template(),
				TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				Paths(
					"testdata/fixtures_dirs/fixtures1",
					"testdata/fixtures_dirs/fixtures2",
				),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}
		assertFixturesLoaded(t, db)
	})

	t.Run("GenerateAndLoad", func(t *testing.T) {
		dir, err := os.MkdirTemp(os.TempDir(), "testfixtures_test")
		if err != nil {
			t.Errorf("cannot create temp dir: %v", err)
			return
		}
		dumper, err := NewDumper(
			DumpDatabase(db),
			DumpDialect(dialect),
			DumpDirectory(dir),
		)
		if err != nil {
			t.Errorf("could not create dumper: %v", err)
			return
		}
		if err := dumper.Dump(); err != nil {
			t.Errorf("cannot generate fixtures: %v", err)
			return
		}

		options := append(
			[]func(*Loader) error{
				Database(db),
				Dialect(dialect),
				Directory(dir),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}
		if err := l.Load(); err != nil {
			t.Error(err)
		}
	})

	t.Run("InsertAfterLoad", func(t *testing.T) {
		// This test was originally written to catch a bug where it
		// wasn't possible to insert a record on PostgreSQL due
		// sequence issues.

		var sql string
		switch dialect {
		case "postgres", "pgx", "clickhouse":
			sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES ($1, $2, $3, $4)"
		case "mysql", "sqlite3", "mssql":
			sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)"
		case "sqlserver", "spanner":
			sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES (@p1, @p2, @p3, @p4)"
		default:
			t.Fatalf("undefined param type for %s dialect, modify switch statement", dialect)
		}

		_, err := db.Exec(sql, "Post title", "Post content", time.Now(), time.Now())
		if err != nil {
			t.Errorf("cannot insert post: %v", err)
		}
	})
}

func assertFixturesLoaded(t *testing.T, db *sql.DB) { //nolint
	assertCount(t, db, "posts", 2)
	assertCount(t, db, "comments", 4)
	assertCount(t, db, "tags", 3)
	assertCount(t, db, "posts_tags", 6)
	assertCount(t, db, "users", 2)
	assertCount(t, db, "assets", 1)
}

func assertCount(t *testing.T, db *sql.DB, table string, expectedCount int) { //nolint
	count := 0
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)

	row := db.QueryRow(sql)
	if err := row.Scan(&count); err != nil {
		t.Errorf("cannot query table: %v", err)
	}

	if count != expectedCount {
		t.Errorf("%s should have %d, but has %d", table, expectedCount, count)
	}
}

func TestQuoteKeyword(t *testing.T) {
	tests := []struct {
		helper   helper
		keyword  string
		expected string
	}{
		{&postgreSQL{}, `posts_tags`, `"posts_tags"`},
		{&postgreSQL{}, `test_schema.posts_tags`, `"test_schema"."posts_tags"`},
		{&sqlserver{}, `posts_tags`, `[posts_tags]`},
		{&sqlserver{}, `test_schema.posts_tags`, `[test_schema].[posts_tags]`},
	}

	for _, test := range tests {
		actual := test.helper.quoteKeyword(test.keyword)

		if test.expected != actual {
			t.Errorf("TestQuoteKeyword keyword %s should have escaped to %s. Received %s instead", test.keyword, test.expected, actual)
		}
	}
}

func TestEnsureTestDatabase(t *testing.T) {
	tests := []struct {
		name           string
		isTestDatabase bool
	}{
		{"db_test", true},
		{"dbTEST", true},
		{"testdb", true},
		{"production", false},
		{"productionTestCopy", true},
		{"t_e_s_t", false},
		{"ТESТ", false}, // cyrillic T
	}

	for _, it := range tests {
		var (
			mockedHelper = NewMockHelper(it.name)
			l            = &Loader{helper: mockedHelper}
			err          = l.EnsureTestDatabase()
		)
		if err != nil && it.isTestDatabase {
			t.Errorf("EnsureTestDatabase() should return nil for name = %s", it.name)
		}
		if err == nil && !it.isTestDatabase {
			t.Errorf("EnsureTestDatabase() should return error for name = %s", it.name)
		}
	}
}
