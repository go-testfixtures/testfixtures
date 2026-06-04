package dbtests

import (
	"testing"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func TestClickhouse(t *testing.T) {
	t.Parallel()

	connStr := createClickhouseContainer(t)
	db := openDB(t, "clickhouse", connStr)
	loadSchemaInBatchesBySplitter(t, db, "testdata/schema/clickhouse.sql", []byte(";\n"))
	testLoader(t, db, "clickhouse")
}
