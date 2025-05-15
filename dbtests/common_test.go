package dbtests

import (
	"bytes"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/joho/godotenv/autoload"
)

//go:embed testdata
var fixtures embed.FS //nolint:unused

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

func testLoader(t *testing.T, db *sql.DB, dialect string, additionalOptions ...func(*testfixtures.Loader) error) { //nolint:unused
	t.Run("LoadFromDirectory", func(t *testing.T) {
		options := append(
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Directory("testdata/fixtures"),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Directory("testdata/fixtures"),
				testfixtures.SkipTableChecksumComputation(),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Directory("testdata/fixtures_dirs/fixtures1"),
				testfixtures.Directory("testdata/fixtures_dirs/fixtures2"),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Files(
					"testdata/fixtures/posts.yml",
					"testdata/fixtures/comments.yml",
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/posts_tags.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/assets.yml",
					"testdata/fixtures/accounts.yml",
					"testdata/fixtures/transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Files(
					"testdata/fixtures/posts.yml",
					"testdata/fixtures/comments.yml",
				),
				testfixtures.Files(
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/posts_tags.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/assets.yml",
					"testdata/fixtures/accounts.yml",
					"testdata/fixtures/transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.FilesMultiTables(
					"testdata/fixtures_multi_tables/posts_comments.yml",
					"testdata/fixtures_multi_tables/tags.yml",
					"testdata/fixtures_multi_tables/users.yml",
					"testdata/fixtures_multi_tables/posts_tags.yml",
					"testdata/fixtures_multi_tables/assets.yml",
					"testdata/fixtures_multi_tables/accounts_transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.FS(fixtures),
				testfixtures.FilesMultiTables(
					"testdata/fixtures_multi_tables/posts_comments.yml",
					"testdata/fixtures_multi_tables/tags.yml",
					"testdata/fixtures_multi_tables/users.yml",
					"testdata/fixtures_multi_tables/posts_tags.yml",
					"testdata/fixtures_multi_tables/assets.yml",
					"testdata/fixtures_multi_tables/accounts_transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Directory("testdata/fixtures_dirs/fixtures1"),
				testfixtures.Files(
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/accounts.yml",
					"testdata/fixtures/transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.FS(fixtures),
				testfixtures.Directory("testdata/fixtures_dirs/fixtures1"),
				testfixtures.Files(
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/accounts.yml",
					"testdata/fixtures/transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Paths(
					"testdata/fixtures_dirs/fixtures1",
					"testdata/fixtures_dirs/fixtures2/tags.yml",
					"testdata/fixtures_dirs/fixtures2/users.yml",
					"testdata/fixtures_dirs/fixtures2/accounts.yml",
					"testdata/fixtures_dirs/fixtures2/transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.FS(fixtures),
				testfixtures.Paths(
					"testdata/fixtures_dirs/fixtures1",
					"testdata/fixtures_dirs/fixtures2/tags.yml",
					"testdata/fixtures_dirs/fixtures2/users.yml",
					"testdata/fixtures_dirs/fixtures2/accounts.yml",
					"testdata/fixtures_dirs/fixtures2/transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Paths(
					"testdata/fixtures/posts.yml",
					"testdata/fixtures/comments.yml",
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/posts_tags.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/assets.yml",
					"testdata/fixtures/accounts.yml",
					"testdata/fixtures/transactions.yml",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Paths(
					"testdata/fixtures_dirs/fixtures1",
					"testdata/fixtures_dirs/fixtures2",
				),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
		dumper, err := testfixtures.NewDumper(
			testfixtures.DumpDatabase(db),
			testfixtures.DumpDialect(dialect),
			testfixtures.DumpDirectory(dir),
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
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Directory(dir),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
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
	assertCount(t, db, "accounts", 2)
	assertCount(t, db, "transactions", 4)
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
