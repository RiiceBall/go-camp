package job

import (
	"context"
	"time"
	"webook/internal/service"
)

type RankingJob struct {
	rs      service.RankingService
	timeout time.Duration
}

func NewRankingJob(svc service.RankingService, timeout time.Duration) *RankingJob {
	return &RankingJob{rs: svc, timeout: timeout}
}

func (rj *RankingJob) Name() string {
	return "ranking"
}

func (rj *RankingJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), rj.timeout)
	defer cancel()

	return rj.rs.TopN(ctx)
}
