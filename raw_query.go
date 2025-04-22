package orm

import (
	"context"
	"database/sql"
)

type RawQuerier[T any] struct {
	core
	sess Session
	sql  string
	args []any
}

func (s RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL:  s.sql,
		Args: s.args,
	}, nil
}

// RawQuery 初始化方法
func RawQuery[T any](sess Session, query string, args ...any) *RawQuerier[T] {
	c := sess.getCore()
	return &RawQuerier[T]{
		sess: sess,
		core: c,
		sql:  query,
		args: args,
	}
}

func (s RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	res := get[T](ctx, s.sess, s.core, &QueryContext{
		Type:    "RAW",
		Builder: s,
		Model:   s.model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (s RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s RawQuerier[T]) Exec(ctx context.Context) Result {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}
	res := exec(ctx, s.sess, s.core, &QueryContext{
		Type:    "RAW",
		Builder: s,
		Model:   s.model,
	})

	var sqlRes sql.Result
	if res.Result != nil {
		sqlRes = res.Result.(sql.Result)
	}
	return Result{
		err: res.Err,
		res: sqlRes,
	}
}
