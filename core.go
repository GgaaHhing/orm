package orm

import (
	"context"
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

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	// 构造查询
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	rows, err := sess.queryContext(ctx, q.SQL, q.Args)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	if !rows.Next() {
		return &QueryResult{
			Err: ErrNoRows,
		}
	}
	tp := new(T)
	val := c.creator(c.model, tp)
	err = val.SetColumn(rows)
	return &QueryResult{
		Result: tp,
		Err:    err,
	}
}

func exec(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, c, qc)
	}
	for j := len(c.mdls) - 1; j >= 0; j-- {
		root = c.mdls[j](root)
	}
	return root(ctx, qc)
}

func execHandler(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Result: Result{
				err: err,
			},
		}
	}
	res, err := sess.execContext(ctx, q.SQL, q.Args...)
	return &QueryResult{
		Result: Result{
			err: err,
			res: res,
		},
	}
}
