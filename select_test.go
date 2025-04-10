package orm

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"web/orm/internal/valuer"
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

func TestSelector_GetMulti(t *testing.T) {
	// 创建一个模拟的数据库连接
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer mockDB.Close()

	testCases := []struct {
		name     string
		query    string
		mockRows *sqlmock.Rows
		wantRes  []*TestModel
		wantErr  error
	}{
		{
			name:  "查询多行",
			query: "SELECT \\* FROM `test_model`;",
			mockRows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				// 使用 sql.NullString 值类型
				rows.AddRow(1, "张三", 18, sql.NullString{String: "Zhang", Valid: true})
				rows.AddRow(2, "李四", 19, sql.NullString{String: "Li", Valid: true})
				rows.AddRow(3, "王五", 20, sql.NullString{String: "", Valid: false})
				return rows
			}(),
			wantRes: []*TestModel{
				{
					Id:        1,
					FirstName: "张三",
					Age:       18,
					LastName:  &sql.NullString{String: "Zhang", Valid: true},
				},
				{
					Id:        2,
					FirstName: "李四",
					Age:       19,
					LastName:  &sql.NullString{String: "Li", Valid: true},
				},
				{
					Id:        3,
					FirstName: "王五",
					Age:       20,
					LastName:  &sql.NullString{String: "", Valid: false},
				},
			},
		},
		{
			name:     "空结果集",
			query:    "SELECT \\* FROM `test_model`;",
			mockRows: sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"}),
			wantRes:  []*TestModel{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectQuery(tc.query).WillReturnRows(tc.mockRows)

			db := &DB{
				db:      mockDB,
				r:       model.NewRegistry(),
				creator: valuer.NewReflectValue,
			}

			_, err = db.r.Register(&TestModel{})
			if err != nil {
				t.Fatal(err)
			}

			selector := NewSelector[TestModel](db)
			res, err := selector.GetMulti(context.Background())

			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, len(tc.wantRes), len(res))

			for i, wantRow := range tc.wantRes {
				assert.Equal(t, wantRow.Id, res[i].Id)
				assert.Equal(t, wantRow.FirstName, res[i].FirstName)
				assert.Equal(t, wantRow.Age, res[i].Age)

				// 修改比较逻辑
				if wantRow.LastName != nil {
					// 直接比较 LastName 的内容，而不是指针
					nullStr := sql.NullString{
						String: wantRow.LastName.String,
						Valid:  wantRow.LastName.Valid,
					}
					if res[i].LastName != nil {
						assert.Equal(t, nullStr, *res[i].LastName)
					}
				}
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("有未满足的期望: %s", err)
			}
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
