package repository

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, time time.Time) error
}

type PreemptJobRepository struct {
	jd dao.JobDAO
}

func NewPreemptJobRepository(dao dao.JobDAO) CronJobRepository {
	return &PreemptJobRepository{jd: dao}
}

func (jr *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := jr.jd.Preempt(ctx)
	return domain.Job{
		Id:         j.Id,
		Expression: j.Expression,
		Executor:   j.Executor,
		Name:       j.Name,
	}, err
}

func (jr *PreemptJobRepository) Release(ctx context.Context, jid int64) error {
	return jr.jd.Release(ctx, jid)
}

func (jr *PreemptJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return jr.jd.UpdateUtime(ctx, id)
}

func (jr *PreemptJobRepository) UpdateNextTime(ctx context.Context, id int64, time time.Time) error {
	return jr.jd.UpdateNextTime(ctx, id, time)
}
