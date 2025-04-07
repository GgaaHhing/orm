package sql_demo

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestSqlMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	defer db.Close()
	require.NoError(t, err)

	mockRows := sqlmock.NewRows([]string{"id"})
	mockRows.AddRow(1)
	mock.ExpectQuery("SELECT id FORM `user` .*").WillReturnRows(mockRows)

	row := db.QueryRowContext(context.Background(), "SELECT id FORM `user` WHERE id=1")
	require.NoError(t, row.Err())
	tm := &TestModel{}
	err = row.Scan(&tm.Id)
	require.NoError(t, err)
	log.Println(tm)
	//for rows.Next() {
	//	tm := &TestModel{}
	//	err = rows.Scan(&tm.Id)
	//	require.NoError(t, err)
	//	log.Println(tm)
	//}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
