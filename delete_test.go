package orm

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"web/orm/internal/errs"
	"web/orm/model"
)

func TestDeleter_Build(t *testing.T) {
	r := &DB{
		r: model.NewRegistry(),
	}
	testCase := []struct {
		name    string
		builder DeleteBuilder

		wantErr   error
		wantQuery *Query
	}{
		{
			name: "success",
			builder: (&Deleter[TestModel]{
				db: r,
			}).Where(C("Age").Eq(12)),
			wantErr: nil,
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `age` = ?;",
				Args: []interface{}{12},
			},
		},
		{
			name: "delete from",
			builder: (&Deleter[TestModel]{
				db: r,
			}).From("`TestModel`"),
			wantQuery: &Query{
				SQL:  "DELETE FROM `TestModel`;",
				Args: nil,
			},
			wantErr: errs.ErrDeleteALL,
		},
		{
			name: "empty from",
			builder: (&Deleter[TestModel]{
				db: r,
			}).From(""),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model`;",
				Args: nil,
			},
			wantErr: errs.ErrDeleteALL,
		},
		{
			name: "long where",
			builder: (&Deleter[TestModel]{
				db: r,
			}).Where(C("Age").Eq(12).And(C("FirstName").Eq("John"))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{12, "John"},
			},
		},
		{
			name:    "Not",
			builder: (&Deleter[TestModel]{db: r}).Where(Not(C("Age").Eq(12))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{12},
			},
		},
		{
			name:    "invalid column",
			builder: (&Deleter[TestModel]{db: r}).Where(C("InvalidColumn").Eq(12)),
			wantErr: errs.NewErrUnknownField("InvalidColumn"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			build, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, build)
		})
	}
}
