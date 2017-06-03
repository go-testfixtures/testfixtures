package testfixtures

import (
	"fmt"
)

// InsertError will be returned if any error happens on database while
// inserting the record
type InsertError struct {
	Err    error
	File   string
	Index  int
	SQL    string
	Params []interface{}
}

func (e *InsertError) Error() string {
	return fmt.Sprintf(
		"testfixtures: error inserting record: %v, on file: %s, index: %d, sql: %s, params: %v",
		e.Err,
		e.File,
		e.Index,
		e.SQL,
		e.Params,
	)
}
