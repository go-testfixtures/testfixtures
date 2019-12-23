// +build postgresql

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestPostgreSQL(t *testing.T) {
	testTestFixtures(
		t,
		"postgres",
		os.Getenv("PG_CONN_STRING"),
		"testdata/schema/postgresql.sql",
	)
}

func TestPostgreSQLWithAlterConstraint(t *testing.T) {
	testTestFixtures(
		t,
		"postgres",
		os.Getenv("PG_CONN_STRING"),
		"testdata/schema/postgresql.sql",
		UseAlterConstraint(),
	)
}
