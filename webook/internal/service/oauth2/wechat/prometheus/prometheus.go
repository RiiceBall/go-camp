package prometheus

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/service/oauth2/wechat"

	"github.com/prometheus/client_golang/prometheus"
)

type Decorator struct {
	wechat.Service
	sum prometheus.Summary
}

func NewDecorator(s wechat.Service, sum prometheus.Summary) *Decorator {
	return &Decorator{
		Service: s,
		sum:     sum,
	}
}

func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.sum.Observe(float64(duration))
	}()
	return d.Service.VerifyCode(ctx, code)
}
