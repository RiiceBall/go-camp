package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{db: db}
}

func (jd *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := jd.db.WithContext(ctx)
	for {
		var j Job
		now := time.Now().UnixMilli()
		oneMinuteAgo := time.Now().Add(-1 * (time.Minute + time.Second*30)).UnixMilli()
		// 如果运行中的任务超过 1分半没有更新过，就可以代表可以重新抢占
		// 这个时间需要根据 cronJobService 中的 refreshInterval 来调整
		err := db.Where("(status = ? AND next_time < ?) OR (status = ? AND utime < ?)",
			jobStatusWaiting, now, jobStatusRunning, oneMinuteAgo).
			First(&j).Error
		if err != nil {
			return j, err
		}
		res := db.WithContext(ctx).Model(&Job{}).
			Where("id = ? AND version = ?", j.Id, j.Version).
			Updates(map[string]any{
				"status":  jobStatusRunning,
				"version": j.Version + 1,
				"utime":   now,
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 没抢到
			continue
		}
		return j, err
	}
}

func (jd *GORMJobDAO) Release(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return jd.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}

func (jd *GORMJobDAO) UpdateUtime(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return jd.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).Updates(map[string]any{
		"utime": now,
	}).Error
}

func (jd *GORMJobDAO) UpdateNextTime(ctx context.Context, jid int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return jd.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", jid).Updates(map[string]any{
		"utime":     now,
		"next_time": t.UnixMilli(),
	}).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string
	Cfg        string
	// 状态来表达，是不是可以抢占，有没有被人抢占
	Status int

	Version int

	NextTime int64 `gorm:"index"`

	Utime int64
	Ctime int64
}

const (
	// jobStatusWaiting 没人抢
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢了
	jobStatusRunning
	// jobStatusPaused 不再需要调度了
	jobStatusPaused
)
