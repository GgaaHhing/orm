package promethus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"web/orm"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  m.Namespace,
		Subsystem:  m.Subsystem,
		Name:       m.Name,
		Help:       m.Help,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"type", "table"})
	prometheus.MustRegister(vector)

	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				vector.WithLabelValues(qc.Type, qc.Model.TableName).Observe(float64(time.Since(startTime).Milliseconds()))
			}()
			return next(ctx, qc)
		}
	}

}
