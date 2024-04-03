package prometheus

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type Builder struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceId string
}

func (b *Builder) BuildResponseTime() gin.HandlerFunc {
	// pattern 是指命中的路由
	labels := []string{"method", "pattern", "status"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		// Namespace，Subsystem 和 Name 不能使用 _ 以外的符号
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name + "_resp_time",
		Help:      b.Help,
		ConstLabels: map[string]string{
			"instance_id": b.InstanceId,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(vector)
	return func(ctx *gin.Context) {
		start := time.Now()
		defer func() {
			// 准备上报 prometheus
			duration := time.Since(start).Milliseconds()
			method := ctx.Request.Method
			pattern := ctx.FullPath()
			status := ctx.Writer.Status()
			vector.WithLabelValues(method, pattern, strconv.Itoa(status)).Observe(float64(duration))
		}()
		ctx.Next()
	}
}

func (b *Builder) BuildActiveRequest() gin.HandlerFunc {
	// 一般我们只关心总的活跃请求数
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name + "_active_req",
		Help:      b.Help,
		ConstLabels: map[string]string{
			"instance_id": b.InstanceId,
		},
	})
	prometheus.MustRegister(gauge)
	return func(ctx *gin.Context) {
		gauge.Inc()
		defer gauge.Dec()
		ctx.Next()
	}
}
