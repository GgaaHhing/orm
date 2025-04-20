package querylog

import (
	"context"
	"log"
	"web/orm"
)

type MiddlewareBuilder struct {
	// 允许用户使用自己的log输出方式
	logFunc func(query string, args []any)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		logFunc: func(query string, args []any) {
			log.Printf("SQL:query: %s args: %v \n", query, args)
		},
	}
}

func (m *MiddlewareBuilder) LogFunc(fn func(query string, args []any)) *MiddlewareBuilder {
	m.logFunc = fn
	return m
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, queryCtx *orm.QueryContext) *orm.QueryResult {
			q, err := queryCtx.Builder.Build()
			if err != nil {
				log.Println("orm: 构造SQL出错", err)
				return &orm.QueryResult{
					Err: err,
				}
			}

			m.logFunc(q.SQL, q.Args)

			res := next(ctx, queryCtx)
			return res
		}
	}
}
