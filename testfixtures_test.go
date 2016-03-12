package testfixtures

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

func TestMain(m *testing.M) {
	var bytes []byte
	var err error

	db, err = sql.Open("postgres", "dbname=testfixtures-test")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	bytes, err = ioutil.ReadFile("test_schema/postgresql.sql")
	if err != nil {
		log.Fatalf("Could not read file postgresql.sql: %v\n", err)
	}

	_, err = db.Exec(string(bytes))
	if err != nil {
		log.Fatalf("Failed to create schema: %v\n", err)
	}

	os.Exit(m.Run())
}

func TestFixtureFile(t *testing.T) {
	f := &FixtureFile{FileName: "posts.yml"}
	file := f.FileNameWithoutExtension()
	if file != "posts" {
		t.Errorf("Should be 'posts', but returned %s", file)
	}
}

func assertCount(t *testing.T, table string, expectedCount int) {
	var count int

	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
	row.Scan(&count)
	if count != expectedCount {
		t.Errorf("%s should have %d, but has %d", table, expectedCount, count)
	}
}

func TestLoadFixtures(t *testing.T) {
	err := LoadFixtures("test_fixtures", db, &PostgreSQLHelper{})
	if err != nil {
		t.Errorf("Error on loading fixtures: %v", err)
	}

	assertCount(t, "posts", 2)
	assertCount(t, "comments", 4)
	assertCount(t, "tags", 3)
	assertCount(t, "posts_tags", 2)
}
