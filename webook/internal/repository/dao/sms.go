package dao

import (
	"context"

	"gorm.io/gorm"
)

type SmsDAO interface {
	Insert(ctx context.Context, sms Sms) error
	FindFirstSms(ctx context.Context) (Sms, error)
	UpdateRetryLeft(ctx context.Context, id int64, retryLeft int) error
	DeleteById(ctx context.Context, id int64) error
}

type GORMSmsDAO struct {
	db *gorm.DB
}

func NewSmsDAO(db *gorm.DB) SmsDAO {
	return &GORMSmsDAO{
		db: db,
	}
}

func (sd *GORMSmsDAO) Insert(ctx context.Context, sms Sms) error {
	err := sd.db.WithContext(ctx).Create(&sms).Error
	return err
}

func (sd *GORMSmsDAO) FindFirstSms(ctx context.Context) (Sms, error) {
	var s Sms
	err := sd.db.WithContext(ctx).First(&s).Error
	return s, err
}

func (sd *GORMSmsDAO) UpdateRetryLeft(ctx context.Context, id int64, retryLeft int) error {
	err := sd.db.WithContext(ctx).Model(&Sms{}).Where("id = ?", id).Update("retry_left", retryLeft).Error
	return err
}

func (sd *GORMSmsDAO) DeleteById(ctx context.Context, id int64) error {
	err := sd.db.WithContext(ctx).Where("id = ?", id).Delete(&Sms{}).Error
	return err
}

type Sms struct {
	Id        int64 `gorm:"primaryKey,autoIncrement"`
	TplId     string
	Args      []string
	Numbers   []string
	RetryLeft int // 剩余重试次数
}
