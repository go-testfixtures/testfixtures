package testfixtures

import (
	"database/sql"
	"strings"
	"testing"
)

func newLabelsLoader(t *testing.T, files map[string]string, order ...string) *Loader {
	t.Helper()

	l := &Loader{
		db:     &sql.DB{},
		helper: NewMockHelper("test_db"),
		labels: map[string]int{},
	}
	for _, name := range order {
		l.fixturesFiles = append(l.fixturesFiles, &fixtureFile{
			path:     name,
			fileName: name,
			content:  []byte(files[name]),
		})
	}
	return l
}

func TestLabels(t *testing.T) {
	files := map[string]string{
		// posts come first on purpose: a label may be used before its row.
		"posts.yml": `
- id: hello_world
  author_id: alice
  title: Hello
- id: second_post
  author_id: bob
  editor_id: 7
`,
		"users.yml": `
- id: alice
- id: bob
`,
	}

	t.Run("NumbersRowsInLoadOrderAndResolvesReferences", func(t *testing.T) {
		l := newLabelsLoader(t, files, "posts.yml", "users.yml")
		if err := l.buildInsertSQLs(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := map[string]int{"hello_world": 1, "second_post": 2, "alice": 3, "bob": 4}
		for label, wantID := range want {
			got, err := l.ID(label)
			if err != nil {
				t.Fatalf("ID(%q): %v", label, err)
			}
			if got != wantID {
				t.Errorf("ID(%q) = %d, want %d", label, got, wantID)
			}
		}

		// The first post is written as id=1, author_id=3 (alice).
		params := l.fixturesFiles[0].insertSQLs[0].params
		if !containsParams(params, 1, 3, "Hello") {
			t.Errorf("params of the first post = %v, want 1, 3 and \"Hello\"", params)
		}
		// A numeric value is left alone.
		params = l.fixturesFiles[0].insertSQLs[1].params
		if !containsParams(params, 2, 4, uint64(7)) {
			t.Errorf("params of the second post = %v, want 2, 4 and 7", params)
		}
	})

	t.Run("UnknownLabel", func(t *testing.T) {
		l := newLabelsLoader(t, map[string]string{"posts.yml": "- id: p\n  author_id: nobody\n"}, "posts.yml")
		err := l.buildInsertSQLs()
		if err == nil || !strings.Contains(err.Error(), `label "nobody"`) {
			t.Errorf("want an error naming the label, got %v", err)
		}
	})

	t.Run("DuplicateLabel", func(t *testing.T) {
		l := newLabelsLoader(t, map[string]string{"users.yml": "- id: alice\n- id: alice\n"}, "users.yml")
		err := l.buildInsertSQLs()
		if err == nil || !strings.Contains(err.Error(), "defined twice") {
			t.Errorf("want a duplicate label error, got %v", err)
		}
	})

	t.Run("IDWithoutOptionOrWithUnknownLabel", func(t *testing.T) {
		if _, err := (&Loader{}).ID("alice"); err == nil {
			t.Error("want an error when Labels() was not given")
		}
		l := newLabelsLoader(t, files, "posts.yml", "users.yml")
		if err := l.buildInsertSQLs(); err != nil {
			t.Fatal(err)
		}
		if _, err := l.ID("nobody"); err == nil {
			t.Error("want an error for an undefined label")
		}
	})

	t.Run("OptionTurnsLabelsOn", func(t *testing.T) {
		l := &Loader{}
		if err := Labels()(l); err != nil {
			t.Fatal(err)
		}
		if l.labels == nil {
			t.Error("Labels() should enable label resolution")
		}
	})
}

func containsParams(params []any, want ...any) bool {
	for _, w := range want {
		found := false
		for _, p := range params {
			if p == w || toInt(p) == toInt(w) && toInt(w) != -1 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func toInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case uint64:
		return int(n)
	}
	return -1
}
