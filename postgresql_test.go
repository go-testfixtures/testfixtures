// +build postgresql

package testfixtures

import (
	_ "github.com/lib/pq"

	"testing"
)

func init() {
	databases = append(databases,
		databaseTest{
			"postgres",
			"PG_CONN_STRING",
			"testdata/schema/postgresql.sql",
			&PostgreSQL{},
		},
		databaseTest{
			"postgres",
			"PG_CONN_STRING",
			"testdata/schema/postgresql.sql",
			&PostgreSQL{UseAlterConstraint: true},
		},
	)
}

type PostgreSQLTableNameTestCase struct {
	FileName string
	Escaped  string
}

func testEscapeTableName(t *testing.T) {
	pql := &PostgreSQL{}

	tables := []PostgreSQLTableNameTestCase{
		PostgreSQLTableNameTestCase{
			FileName: "posts_tags",
			Escaped:  `"posts_tags"`,
		},
		PostgreSQLTableNameTestCase{
			FileName: "comments",
			Escaped:  `"comments"`,
		},
		PostgreSQLTableNameTestCase{
			FileName: "posts",
			Escaped:  `"posts"`,
		},
		PostgreSQLTableNameTestCase{
			FileName: "tags",
			Escaped:  `"tags"`,
		},
		PostgreSQLTableNameTestCase{
			FileName: "posts_tags",
			Escaped:  `"posts_tags"`,
		},
		PostgreSQLTableNameTestCase{
			FileName: "anotherschema.anothertable",
			Escaped:  `"anotherschema"."anothertable"`,
		},
	}

	for _, table := range tables {
		escaped := pql.quoteKeyword(table.FileName)

		if escaped != table.Escaped {
			t.Errorf("PostgreSQL Escaped table name %s should have equalled %s. Received %s instead", table, table.Escaped, escaped)
		}
	}
}
