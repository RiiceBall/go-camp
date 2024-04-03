package prometheus

import (
	"context"
	"time"
	"webook/internal/service/sms"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func (p *PrometheusDecorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		p.vector.WithLabelValues(tplId).
			Observe(float64(duration))
	}()
	return p.svc.Send(ctx, tplId, args, numbers...)
}

func NewPrometheusDecorator(svc sms.Service, opt prometheus.SummaryOpts) *PrometheusDecorator {
	vector := prometheus.NewSummaryVec(opt, []string{"tpl_id"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		vector: vector,
		svc:    svc,
	}
}
