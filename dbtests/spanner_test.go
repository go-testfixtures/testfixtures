package dbtests

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/go-testfixtures/testfixtures/v3/shared"
	spannerdriver "github.com/googleapis/go-sql-spanner"
)

func TestSpanner(t *testing.T) {
	t.Parallel()

	emulatorEndpoint := createSpannerContainer(t)

	dialect := "spanner"
	db := openSpannerDB(t, emulatorEndpoint)
	loadSchemaInBatchesBySplitter(t, db, "testdata/schema/spanner.sql", []byte(";\n"))
	additionalOptions := []func(*testfixtures.Loader) error{testfixtures.DangerousSkipTestDatabaseCheck()}

	t.Run("standard suite of tests", func(t *testing.T) {
		testLoader(t, db, dialect, additionalOptions...)
	})

	t.Run("SpannerConstraints", func(t *testing.T) {
		options := append(
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Files(
					"testdata/fixtures/accounts.yml",
					"testdata/fixtures/transactions.yml",
					"testdata/fixtures/assets.yml",
					"testdata/fixtures/users.yml",
					"testdata/fixtures/posts.yml",
					"testdata/fixtures/comments.yml",
					"testdata/fixtures/tags.yml",
					"testdata/fixtures/posts_tags.yml",
					"testdata/fixtures/votes.yml",
				),
				testfixtures.SkipTableChecksumComputation(),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}

		constraintsBefore, _ := shared.GetConstraints(db)

		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		constraintsAfter, _ := shared.GetConstraints(db)

		assertSpannerConstraints(t, constraintsBefore, constraintsAfter)

		// Call load again to test against a database with existing data.
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		assertFixturesLoaded(t, db)
	})

	t.Run("SpannerMultiTablesWithInterleavedTables", func(t *testing.T) {
		options := append(
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.FilesMultiTables(
					"testdata/fixtures_multi_tables/users.yml",
					"testdata/fixtures_multi_tables/posts_comments.yml",
					"testdata/fixtures_multi_tables/posts_tags.yml",
					"testdata/fixtures_multi_tables/tags.yml",
					"testdata/fixtures_multi_tables/accounts_transactions.yml", // Transactions is interleaved in Accounts
					"testdata/fixtures_multi_tables/assets.yml",
				),
				testfixtures.SkipTableChecksumComputation(),
			},
			additionalOptions...,
		)
		l, err := testfixtures.New(options...)
		if err != nil {
			t.Errorf("failed to create Loader: %v", err)
			return
		}

		constraintsBefore, _ := shared.GetConstraints(db)

		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		constraintsAfter, _ := shared.GetConstraints(db)

		assertSpannerConstraints(t, constraintsBefore, constraintsAfter)

		// Call load again to test against a database with existing data.
		if err := l.Load(); err != nil {
			t.Errorf("cannot load fixtures: %v", err)
		}

		assertFixturesLoaded(t, db)
	})

	t.Run("DirectoryNotSupported", func(t *testing.T) {
		options := append(
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Directory("testdata/fixtures_dirs/fixtures1"),
				testfixtures.SkipTableChecksumComputation(),
			},
			additionalOptions...,
		)
		_, err := testfixtures.New(options...)

		if err.Error() != fmt.Sprintf(shared.ErrorMessage_NotSupportedLoadingMethod, "Directory") {
			t.Errorf("error should be %s:", fmt.Sprintf(shared.ErrorMessage_NotSupportedLoadingMethod, "Directory"))
		}
	})

	t.Run("PathsNotSupported", func(t *testing.T) {
		options := append(
			[]func(*testfixtures.Loader) error{
				testfixtures.Database(db),
				testfixtures.Dialect(dialect),
				testfixtures.Template(),
				testfixtures.TemplateData(map[string]interface{}{
					"PostIds": []int{1, 2},
					"TagIds":  []int{1, 2, 3},
				}),
				testfixtures.Paths("testdata/fixtures_dirs/fixtures1"),
				testfixtures.SkipTableChecksumComputation(),
			},
			additionalOptions...,
		)
		_, err := testfixtures.New(options...)

		if err.Error() != fmt.Sprintf(shared.ErrorMessage_NotSupportedLoadingMethod, "Paths") {
			t.Errorf("error should be: %s", fmt.Sprintf(shared.ErrorMessage_NotSupportedLoadingMethod, "Paths"))
		}
	})
}

func openSpannerDB(t *testing.T, emulatorEndpoint string) *sql.DB {
	t.Helper()

	config := spannerdriver.ConnectorConfig{
		Host:               emulatorEndpoint,
		Project:            "test-project",
		Instance:           "test-instance",
		Database:           "testdb",
		AutoConfigEmulator: true,
	}

	connector, err := spannerdriver.CreateConnector(config)
	if err != nil {
		t.Fatalf("Failed to create Spanner connector: %v", err)
	}

	// This is much more convenient than the standard sql.Open, because CreateConnector also create a database.
	db := sql.OpenDB(connector)
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to connect to Spanner database: %v", err)
	}

	return db
}

func assertSpannerConstraints(t *testing.T, constraintsBefore, constraintsAfter map[string][]shared.SpannerConstraint) {
	if len(constraintsBefore) != len(constraintsAfter) {
		t.Errorf("constraints before and after should have the same length")
	}

	for key, cBefore := range constraintsBefore {
		cAfter := constraintsAfter[key]
		if len(cBefore) != len(cAfter) {
			t.Errorf("constraints before and after should have the same length")
		}
		for _, cBefore := range cBefore {
			found := false
			for _, cAfter := range cAfter {
				if reflect.DeepEqual(cBefore, cAfter) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("constraint %s not found in constraintsAfter", cBefore.ConstraintName)
			}
		}
	}
}
