package testfixtures

import (
	"fmt"
	"strings"
)

// Labels lets fixtures name their rows and refer to each other by name,
// the way Rails fixtures do, instead of by hand-picked numeric ids.
//
// A row is named by writing a string where its primary key goes. Another row
// refers to it by writing the same string in a foreign key column:
//
//	# users.yml
//	- id: alice
//	  name: Alice
//
//	# posts.yml
//	- id: hello_world
//	  author_id: alice
//
// The loader gives every named row an integer id before anything is inserted
// and replaces each label with it. Ids count up from 1 in the order the rows
// are loaded (files in the order they are given, rows in the order they are
// written), so a row written first is older than one written later, and no two
// rows of any table share an id.
//
// A label is any string found in a column called "id" or ending in "_id". The
// "id" column defines a label; the others refer to one, and a label that no
// row defines or that is defined twice is an error rather than a quiet
// mismatch. Numeric values are left alone, so a fixture can still mix both
// styles. The option therefore suits integer keys: do not use it for tables
// whose "id" or "*_id" columns hold strings (UUIDs, for example).
//
// Write rows as a list: a fixture written as a map of rows is read in no
// particular order, so its ids would not be reproducible.
//
// Labels are resolved after templating, so they can be produced by Template().
// Use Loader.ID to get the id of a labelled row in a test.
func Labels() func(*Loader) error {
	return func(l *Loader) error {
		l.labels = map[string]int{}
		return nil
	}
}

// ID returns the id the loader gave the row labelled label. It is an error for
// a label no row defines, or when the Labels() option was not given.
func (l *Loader) ID(label string) (int, error) {
	if l.labels == nil {
		return 0, fmt.Errorf("testfixtures: the Labels() option is required in order to use ID()")
	}
	id, ok := l.labels[label]
	if !ok {
		return 0, fmt.Errorf("testfixtures: label %q is not defined by any fixture row", label)
	}
	return id, nil
}

func isLabelColumn(name string) bool {
	return name == "id" || strings.HasSuffix(name, "_id")
}

// resolveLabels numbers the rows named by a string "id" and replaces every
// label in the records with its id. files and records are parallel: records[i]
// are the rows of files[i].
func (l *Loader) resolveLabels(files []*fixtureFile, records [][]map[string]any) error {
	l.labels = map[string]int{}
	next := 1

	for i, rows := range records {
		for n, row := range rows {
			label, ok := row["id"].(string)
			if !ok {
				continue
			}
			if _, dup := l.labels[label]; dup {
				return fmt.Errorf("testfixtures: label %q is defined twice (again in %s, index %d)", label, files[i].fileName, n)
			}
			l.labels[label] = next
			next++
		}
	}

	for i, rows := range records {
		for n, row := range rows {
			for column, value := range row {
				label, ok := value.(string)
				if !ok || !isLabelColumn(column) {
					continue
				}
				id, ok := l.labels[label]
				if !ok {
					return fmt.Errorf("testfixtures: label %q in column %q is not defined by any fixture row (file %s, index %d)", label, column, files[i].fileName, n)
				}
				row[column] = id
			}
		}
	}

	return nil
}
