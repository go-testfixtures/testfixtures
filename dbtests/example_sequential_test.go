package dbtests

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleSequential(t *testing.T) {
	t.Parallel()
	// Set up one database for all tests.
	// You can use any external tool like docker-compose.
	// Here we use testcontainers.
	connString := createPostgreSQLContainer(t)
	db := openDB(t, "postgres", connString)
	loadSchemaInOneQuery(t, db, "testdata/schema/postgresql.sql")

	// You can share a single `fixtures` object as a global variable or create a new one for each test.
	// The single 'fixtures' object is recommended as it keeps track of the state to perform fewer db operations.
	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("testdata/fixtures_dirs/fixtures2"),
	)
	require.NoError(t, err)

	// We use subtests here, but in normal code just use Test*** functions.
	t.Run("first test", func(t *testing.T) {
		// no t.Parallel() here

		require.NoError(t, fixtures.Load())

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
		// no t.Parallel() here

		require.NoError(t, fixtures.Load()) // this one will clean the database and load fixtures once again

		var currency string
		err := db.QueryRow("SELECT currency FROM accounts WHERE id = 1").Scan(&currency)
		require.NoError(t, err)
		assert.Equal(t, "USD", currency)
	})
}
