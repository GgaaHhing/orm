package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelect_Build(t *testing.T) {
	testCase := []struct {
		name string

		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "select no from",
			//
			builder: &Selector[TestModel]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},
		{
			name:    "select from",
			builder: (&Selector[TestModel]{}).From("`TestModel`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},
		{
			name:    "empty from",
			builder: (&Selector[TestModel]{}).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
