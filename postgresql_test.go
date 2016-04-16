// +build postgresql

package testfixtures

import (
	_ "github.com/lib/pq"
)

func init() {
	databases = append(databases,
		databaseTest{
			"postgres",
			"PG_CONN_STRING",
			"test_schema/postgresql.sql",
			&PostgreSQLHelper{},
		},
		databaseTest{
			"postgres",
			"PG_CONN_STRING",
			"test_schema/postgresql.sql",
			&PostgreSQLHelper{UseAlterConstraint: true},
		},
	)
}
