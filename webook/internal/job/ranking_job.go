package job

import (
	"context"
	"math/rand"
	"sync"
	"time"
	"webook/internal/service"
	"webook/pkg/logger"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
)

type RankingJob struct {
	rs service.RankingService
	l  logger.LoggerV1

	client  *rlock.Client
	key     string
	timeout time.Duration

	localLock *sync.Mutex
	lock      *rlock.Lock

	// 作业提示
	// 随机生成一个，就代表当前负载。你可以每隔一分钟生成一个
	load        int32
	nodeId      string
	loadKey     string
	loadTicker  *time.Ticker
	redisClient *redis.Client
}

func NewRankingJob(
	svc service.RankingService,
	l logger.LoggerV1,
	client *rlock.Client,
	timeout time.Duration,
	nodeId string,
	redisClient *redis.Client) *RankingJob {
	rankingJob := &RankingJob{
		rs:        svc,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
		timeout:   timeout,
		// 0-100
		load:    rand.Int31n(101),
		nodeId:  nodeId,
		loadKey: "job:ranking:load:",
		// 随机负责刷新器
		// loadTicker:  time.NewTicker(1 * time.Minute),
		loadTicker:  time.NewTicker(5 * time.Second),
		redisClient: redisClient,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 创建时就将自己的负载加入到 redis 中
	rankingJob.redisClient.ZAdd(ctx, rankingJob.loadKey, redis.Z{
		Score:  float64(rankingJob.load),
		Member: rankingJob.nodeId,
	})
	// 生成负载
	go func() {
		rankingJob.generateLoad()
	}()
	return rankingJob
}

func (rj *RankingJob) Name() string {
	return "ranking"
}

func (rj *RankingJob) Run() error {
	if !rj.isMinLoadNode() {
		return nil
	}
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
			rj.l.Info("获取分布式锁失败", logger.String("nodeId", rj.nodeId), logger.Error(err))
			return nil
		}
		rj.lock = lock
		rj.l.Info("获取分布式锁成功", logger.String("nodeId", rj.nodeId))
		// 启动负载监控
		go rj.monitorLoad()
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
	rj.localLock.Unlock()
	// 这边就是你拿到了锁
	ctx, cancel := context.WithTimeout(context.Background(), rj.timeout)
	defer cancel()

	return rj.rs.TopN(ctx)
}

func (rj *RankingJob) Close() error {
	rj.localLock.Lock()
	lock := rj.lock
	rj.localLock.Unlock()
	// 关闭负载生成
	if rj.loadTicker != nil {
		rj.loadTicker.Stop()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 删除负载
	rj.redisClient.ZRem(ctx, rj.loadKey, rj.nodeId)
	return lock.Unlock(ctx)
}

func (rj *RankingJob) monitorLoad() {
	// checkTicker := time.NewTicker(time.Minute) // 每分钟检查一次
	checkTicker := time.NewTicker(4 * time.Second) // 每 4秒检查一次
	defer checkTicker.Stop()

	rj.localLock.Lock()
	lock := rj.lock
	rj.localLock.Unlock()

	if rj.lock != nil {
		for range checkTicker.C {
			// 检查当前节点是否仍是最小负载节点，如果不是，释放锁
			if rj.isMinLoadNode() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				rj.l.Info("当前节点不是最小负载节点，释放锁", logger.String("nodeId", rj.nodeId))
				rj.localLock.Lock()
				rj.lock = nil
				lock.Unlock(ctx)
				rj.localLock.Unlock()
				cancel()
				checkTicker.Stop()
			}
		}
	}
}

func (rj *RankingJob) isMinLoadNode() bool {
	nodes, err := rj.redisClient.ZRangeWithScores(context.Background(), rj.loadKey, 0, -1).Result()
	if err != nil {
		rj.l.Error("获取所有节点负载失败", logger.Error(err))
		return false
	}

	// 获取最小负载节点
	minLoadNode := nodes[0].Member.(string)
	return minLoadNode == rj.nodeId
}

func (rj *RankingJob) generateLoad() {
	// 每分钟生成一次负载，测试是每 5秒生成一次
	for range rj.loadTicker.C {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		rj.load = rand.Int31n(101)
		rj.redisClient.ZAdd(ctx, rj.loadKey, redis.Z{
			Score:  float64(rj.load),
			Member: rj.nodeId,
		})
		cancel()
	}
}
