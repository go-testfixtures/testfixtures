package testfixtures

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
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

func assertCount(t *testing.T, db *sql.DB, h Helper, table string, expectedCount int) {
	var count int

	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", h.quoteKeyword(table)))
	row.Scan(&count)
	if count != expectedCount {
		t.Errorf("%s should have %d, but has %d", table, expectedCount, count)
	}
}

func testLoadFixtures(t *testing.T, db *sql.DB, h Helper) {
	err := LoadFixtures("testdata/fixtures", db, h)
	if err != nil {
		t.Errorf("Error on loading fixtures: %v", err)
	}

	assertCount(t, db, h, "posts", 2)
	assertCount(t, db, h, "comments", 4)
	assertCount(t, db, h, "tags", 3)
	assertCount(t, db, h, "posts_tags", 2)

	// this insert is to test if the PostgreSQL sequences were reset
	var sql string
	switch h.paramType() {
	case paramTypeDollar:
		sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES ($1, $2, $3, $4)"
	case paramTypeQuestion:
		sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)"
	case paramTypeColon:
		sql = "INSERT INTO posts (id, title, content, created_at, updated_at) VALUES (POSTS_SEQ.NEXTVAL, :1, :2, :3, :4)"
	}
	_, err = db.Exec(
		sql,
		"Post title",
		"Post content",
		time.Now(),
		time.Now(),
	)
	if err != nil {
		t.Errorf("Error inserting post: %v", err)
	}
}

var fixturesFiles = []string{
	"testdata/fixtures/posts.yml",
	"testdata/fixtures/comments.yml",
	"testdata/fixtures/tags.yml",
	"testdata/fixtures/posts_tags.yml",
}

func testLoadFixtureFiles(t *testing.T, db *sql.DB, h Helper) {
	db.Exec("DELETE FROM %s", h.quoteKeyword("posts_tags"))
	db.Exec("DELETE FROM %s", h.quoteKeyword("comments"))
	db.Exec("DELETE FROM %s", h.quoteKeyword("posts"))
	db.Exec("DELETE FROM %s", h.quoteKeyword("tags"))

	err := LoadFixtureFiles(db, h, fixturesFiles...)
	if err != nil {
		t.Errorf("Error on loading fixtures: %v", err)
	}

	assertCount(t, db, h, "posts", 2)
	assertCount(t, db, h, "comments", 4)
	assertCount(t, db, h, "tags", 3)
	assertCount(t, db, h, "posts_tags", 2)
}

type databaseTest struct {
	name       string
	connEnv    string
	schemaFile string
	helper     Helper
}

var databases = []databaseTest{}

func TestLoadFixtures(t *testing.T) {
	if len(databases) == 0 {
		t.Error("No database choosen for tests!")
	}

	for _, database := range databases {
		connString := os.Getenv(database.connEnv)

		var bytes []byte

		fmt.Printf("Test for %s\n", database.name)

		db, err := sql.Open(database.name, connString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v\n", err)
		}

		defer db.Close()

		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v\n", err)
		}

		bytes, err = ioutil.ReadFile(database.schemaFile)
		if err != nil {
			log.Fatalf("Could not read file %s: %v\n", database.schemaFile, err)
		}

		_, err = db.Exec(string(bytes))
		if err != nil {
			log.Fatalf("Failed to create schema: %v\n", err)
		}

		testLoadFixtures(t, db, database.helper)
		testLoadFixtureFiles(t, db, database.helper)
	}
}

func TestInterfaces(t *testing.T) {
	// helpers should implement interface
	helpers := []interface{}{
		&PostgreSQL{},
		&MySQL{},
		&SQLite{},
		&SQLServer{},
		&Oracle{},

		&PostgreSQLHelper{},
		&MySQLHelper{},
		&SQLiteHelper{},
		&SQLServerHelper{},
		&OracleHelper{},
	}
	for _, h := range helpers {
		if _, ok := h.(Helper); !ok {
			t.Errorf("Helper doesn't implement interface")
		}
	}
}
