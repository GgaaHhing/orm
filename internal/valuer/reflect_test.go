package valuer

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"testing"
	"web/orm/model"
)

func TestReflectValue(t *testing.T) {
	testCases := []struct {
		name       string
		entity     any
		rows       func() *sqlmock.Rows
		wantErr    error
		wantEntity any
	}{
		{
			name:   "order",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"id"}).AddRow(1)
			},
			wantEntity: &TestModel{
				Id: 1,
			},
			wantErr: nil,
		},
	}

	r := model.NewRegistry()
	// 构建mock
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// 通过mock构建出想要的rows
			mockRows := tc.rows()
			mock.ExpectQuery("SELECT XX").WillReturnRows(mockRows)
			rows, err := mockDB.Query("SELECT XX")
			require.NoError(t, err)
			rows.Next()

			m, err := r.Get(tc.entity)
			require.NoError(t, err)
			val := NewReflectValue(m, tc.entity)
			err = val.SetColumn(rows)
			require.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			require.Equal(t, tc.wantEntity, tc.entity)
		})
	}
}

type TestModel struct {
	Id int64
}
