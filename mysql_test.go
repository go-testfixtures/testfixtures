// +build mysql

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQL(t *testing.T) {
	testLoader(
		t,
		"mysql",
		"mysql",
		os.Getenv("MYSQL_CONN_STRING"),
		"testdata/schema/mysql.sql",
	)
}
