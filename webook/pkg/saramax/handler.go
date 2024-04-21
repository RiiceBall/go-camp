package saramax

import (
	"encoding/json"
	"time"
	"webook/pkg/logger"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
)

type Handler[T any] struct {
	l      logger.LoggerV1
	fn     func(msg *sarama.ConsumerMessage, evt T) error
	vector *prometheus.SummaryVec
}

func NewHandler[T any](l logger.LoggerV1, fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "riiceball",
		Subsystem: "webook",
		Name:      "sarama 消费",
		Help:      "统计 sarama 消费的耗时",
	}, []string{"topic"})
	return &Handler[T]{
		l:      l,
		fn:     fn,
		vector: vector,
	}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		// 计算消费耗时
		start := time.Now()
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.l.Error("反序列化失败",
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
				logger.Error(err))
			continue
		}
		err = h.fn(msg, t)
		if err != nil {
			h.l.Error("消费失败",
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
				logger.Error(err))
			continue
		}
		duration := time.Since(start).Milliseconds()
		h.vector.WithLabelValues(msg.Topic).
			Observe(float64(duration))
		session.MarkMessage(msg, "")
	}
	return nil
}
