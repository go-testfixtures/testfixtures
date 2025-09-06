module github.com/go-testfixtures/testfixtures/cmd/testfixtures/v3

go 1.23.0

toolchain go1.25.0

require (
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/go-sql-driver/mysql v1.9.3
	github.com/go-testfixtures/testfixtures/v3 v3.0.0
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/spf13/pflag v1.0.10
)

replace github.com/go-testfixtures/testfixtures/v3 => ../..

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
)
