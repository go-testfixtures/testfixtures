package testfixtures

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

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
		if err != errDatabaseIsRequired {
			t.Error("should return an error if database if not given")
		}
	})

	t.Run("DialectIsRequired", func(t *testing.T) {
		_, err := New(Database(&sql.DB{}))
		if err != errDialectIsRequired {
			t.Error("should return an error if dialect if not given")
		}
	})
}

func testLoader(t *testing.T, dialect, connStr, schemaFilePath string, additionalOptions ...func(*Loader) error) {
	db, err := sql.Open(dialect, connStr)
	if err != nil {
		t.Errorf("failed to open database: %v", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("failed to connect to database: %v", err)
		return
	}

	schema, err := ioutil.ReadFile(schemaFilePath)
	if err != nil {
		t.Errorf("cannot read schema file: %v", err)
		return
	}
	helper, err := helperForDialect(dialect)
	if err != nil {
		t.Errorf("cannot get helper: %v", err)
		return
	}
	if err := helper.init(db); err != nil {
		t.Errorf("cannot init helper: %v", err)
		return
	}

	var batches [][]byte
	if h, ok := helper.(batchSplitter); ok {
		batches = append(batches, bytes.Split(schema, h.splitter())...)
	} else {
		batches = append(batches, schema)
	}

	for _, b := range batches {
		if _, err = db.Exec(string(b)); err != nil {
			t.Errorf("cannot load schema: %v", err)
			return
		}
	}

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

		assertFixturesLoaded(t, l)
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
		assertFixturesLoaded(t, l)
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
		assertFixturesLoaded(t, l)
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
		assertFixturesLoaded(t, l)
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
		assertFixturesLoaded(t, l)
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
		assertFixturesLoaded(t, l)
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
		assertFixturesLoaded(t, l)
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
		assertFixturesLoaded(t, l)
	})

	t.Run("GenerateAndLoad", func(t *testing.T) {
		dir, err := ioutil.TempDir(os.TempDir(), "testfixtures_test")
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
		switch helper.paramType() {
		case paramTypeDollar:
			sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES ($1, $2, $3, $4)"
		case paramTypeQuestion:
			sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)"
		case paramTypeAtSign:
			sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES (@p1, @p2, @p3, @p4)"
		default:
			panic("unrecognized param type")
		}

		_, err = db.Exec(sql, "Post title", "Post content", time.Now(), time.Now())
		if err != nil {
			t.Errorf("cannot insert post: %v", err)
		}
	})
}

func assertFixturesLoaded(t *testing.T, l *Loader) {
	assertCount(t, l, "posts", 2)
	assertCount(t, l, "comments", 4)
	assertCount(t, l, "tags", 3)
	assertCount(t, l, "posts_tags", 6)
	assertCount(t, l, "users", 2)
}

func assertCount(t *testing.T, l *Loader, table string, expectedCount int) {
	count := 0
	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s", l.helper.quoteKeyword(table))

	row := l.db.QueryRow(sql)
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
