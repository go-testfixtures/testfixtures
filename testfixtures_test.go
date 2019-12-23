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

func testLoader(t *testing.T, driver, connStr, schemaFilePath string, additionalOptions ...func(*Loader) error) {
	db, err := sql.Open(driver, connStr)
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
	helper, err := helperForDriver(driver)
	if err != nil {
		t.Errorf("cannot get helper: %v", err)
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
				Driver(driver),
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
		assertFixturesLoaded(t, l)
	})

	t.Run("LoadFromFiles", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Driver(driver),
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

	t.Run("GenerateAndLoad", func(t *testing.T) {
		options := append(
			[]func(*Loader) error{
				Database(db),
				Driver(driver),
				Directory("testdata/fixtures"),
			},
			additionalOptions...,
		)
		l, err := New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}

		dir, err := ioutil.TempDir(os.TempDir(), "testfixtures_test")
		if err != nil {
			t.Errorf("cannot create temp dir: %v", err)
			return
		}
		if err := GenerateFixtures(db, l.helper, dir); err != nil {
			t.Errorf("cannot generate fixtures: %v", err)
			return
		}

		if err := l.Load(); err != nil {
			t.Error(err)
		}
	})

	t.Run("InserAfterLoad", func(t *testing.T) {
		// This test was originally written to catch a bug where it
		// wasn't possible to insert a record on PostgreSQL due
		// sequence issues.

		var sql string
		switch helper.paramType() {
		case paramTypeDollar:
			sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES ($1, $2, $3, $4)"
		case paramTypeQuestion:
			sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)"
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
	assertCount(t, l, "posts_tags", 2)
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
		Helper   Helper
		Keyword  string
		Expected string
	}{
		{&PostgreSQL{}, `posts_tags`, `"posts_tags"`},
		{&PostgreSQL{}, `test_schema.posts_tags`, `"test_schema"."posts_tags"`},
		{&SQLServer{}, `posts_tags`, `[posts_tags]`},
		{&SQLServer{}, `test_schema.posts_tags`, `[test_schema].[posts_tags]`},
	}

	for _, test := range tests {
		actual := test.Helper.quoteKeyword(test.Keyword)

		if test.Expected != actual {
			t.Errorf("TestQuoteKeyword keyword %s should have escaped to %s. Received %s instead", test.Keyword, test.Expected, actual)
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
