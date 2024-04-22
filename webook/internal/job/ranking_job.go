package job

import (
	"context"
	"sync"
	"time"
	"webook/internal/service"
	"webook/pkg/logger"

	rlock "github.com/gotomicro/redis-lock"
)

type RankingJob struct {
	rs service.RankingService
	l  logger.LoggerV1

	client  *rlock.Client
	key     string
	timeout time.Duration

	localLock *sync.Mutex
	lock      *rlock.Lock
}

func NewRankingJob(
	svc service.RankingService,
	l logger.LoggerV1,
	client *rlock.Client,
	timeout time.Duration) *RankingJob {
	return &RankingJob{
		rs:        svc,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
		timeout:   timeout,
	}
}

func (rj *RankingJob) Name() string {
	return "ranking"
}

func (rj *RankingJob) Run() error {
	rj.localLock.Lock()
	lock := rj.lock
	if lock == nil {
		// 抢分布式锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		lock, err := rj.client.Lock(ctx, rj.key, rj.timeout,
			&rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
				// 重试的超时
			}, time.Second)
		if err != nil {
			rj.localLock.Unlock()
			rj.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		rj.lock = lock
		rj.localLock.Unlock()
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(rj.timeout/2, rj.timeout)
			if er != nil {
				// 续约失败了
				// 你也没办法中断当下正在调度的热榜计算（如果有）
				rj.localLock.Lock()
				rj.lock = nil
				//lock.Unlock()
				rj.localLock.Unlock()
			}
		}()
	}
	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), rj.timeout)
	defer cancel()

	return rj.rs.TopN(ctx)
}

func (rj *RankingJob) Close() error {
	rj.localLock.Lock()
	lock := rj.lock
	rj.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
