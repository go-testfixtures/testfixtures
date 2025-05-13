// Shared code between implementation and tests. DO NOT IMPORT IT as the contract may change between versions
package shared

import (
	"database/sql"
)

type Queryable interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}

type SpannerConstraint struct {
	TableName        string
	ConstraintName   string
	ColumnName       string
	Position         int
	ReferencedTable  string
	ReferencedColumn string
}

func GetConstraints(q Queryable) (map[string][]SpannerConstraint, error) {
	var constraints = make(map[string][]SpannerConstraint)

	rows, err := q.Query(SpannerConstraintsQuery)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var constraint SpannerConstraint
		if err = rows.Scan(
			&constraint.TableName,
			&constraint.ConstraintName,
			&constraint.ColumnName,
			&constraint.Position,
			&constraint.ReferencedTable,
			&constraint.ReferencedColumn,
		); err != nil {
			return nil, err
		}

		if constraints[constraint.ConstraintName] == nil {
			constraints[constraint.ConstraintName] = []SpannerConstraint{}
		}
		constraints[constraint.ConstraintName] = append(constraints[constraint.ConstraintName], constraint)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return constraints, nil
}

const SpannerConstraintsQuery = `
	SELECT 
			tc.TABLE_NAME AS table_name,
			tc.CONSTRAINT_NAME AS constraint_name,
			kcu.COLUMN_NAME AS column_name,
			kcu.ORDINAL_POSITION AS position,
			kcu2.TABLE_NAME AS referenced_table,
			kcu2.COLUMN_NAME AS referenced_column
		FROM information_schema.TABLE_CONSTRAINTS tc
		JOIN information_schema.KEY_COLUMN_USAGE kcu
			ON tc.CONSTRAINT_SCHEMA = kcu.CONSTRAINT_SCHEMA
			AND tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
		JOIN information_schema.REFERENTIAL_CONSTRAINTS rc
			ON tc.CONSTRAINT_SCHEMA = rc.CONSTRAINT_SCHEMA
			AND tc.CONSTRAINT_NAME = rc.CONSTRAINT_NAME
		JOIN information_schema.KEY_COLUMN_USAGE kcu2
			ON rc.UNIQUE_CONSTRAINT_SCHEMA = kcu2.CONSTRAINT_SCHEMA
			AND rc.UNIQUE_CONSTRAINT_NAME = kcu2.CONSTRAINT_NAME
			AND kcu.ORDINAL_POSITION = kcu2.ORDINAL_POSITION
		WHERE tc.CONSTRAINT_TYPE = 'FOREIGN KEY'
		ORDER BY tc.TABLE_NAME, tc.CONSTRAINT_NAME, kcu.ORDINAL_POSITION;
`
