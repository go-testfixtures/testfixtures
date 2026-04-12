package testfixtures

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
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
	t.Run("DialectWithPlaceholder", func(t *testing.T) {
		loader, err := New(Database(&sql.DB{}), Dialect("clickhouse", WithCustomPlaceholder(ParamTypeQuestion)))
		if err != nil {
			t.Error("should return nil error")
		}
		if paramType := loader.helper.paramType(); paramType != ParamTypeQuestion {
			t.Errorf("incorrect param type returned: %s", paramType)
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

func TestLoadPendingSources(t *testing.T) {
	t.Run("SpannerRejectsDirectory", func(t *testing.T) {
		l := &Loader{
			db:                 &sql.DB{},
			helper:             &spanner{},
			templateLeftDelim:  "{{",
			templateRightDelim: "}}",
			templateOptions:    []string{"missingkey=zero"},
			fs:                 defaultFS{},
			pendingSources: []pendingSource{
				{kind: sourceDirectory, paths: []string{"testdata/fixtures_template"}},
			},
		}
		err := l.loadPendingSources()
		wantErr := `
testfixtures: Directory is not supported for Spanner to ensure support for INTERLEAVED tables.
Use Files():
  ensure the order of the files is correct, parents loaded before children or
Use FilesMultiTables():
  and order your table keys in the yaml files from parent to child`
		if err == nil || err.Error() != wantErr {
			t.Errorf("unexpected error\nwant: %s\ngot:  %v", wantErr, err)
		}
	})

	t.Run("SpannerRejectsPaths", func(t *testing.T) {
		l := &Loader{
			db:                 &sql.DB{},
			helper:             &spanner{},
			templateLeftDelim:  "{{",
			templateRightDelim: "}}",
			templateOptions:    []string{"missingkey=zero"},
			fs:                 defaultFS{},
			pendingSources: []pendingSource{
				{kind: sourcePaths, paths: []string{"testdata/fixtures_template"}},
			},
		}
		err := l.loadPendingSources()
		wantErr := `
testfixtures: Paths is not supported for Spanner to ensure support for INTERLEAVED tables.
Use Files():
  ensure the order of the files is correct, parents loaded before children or
Use FilesMultiTables():
  and order your table keys in the yaml files from parent to child`
		if err == nil || err.Error() != wantErr {
			t.Errorf("unexpected error\nwant: %s\ngot:  %v", wantErr, err)
		}
	})

	t.Run("NoPendingSources", func(t *testing.T) {
		l := &Loader{
			db:                 &sql.DB{},
			helper:             NewMockHelper("test_db"),
			templateLeftDelim:  "{{",
			templateRightDelim: "}}",
			templateOptions:    []string{"missingkey=zero"},
			fs:                 defaultFS{},
		}
		err := l.loadPendingSources()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(l.fixturesFiles) != 0 {
			t.Errorf("expected no fixture files, got %d", len(l.fixturesFiles))
		}
	})
}

// Test that Template options work regardless of whether they come
// before or after Directory/Files/Paths options, and that all
// pending source kinds are loaded correctly.
// See: https://github.com/go-testfixtures/testfixtures/pull/349
func TestLoadPendingSourcesTemplateOptionOrdering(t *testing.T) {
	templateData := map[string]interface{}{
		"Ids": []int{1, 2, 3},
	}

	type item struct {
		ID   int    `yaml:"id"`
		Name string `yaml:"name,omitempty"`
	}

	wantItems := []item{
		{ID: 1, Name: "item-1"},
		{ID: 2, Name: "item-2"},
		{ID: 3, Name: "item-3"},
	}

	wantIdsOnly := []item{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	assertSingleTable := func(expected []item) func(*testing.T, *Loader) {
		return func(t *testing.T, l *Loader) {
			t.Helper()
			if got, want := len(l.fixturesFiles), 1; got != want {
				t.Fatalf("fixturesFiles length = %d, want %d", got, want)
			}
			content := l.fixturesFiles[0].content
			if strings.Contains(string(content), "{{") {
				t.Error("content still contains unresolved template marker \"{{\"")
			}

			var got []item
			if err := yaml.Unmarshal(content, &got); err != nil {
				t.Fatalf("yaml.Unmarshal: %v", err)
			}
			if len(got) != len(expected) {
				t.Fatalf("got %d items, want %d", len(got), len(expected))
			}
			for i := range expected {
				if got[i] != expected[i] {
					t.Errorf("item[%d] = %+v, want %+v", i, got[i], expected[i])
				}
			}
		}
	}

	assertMultiTable := func(expectedByFile map[string][]item) func(*testing.T, *Loader) {
		return func(t *testing.T, l *Loader) {
			t.Helper()
			if got, want := len(l.fixturesFiles), len(expectedByFile); got != want {
				t.Fatalf("fixturesFiles length = %d, want %d", got, want)
			}
			for _, f := range l.fixturesFiles {
				if strings.Contains(string(f.content), "{{") {
					t.Errorf("file %s: content still contains unresolved template marker \"{{\"", f.fileName)
				}

				expected, ok := expectedByFile[f.fileName]
				if !ok {
					t.Fatalf("unexpected fixture file name: %s", f.fileName)
				}

				var got []item
				if err := yaml.Unmarshal(f.content, &got); err != nil {
					t.Fatalf("file %s: yaml.Unmarshal: %v", f.fileName, err)
				}
				if len(got) != len(expected) {
					t.Fatalf("file %s: got %d items, want %d", f.fileName, len(got), len(expected))
				}
				for i := range expected {
					if got[i] != expected[i] {
						t.Errorf("file %s: item[%d] = %+v, want %+v", f.fileName, i, got[i], expected[i])
					}
				}
			}
		}
	}
	tests := []struct {
		name    string
		options []func(*Loader) error
		assert  func(*testing.T, *Loader)
	}{
		{
			name: "TemplateBeforeDirectory",
			options: []func(*Loader) error{
				Template(), TemplateData(templateData),
				Directory("testdata/fixtures_template"),
			},
			assert: assertSingleTable(wantItems),
		},
		{
			name: "TemplateAfterDirectory",
			options: []func(*Loader) error{
				Directory("testdata/fixtures_template"),
				Template(), TemplateData(templateData),
			},
			assert: assertSingleTable(wantItems),
		},
		{
			name: "TemplateBeforeFiles",
			options: []func(*Loader) error{
				Template(), TemplateData(templateData),
				Files("testdata/fixtures_template/items.yml"),
			},
			assert: assertSingleTable(wantItems),
		},
		{
			name: "TemplateAfterFiles",
			options: []func(*Loader) error{
				Files("testdata/fixtures_template/items.yml"),
				Template(), TemplateData(templateData),
			},
			assert: assertSingleTable(wantItems),
		},
		{
			name: "TemplateBeforePaths",
			options: []func(*Loader) error{
				Template(), TemplateData(templateData),
				Paths("testdata/fixtures_template"),
			},
			assert: assertSingleTable(wantItems),
		},
		{
			name: "TemplateAfterPaths",
			options: []func(*Loader) error{
				Paths("testdata/fixtures_template"),
				Template(), TemplateData(templateData),
			},
			assert: assertSingleTable(wantItems),
		},
		{
			name: "PathsWithFile",
			options: []func(*Loader) error{
				Template(), TemplateData(templateData),
				Paths("testdata/fixtures_template/items.yml"),
			},
			assert: assertSingleTable(wantItems),
		},
		{
			name: "TemplateBeforeFilesMultiTables",
			options: []func(*Loader) error{
				Template(), TemplateData(templateData),
				FilesMultiTables("testdata/fixtures_template_multi/multi_tables.yml"),
			},
			assert: assertMultiTable(map[string][]item{
				"items.yml":       wantItems,
				"other_items.yml": wantIdsOnly,
			}),
		},
		{
			name: "TemplateAfterFilesMultiTables",
			options: []func(*Loader) error{
				FilesMultiTables("testdata/fixtures_template_multi/multi_tables.yml"),
				Template(), TemplateData(templateData),
			},
			assert: assertMultiTable(map[string][]item{
				"items.yml":       wantItems,
				"other_items.yml": wantIdsOnly,
			}),
		},
		{
			name: "MultipleSources",
			options: []func(*Loader) error{
				Template(), TemplateData(templateData),
				Directory("testdata/fixtures_template"),
				Files("testdata/fixtures_template/items.yml"),
			},
			assert: func(t *testing.T, l *Loader) {
				t.Helper()
				// Directory has 1 file (items.yml), Files adds 1 more
				if got, want := len(l.fixturesFiles), 2; got != want {
					t.Fatalf("fixturesFiles length = %d, want %d", got, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullOpts := []func(*Loader) error{
				Database(&sql.DB{}),
				Dialect("clickhouse"),
			}
			fullOpts = append(fullOpts, tt.options...)
			l, err := New(fullOpts...)
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			tt.assert(t, l)
		})
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
