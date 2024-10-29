//go:build clickhouse

package dbtests

import (
	"os"
	"testing"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func TestClickhouse(t *testing.T) {
	db := openDB(t, "clickhouse", os.Getenv("CLICKHOUSE_CONN_STRING"))
	loadSchemaInBatchesBySplitter(t, db, "testdata/schema/clickhouse.sql", []byte(";\n"))
	testLoader(t, db, "clickhouse")
}
