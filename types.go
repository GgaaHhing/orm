package orm

import (
	"context"
)

// Querier 用于查询
type Querier[T any] interface {
	// Get 使用指针可以避开一些在反射的问题
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Executor 用于增删改的注册，也就是和数据库的交互
type Executor interface {
	Exec(ctx context.Context) Result
}

type QueryBuilder interface {
	// Build 这里传指针是为了方便在AOP里直接进行修改
	Build() (*Query, error)
}

type DeleteBuilder interface {
	Build() (*Query, error)
}

type Query struct {
	// 占位符，Args中的参数通过占位符传入SQL中
	SQL  string
	Args []any
}
