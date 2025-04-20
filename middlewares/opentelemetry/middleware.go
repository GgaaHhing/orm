package opentelemetry

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"web/orm"
)

const instrumentationName = "orm/middlewares/openTelemetry"

type MiddlewareBuilder struct {
	Tracer trace.Tracer
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {

			spanName := fmt.Sprintf("%s_%s", qc.Type, qc.Model.TableName)
			spanCtx, span := m.Tracer.Start(ctx, spanName)
			defer span.End()

			span.SetAttributes(attribute.String("table", spanName))

			res := next(spanCtx, qc)
			if res.Err != nil {
				span.RecordError(res.Err)
			}
			return res
		}
	}
}
