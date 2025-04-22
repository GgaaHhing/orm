//go:build e2e

package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"web/orm"
	"web/orm/test"
)

type SelectSuite struct {
	Suite
}

func (s *SelectSuite) TearDownSuite() {
	orm.RawQuery[test.SimpleStruct](s.db, "TRUNCATE TABLE `simple_struct`").Exec(context.Background())
}

func (s *SelectSuite) SetupTest() {
	// 奇技淫巧，先初始化，然后我们再插入一个数据，
	// 因为如果我们不插入数据，Get拿不到数据
	s.SetupSuite()
	orm.NewInserter[test.SimpleStruct](s.db).Values(test.NewSimpleStruct(12))
}

func TestMySQLSelect(t *testing.T) {
	suite.Run(t, &SelectSuite{
		Suite: Suite{
			driver: "mysql",
			dsn:    "root:root@tcp(localhost:13306)/integration_test",
		},
	})
}

func (s *SelectSuite) TestGet() {
	testCases := []struct {
		name    string
		s       *orm.Selector[test.SimpleStruct]
		wantRes *test.SimpleStruct
		wantErr error
	}{
		{
			name: "get data",
			s:    orm.NewSelector[test.SimpleStruct](s.db).Where(orm.C("Id").Eq(12)),
		},
		{
			name:    "error ormNoRow",
			s:       orm.NewSelector[test.SimpleStruct](s.db).Where(orm.C("Id").Eq(120)),
			wantErr: orm.ErrNoRows,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			res, err := tc.s.Get(ctx)
			assert.Equal(s.T(), tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
