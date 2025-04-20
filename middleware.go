package orm

import (
	"context"
	"web/orm/model"
)

// QueryContext 存放一些查询之后得到的信息
type QueryContext struct {
	// 查询类型，标记增删改查
	Type string
	// 代表查询本身
	Builder QueryBuilder

	Model *model.Model
}

type QueryResult struct {
	Result any
	Err    error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

type Middleware func(next Handler) Handler
