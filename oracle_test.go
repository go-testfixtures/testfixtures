// +build oracle

package testfixtures

import (
	_ "gopkg.in/rana/ora.v3"
)

func init() {
	databases = append(databases,
		databaseTest{
			"ora",
			"ORACLE_CONN_STRING",
			"testdata/schema/oracle.sql",
			&OracleHelper{},
		},
	)
}
