package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	//Release(ctx context.Context, job domain.Job) error
	// 暴露 job 的增删改查方法
}

type cronJobService struct {
	cjr             repository.CronJobRepository
	l               logger.LoggerV1
	refreshInterval time.Duration
}

func NewCronJobService(repo repository.CronJobRepository, l logger.LoggerV1) CronJobService {
	return &cronJobService{
		cjr:             repo,
		l:               l,
		refreshInterval: time.Minute,
	}
}

func (cjs *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := cjs.cjr.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	ticker := time.NewTicker(cjs.refreshInterval)
	go func() {
		for range ticker.C {
			cjs.refresh(j.Id)
		}
	}()
	j.CancelFunc = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := cjs.cjr.Release(ctx, j.Id)
		if err != nil {
			cjs.l.Error("释放 job 失败",
				logger.Error(err),
				logger.Int64("jib", j.Id))
		}
	}
	return j, err
}
func (cjs *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return cjs.cjr.UpdateNextTime(ctx, j.Id, nextTime)
}

func (cjs *cronJobService) refresh(id int64) {
	// 本质上就是更新一下更新时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := cjs.cjr.UpdateUtime(ctx, id)
	if err != nil {
		cjs.l.Error("续约失败", logger.Error(err),
			logger.Int64("jid", id))
	}
}
