package integration

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"web/orm"
)

type Suite struct {
	suite.Suite
	db     *orm.DB
	driver string
	dsn    string
}

// SetupSuite 在suite执行之前要执行的代码
func (i *Suite) SetupSuite() {
	db, err := orm.Open(i.driver, i.dsn)
	require.NoError(i.T(), err)
	db.Wait()
	i.db = db
}
