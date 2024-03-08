package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

var ErrSmsNotFound = dao.ErrRecordNotFound

type SmsRepository interface {
	Create(ctx context.Context, sms domain.Sms) error
	FindFirstSms(ctx context.Context) (domain.Sms, error)
	UpdateRetryLeft(ctx context.Context, id int64, retryLeft int) error
	DeleteById(ctx context.Context, id int64) error
}

type SmsRepositoryStruct struct {
	sd dao.SmsDAO
}

func NewSmsRepository(sd dao.SmsDAO) SmsRepository {
	return &SmsRepositoryStruct{
		sd: sd,
	}
}

func (s *SmsRepositoryStruct) Create(ctx context.Context, sms domain.Sms) error {
	return s.sd.Insert(ctx, s.toEntity(sms))
}

func (s *SmsRepositoryStruct) FindFirstSms(ctx context.Context) (domain.Sms, error) {
	sms, err := s.sd.FindFirstSms(ctx)
	if err != nil {
		return domain.Sms{}, err
	}
	return s.toDomain(sms), nil
}

func (s *SmsRepositoryStruct) UpdateRetryLeft(ctx context.Context, id int64, retryLeft int) error {
	return s.sd.UpdateRetryLeft(ctx, id, retryLeft)
}

func (s *SmsRepositoryStruct) DeleteById(ctx context.Context, id int64) error {
	return s.sd.DeleteById(ctx, id)
}

func (s *SmsRepositoryStruct) toDomain(sms dao.Sms) domain.Sms {
	return domain.Sms{
		Id:        sms.Id,
		TplId:     sms.TplId,
		Args:      sms.Args,
		Numbers:   sms.Numbers,
		RetryLeft: sms.RetryLeft,
	}
}

func (s *SmsRepositoryStruct) toEntity(sms domain.Sms) dao.Sms {
	return dao.Sms{
		Id:        sms.Id,
		TplId:     sms.TplId,
		Args:      sms.Args,
		Numbers:   sms.Numbers,
		RetryLeft: sms.RetryLeft,
	}
}
