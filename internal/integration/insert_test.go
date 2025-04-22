//go:build e2e

// 标记一下是集成测试
package integration

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// 测试辅助工具包，主要用于编写更结构化和可组织的测试代码
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"web/orm"
	"web/orm/test"
)

// InsertSuite
type InsertSuite struct {
	Suite
}

func (i *InsertSuite) SetupTest() {
	i.SetupSuite()
}

func (i *InsertSuite) TearDownTest() {
	orm.RawQuery[test.SimpleStruct](i.db, "TRUNCATE TABLE `simple_struct`").Exec(context.Background())
}

func TestMysqlInsert(t *testing.T) {
	suite.Run(t, &InsertSuite{
		Suite: Suite{
			driver: "mysql",
			dsn:    "root:root@tcp(localhost:13306)/integration_test",
		},
	})
}

func (i *InsertSuite) TestInsert() {
	db := i.db
	// t 从suite中获取
	t := i.T()
	testCases := []struct {
		name         string
		i            *orm.Inserter[test.SimpleStruct]
		wantAffected int64 // 插入了多少行
	}{
		{
			name:         "insert one",
			i:            orm.NewInserter[test.SimpleStruct](db).Values(test.NewSimpleStruct(12)),
			wantAffected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			res := tc.i.Exec(ctx)
			require.NoError(t, res.Err())
			affected, err := res.RowsAffected()
			require.NoError(t, err)
			assert.Equal(t, tc.wantAffected, affected)
		})
	}
}

//// TestMysqlInsert 方法一、只需要更改driver和dsn就能修改对应的数据库
//func TestMysqlInsert(t *testing.T) {
//	testInsert(t, "mysql", "root:root@tcp(Localhost:13306)/integration_test")
//}
//
//func testInsert(t *testing.T, driver, dsn string) {
//	db, err := orm.Open(driver, dsn)
//	require.NoError(t, err)
//	testCases := []struct {
//		name         string
//		i            *orm.Inserter[test.SimpleStruct]
//		wantAffected int64 // 插入了多少行
//	}{
//		{
//			name:         "insert one",
//			i:            orm.NewInserter[test.SimpleStruct](db).Values(test.NewSimpleStruct(12)),
//			wantAffected: 1,
//		},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
//			defer cancel()
//			res := tc.i.Exec(ctx)
//			require.NoError(t, res.Err())
//			affected, err := res.RowsAffected()
//			require.NoError(t, err)
//			assert.Equal(t, tc.wantAffected, affected)
//		})
//	}
//}
