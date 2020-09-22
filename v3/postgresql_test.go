// +build postgresql

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
)

func TestPostgreSQL(t *testing.T) {
	for _, dialect := range []string{"postgres", "pgx"} {
		testLoader(
			t,
			dialect,
			os.Getenv("PG_CONN_STRING"),
			"testdata/schema/postgresql.sql",
		)
	}
}

func TestPostgreSQLWithAlterConstraint(t *testing.T) {
	for _, dialect := range []string{"postgres", "pgx"} {
		testLoader(
			t,
			dialect,
			os.Getenv("PG_CONN_STRING"),
			"testdata/schema/postgresql.sql",
			UseAlterConstraint(),
		)
	}
}

func TestPostgreSQLWithDropConstraint(t *testing.T) {
	for _, dialect := range []string{"postgres", "pgx"} {
		testLoader(
			t,
			dialect,
			os.Getenv("PG_CONN_STRING"),
			"testdata/schema/postgresql.sql",
			UseDropConstraint(),
		)
	}
}
