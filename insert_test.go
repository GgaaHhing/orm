package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"web/orm/internal/errs"
	"web/orm/internal/valuer"
	"web/orm/model"
)

func TestInserter_Build(t *testing.T) {
	db := &DB{
		core: core{
			r:       model.NewRegistry(),
			dialect: DialectMySOL,
			creator: valuer.NewReflectValue,
			model:   &model.Model{},
		},
	}
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 插入单行
			name: "single row",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);",
				Args: []any{int64(1), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}},
			},
		},
		{
			// 插入多行
			name: "multi-row",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        2,
				FirstName: "Bob",
				Age:       19,
				LastName:  &sql.NullString{String: "Smith", Valid: true},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{int64(1), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}, int64(2), "Bob", int8(19), &sql.NullString{String: "Smith", Valid: true}},
			},
		},
		{
			// 指定列
			name: "specify columns",
			q: NewInserter[TestModel](db).Values(&TestModel{
				FirstName: "Tom",
				Age:       18,
			}).Columns("FirstName", "Age"),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`first_name`,`age`) VALUES (?,?);",
				Args: []any{"Tom", int8(18)},
			},
		},
		{
			// 没有值
			name:    "no values",
			q:       NewInserter[TestModel](db),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			// 使用 Upsert
			name: "on duplicate key",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
			}).OnDuplicateKey().Update(Assign("Age", 19)),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `age`=?;",
				Args: []any{int64(1), "Tom", int8(18), (*sql.NullString)(nil), 19},
			},
		},
		{
			// 使用 Upsert 和 Columns
			name: "on duplicate key with columns",
			q: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
			}).Columns("Id", "FirstName", "Age").OnDuplicateKey().Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`) VALUES (?,?,?) ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`age`=VALUES(`age`);",
				Args: []any{int64(1), "Tom", int8(18)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}
