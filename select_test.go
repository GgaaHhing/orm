package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"web/orm/model"
)

func TestSelect_Build(t *testing.T) {
	r := &DB{
		r: model.NewRegistry(),
	}
	testCase := []struct {
		name string

		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "select no from",
			//
			builder: &Selector[TestModel]{
				db: r,
			},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name: "select from",
			builder: (&Selector[TestModel]{
				db: r,
			}).From("`TestModel`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},
		{
			name: "empty from",
			builder: (&Selector[TestModel]{
				db: r,
			}).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name: "where",
			builder: (&Selector[TestModel]{
				db: r,
			}).Where(C("Age").Eq(12)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{12},
			},
		},
		{
			name: "long where",
			builder: (&Selector[TestModel]{
				db: r,
			}).Where(C("Age").Eq(12).And(C("FirstName").Eq("John"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{12, "John"},
			},
		},
		{
			name: "Not",
			builder: (&Selector[TestModel]{
				db: r,
			}).Where(Not(C("Age").Eq(12))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{12},
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
