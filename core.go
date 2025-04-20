package orm

import (
	"web/orm/internal/valuer"
	"web/orm/model"
)

// core 核心
// 封装了 ORM 框架中最核心的几个组件
// - 代码复用： DB 和 Tx 都需要这些核心组件
// - 关注点分离：将核心功能与具体的数据库操作分开
// - 避免重复：不需要在 DB 和 Tx 中重复定义这些字段
type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       model.Registry
	mdls    []Middleware
}
