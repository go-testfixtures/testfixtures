package dbtests

import (
	"database/sql"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/assert"
)

func TestExampleSeparateDatabasePerTest(t *testing.T) {
	t.Parallel()

	// Create a separate database for each test.
	setupDB := func(t *testing.T) *sql.DB {
		t.Helper()

		// Here we use testcontainers.
		connString := createPostgreSQLContainer(t)
		db := openDB(t, "postgres", connString)
		loadSchemaInOneQuery(t, db, "testdata/schema/postgresql.sql")

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
		db := setupDB(t)

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
		db := setupDB(t)

		var currency string
		err := db.QueryRow("SELECT currency FROM accounts WHERE id = 1").Scan(&currency)
		require.NoError(t, err)
		assert.Equal(t, "USD", currency)
	})
}
