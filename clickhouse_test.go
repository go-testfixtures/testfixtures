// +build mysql

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestClickHouse(t *testing.T) {
	testLoader(
		t,
		"clickhouse",
		os.Getenv("CLICKHOUSE_CONN_STRING"),
		"testdata/schema/clickhouse.sql",
	)
}
