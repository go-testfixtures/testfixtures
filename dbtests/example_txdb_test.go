package dbtests

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleTxdb(t *testing.T) {
	t.Parallel()

	// Set up one database for all tests.
	// You can use any external tool like docker-compose.
	// Here we use testcontainers.
	connString := createPostgreSQLContainer(t)
	db := openDB(t, "postgres", connString)
	loadSchemaInOneQuery(t, db, "testdata/schema/postgresql.sql")
	require.NoError(t, db.Close())

	setupTxdbAndTestfixtures := func(t *testing.T) *sql.DB {
		t.Helper()

		// Create a txdb wrapper over the original database.
		db := sql.OpenDB(txdb.New("pgx", connString))
		// and use it for setup and tests
		t.Cleanup(func() {
			_ = db.Close() // remember to close the db wrapper to ROLLBACK the transaction
		})

		fixtures, err := testfixtures.New(
			testfixtures.Database(db),
			testfixtures.Dialect("postgres"),
			testfixtures.Directory("testdata/fixtures_dirs/fixtures2"),
			testfixtures.SkipTableChecksumComputation(), // not needed in this example as we use fixtures only once per database
		)
		require.NoError(t, err)
		require.NoError(t, fixtures.Load())
		return db
	}

	// We use subtests here, but in normal code just use Test*** functions.
	t.Run("first test", func(t *testing.T) {
		t.Parallel()
		db := setupTxdbAndTestfixtures(t)

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
		db := setupTxdbAndTestfixtures(t)

		var currency string
		err := db.QueryRow("SELECT currency FROM accounts WHERE id = 1").Scan(&currency)
		require.NoError(t, err)
		assert.Equal(t, "USD", currency)
	})
}
