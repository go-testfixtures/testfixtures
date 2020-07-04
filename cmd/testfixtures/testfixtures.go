package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/pflag"

	"github.com/go-testfixtures/testfixtures/v3"
)

var version = "master"

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	var (
		versionFlag           bool
		dialect               string
		connString            string
		dir                   string
		files                 []string
		paths                 []string
		useAlterContraint     bool
		skipResetSequences    bool
		resetSequencesTo      int64
		skipTestDatabaseCheck bool
	)

	pflag.BoolVar(&versionFlag, "version", false, "show testfixtures version")
	pflag.StringVarP(&dialect, "dialect", "d", "", "which database system you're using (postgres, timescaledb, mysql, mariadb, sqlite or sqlserver)")
	pflag.StringVarP(&connString, "conn", "c", "", "a database connection string")
	pflag.StringVarP(&dir, "dir", "D", "", "a directory of YAML fixtures to load")
	pflag.StringSliceVarP(&files, "files", "f", nil, "a list of YAML files to load")
	pflag.StringSliceVarP(&paths, "paths", "p", nil, "a list of fixture paths to load (directory or file)")
	pflag.BoolVar(&useAlterContraint, "alter-constraint", false, "use ALTER CONSTRAINT to disable referential integrity (PostgreSQL only)")
	pflag.BoolVar(&skipResetSequences, "no-reset-sequences", false, "skip reset of sequences after loading (PostgreSQL only)")
	pflag.Int64Var(&resetSequencesTo, "reset-sequences-to", 0, "sets the number sequences will be reset after loading fixtures (PostgreSQL only, defaults to 10000)")
	pflag.BoolVar(&skipTestDatabaseCheck, "dangerous-no-test-database-check", false, `skips check for "test" in database name (use with caution)`)
	pflag.Parse()

	if versionFlag {
		log.Printf("testfixtures version: %s", version)
		return
	}

	if dialect == "" && connString == "" {
		log.Fatal("testfixtures: both --dialect (-d) and --conn (-c) are required")
		return
	}
	if dir == "" && len(files) == 0 && len(paths) == 0 {
		log.Fatal("testfixtures: either --dir (-D) or --files (-f) or --paths (-p) need to be given")
		return
	}

	driver := dialect
	if driver == "cockroach" || driver == "cockroachdb" {
		driver = "postgres"
	}

	dialect, err := getDialect(dialect)
	if err != nil {
		log.Fatal(err)
		return
	}

	db, err := sql.Open(driver, connString)
	if err != nil {
		log.Fatalf("testfixtures: could not connect to database: %v", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("testfixtures: could not ping database: %v", err)
		return
	}

	options := []func(*testfixtures.Loader) error{
		testfixtures.Database(db),
		testfixtures.Dialect(dialect),
	}
	if dir != "" {
		options = append(options, testfixtures.Directory(dir))
	}
	if len(files) > 0 {
		options = append(options, testfixtures.Files(files...))
	}
	if len(paths) > 0 {
		options = append(options, testfixtures.Paths(paths...))
	}
	if useAlterContraint {
		options = append(options, testfixtures.UseAlterConstraint())
	}
	if skipResetSequences {
		options = append(options, testfixtures.SkipResetSequences())
	}
	if resetSequencesTo > 0 {
		options = append(options, testfixtures.ResetSequencesTo(resetSequencesTo))
	}
	if skipTestDatabaseCheck {
		options = append(options, testfixtures.DangerousSkipTestDatabaseCheck())
	}

	loader, err := testfixtures.New(options...)
	if err != nil {
		log.Fatal(err)
	}
	if err := loader.Load(); err != nil {
		log.Fatal(err)
	}
	log.Printf("testfixtures: fixtures loaded successfully")
}

func getDialect(dialect string) (string, error) {
	switch dialect {
	case "postgres", "postgresql", "timescaledb":
		return "postgres", nil
	case "cockroach", "cockroachdb":
		return "cockroachdb", nil
	case "mysql", "mariadb":
		return "mysql", nil
	case "sqlite", "sqlite3":
		if !isSQLiteSupported() {
			return "", fmt.Errorf("testfixtures: SQLite is not supported in this build")
		}
		return "sqlite3", nil
	case "mssql", "sqlserver":
		return "sqlserver", nil
	default:
		return "", fmt.Errorf(`testfixtures: unrecognized dialect "%s"`, dialect)
	}
}

func isSQLiteSupported() bool {
	for _, d := range sql.Drivers() {
		if d == "sqlite3" {
			return true
		}
	}
	return false
}
