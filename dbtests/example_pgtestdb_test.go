package dbtests

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/peterldowns/pgtestdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamplePgtestdb(t *testing.T) {
	t.Parallel()

	// Set up one database for all tests.
	// You can use any external tool like docker-compose.
	// Here we use testcontainers.
	connString := createPostgreSQLContainer(t)
	db := openDB(t, "postgres", connString)
	require.NoError(t, db.Close())

	setupPgtestdbAndTestfixtures := func(t *testing.T) *sql.DB {
		t.Helper()

		// Parse connection string into pars, which are required by a pgtestdb interface
		u, err := url.Parse(connString)
		require.NoError(t, err)

		host := u.Hostname()
		port := u.Port()
		user := u.User.Username()
		password, _ := u.User.Password()
		database := u.Path[1:] // remove the leading slash

		db := pgtestdb.New(t, pgtestdb.Config{
			DriverName: "pgx",
			Host:       host,
			Port:       port,
			User:       user,
			Password:   password,
			Database:   database,
			Options:    "sslmode=disable",
		}, schemaMigrator{})
		t.Cleanup(func() {
			_ = db.Close()
		})
		fixtures, err := testfixtures.New(
			testfixtures.Database(db),
			testfixtures.Dialect("postgres"),
			testfixtures.Directory("testdata/fixtures_dirs/fixtures2"),
			testfixtures.SkipTableChecksumComputation(), // not needed in this example as we use fixtures only once per database
			testfixtures.UseAlterConstraint(),           // use ALTER CONSTRAINT instead of trigger manipulation, the default seems to not work with pgtestdb
		)
		require.NoError(t, err)
		require.NoError(t, fixtures.Load())
		return db
	}

	// We use subtests here, but in normal code just use Test*** functions.
	t.Run("first test", func(t *testing.T) {
		t.Parallel()
		db := setupPgtestdbAndTestfixtures(t)

		res, err := db.Exec("UPDATE accounts SET currency = 'GBP', balance = 75000 WHERE id = 1")
		require.NoError(t, err)
		rowsAffected, err := res.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		var currency string
		err = db.QueryRow("SELECT currency FROM accounts WHERE id = 1").Scan(&currency)
		require.NoError(t, err)
		assert.Equal(t, "GBP", currency)
	})

	t.Run("second test", func(t *testing.T) {
		t.Parallel()
		db := setupPgtestdbAndTestfixtures(t)

		var currency string
		err := db.QueryRow("SELECT currency FROM accounts WHERE id = 1").Scan(&currency)
		require.NoError(t, err)
		assert.Equal(t, "USD", currency)
	})
}

// schemaMigrator implements the pgtestdb.Migrator interface to load schema
// pgtestdb.Migrator handles migrations on its own
type schemaMigrator struct{}

func (m schemaMigrator) Migrate(ctx context.Context, db *sql.DB, conf pgtestdb.Config) error {
	schema, err := os.ReadFile("testdata/schema/postgresql.sql")
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, string(schema))
	return err
}

func (m schemaMigrator) Hash() (string, error) {
	schema, err := os.ReadFile("testdata/schema/postgresql.sql")
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(schema)
	return fmt.Sprintf("%x", hash), nil
}
