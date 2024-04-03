package redisx

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(opt prometheus.SummaryOpts) *PrometheusHook {
	vector := prometheus.NewSummaryVec(opt, []string{"cmd", "key_exist"})
	return &PrometheusHook{
		vector: vector,
	}
}

func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		var err error
		defer func() {
			duration := time.Since(start).Milliseconds()
			keyExist := err == redis.Nil
			// 集群下可以考虑使用 FullName
			p.vector.WithLabelValues(cmd.Name(),
				strconv.FormatBool(keyExist)).
				Observe(float64(duration))
		}()
		err = next(ctx, cmd)
		return err
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
