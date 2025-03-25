package orm

import (
	"context"
	"database/sql"
)

// Querier 用于查询
type Querier[T any] interface {
	// Get 使用指针可以避开一些在反射的问题
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Executor 用于增删改
type Executor interface {
	Exec(ctx context.Context) (sql.Result, error)
}

type QueryBuilder interface {
	// Build 这里传指针是为了方便在AOP里直接进行修改
	Build() (*Query, error)
}

type Query struct {
	SQL  string
	Args []any
}
