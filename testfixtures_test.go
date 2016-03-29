package testfixtures

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func TestFixtureFile(t *testing.T) {
	f := &fixtureFile{fileName: "posts.yml"}
	file := f.fileNameWithoutExtension()
	if file != "posts" {
		t.Errorf("Should be 'posts', but returned %s", file)
	}
}

func assertCount(t *testing.T, db *sql.DB, table string, expectedCount int) {
	var count int

	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
	row.Scan(&count)
	if count != expectedCount {
		t.Errorf("%s should have %d, but has %d", table, expectedCount, count)
	}
}

func testLoadFixtures(t *testing.T, db *sql.DB, helper DataBaseHelper) {
	err := LoadFixtures("test_fixtures", db, helper)
	if err != nil {
		t.Errorf("Error on loading fixtures: %v", err)
	}

	assertCount(t, db, "posts", 2)
	assertCount(t, db, "comments", 4)
	assertCount(t, db, "tags", 3)
	assertCount(t, db, "posts_tags", 2)

	// this insert is to test if the PostgreSQL sequences were reset
	var sql string
	if helper.paramType() == paramTypeDollar {
		sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES ($1, $2, $3, $4)"
	} else {
		sql = "INSERT INTO posts (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)"
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

type databaseTest struct {
	name       string
	connEnv    string
	schemaFile string
	helper     DataBaseHelper
}

var databases = []databaseTest{
	{"postgres", "PG_CONN_STRING", "test_schema/postgresql.sql", &PostgreSQLHelper{}},
	{"postgres", "PG_CONN_STRING", "test_schema/postgresql.sql", &PostgreSQLHelper{UseAlterConstraint: true}},
	{"mysql", "MYSQL_CONN_STRING", "test_schema/mysql.sql", &MySQLHelper{}},
}

func TestLoadFixtures(t *testing.T) {
	for _, database := range databases {
		connString := os.Getenv(database.connEnv)
		if connString == "" {
			continue
		}

		var bytes []byte

		fmt.Printf("Test for %s\n", database.name)

		db, err := sql.Open(database.name, connString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v\n", err)
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
	}
}
