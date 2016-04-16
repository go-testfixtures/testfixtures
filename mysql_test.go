// +build mysql

package testfixtures

import (
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	databases = append(databases,
		databaseTest{
			"mysql",
			"MYSQL_CONN_STRING",
			"test_schema/mysql.sql",
			&MySQLHelper{},
		},
	)
}
