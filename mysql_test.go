// +build mysql

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQL(t *testing.T) {
	testTestFixtures(
		t,
		"mysql",
		os.Getenv("MYSQL_CONN_STRING"),
		"testdata/schema/mysql.sql",
	)
}
