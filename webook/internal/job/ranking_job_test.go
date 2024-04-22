package job

import (
	"testing"
	"time"
	svcmocks "webook/internal/service/mocks"
	"webook/pkg/logger"

	rlock "github.com/gotomicro/redis-lock"

	"github.com/redis/go-redis/v9"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestRankingJob(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	svc := svcmocks.NewMockRankingService(ctrl)
	svc.EXPECT().TopN(gomock.Any()).AnyTimes().Return(nil)
	zl, _ := zap.NewDevelopment()
	l := logger.NewZapLogger(zl)
	job1 := NewRankingJob(
		svc,
		l,
		rlock.NewClient(redisClient),
		time.Second*15,
		"node1",
		redisClient,
	)
	job2 := NewRankingJob(
		svc,
		l,
		rlock.NewClient(redisClient),
		time.Second*15,
		"node2",
		redisClient,
	)

	ticker := time.NewTicker(time.Second * 2)
	// 每 2秒执行一次 Run
	for range ticker.C {
		job1.Run()
		job2.Run()
		l.Info("当前节点负载", logger.String("节点", job1.nodeId), logger.Int32("负载", job1.load))
		l.Info("当前节点负载", logger.String("节点", job2.nodeId), logger.Int32("负载", job2.load))
	}
}
