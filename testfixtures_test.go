package testfixtures

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
)

func TestFixtureFile(t *testing.T) {
	f := &fixtureFile{fileName: "posts.yml"}
	file := f.fileNameWithoutExtension()
	if file != "posts" {
		t.Errorf("Should be 'posts', but returned %s", file)
	}
}

func TestRequiredOptions(t *testing.T) {
	t.Run("DatabaseIsRequired", func(t *testing.T) {
		_, err := New()
		if !errors.Is(err, errDatabaseIsRequired) {
			t.Error("should return an error if database if not given")
		}
	})

	t.Run("DialectIsRequired", func(t *testing.T) {
		_, err := New(Database(&sql.DB{}))
		if !errors.Is(err, errDialectIsRequired) {
			t.Error("should return an error if dialect if not given")
		}
	})
}

func TestQuoteKeyword(t *testing.T) {
	tests := []struct {
		helper   helper
		keyword  string
		expected string
	}{
		{&postgreSQL{}, `posts_tags`, `"posts_tags"`},
		{&postgreSQL{}, `"posts.tags"`, `"posts.tags"`},
		{&postgreSQL{}, `test_schema.posts_tags`, `"test_schema"."posts_tags"`},
		{&sqlserver{}, `posts_tags`, `[posts_tags]`},
		{&sqlserver{}, `test_schema.posts_tags`, `[test_schema].[posts_tags]`},
	}

	for _, test := range tests {
		actual := test.helper.quoteKeyword(test.keyword)

		if test.expected != actual {
			t.Errorf("TestQuoteKeyword keyword %s should have escaped to %s. Received %s instead", test.keyword, test.expected, actual)
		}
	}
}

func TestEnsureTestDatabase(t *testing.T) {
	tests := []struct {
		name           string
		isTestDatabase bool
	}{
		{"db_test", true},
		{"dbTEST", true},
		{"testdb", true},
		{"production", false},
		{"productionTestCopy", true},
		{"t_e_s_t", false},
		{"ТESТ", false}, // cyrillic T
	}

	for _, it := range tests {
		var (
			mockedHelper = NewMockHelper(it.name)
			l            = &Loader{helper: mockedHelper}
			err          = l.EnsureTestDatabase()
		)
		if err != nil && it.isTestDatabase {
			t.Errorf("EnsureTestDatabase() should return nil for name = %s", it.name)
		}
		if err == nil && !it.isTestDatabase {
			t.Errorf("EnsureTestDatabase() should return error for name = %s", it.name)
		}
	}
}

func TestExtractOrderedTablesFromYaml(t *testing.T) {
	yamlData := []byte(`
accounts:
	- SomeField: SomeValue
	- SomeField: SomeValue

transactions:
	- SomeField: SomeValue
	- SomeField: SomeValue

users:
	- SomeField: SomeValue
	- SomeField: SomeValue`)

	orderedTables, err := extractOrderedTablesFromYaml(yamlData)
	if err != nil {
		t.Errorf("Error extracting ordered tables from yaml: %s", err)
	}

	if len(orderedTables) != 3 {
		t.Errorf("Expected 3 tables, got %d", len(orderedTables))
	}

	if !reflect.DeepEqual(orderedTables, []string{"accounts", "transactions", "users"}) {
		t.Errorf("Expected 'accounts', 'transactions', 'users' to be the tables, got %v", orderedTables)
	}
}
