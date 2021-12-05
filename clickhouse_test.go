//go:build clickhouse
// +build clickhouse

package testfixtures

import (
	"os"
	"testing"

	_ "github.com/ClickHouse/clickhouse-go"
)

func TestClickhouse(t *testing.T) {
	testLoader(
		t,
		"clickhouse",
		os.Getenv("CLICKHOUSE_CONN_STRING"),
		"testdata/schema/clickhouse.sql",
	)
}
