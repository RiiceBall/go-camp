package service

import (
	"context"
	"fmt"
	"math/rand"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

//go:generate mockgen -source=./code.go -package=svcmocks -destination=./mocks/code.mock.go CodeService
type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type codeService struct {
	cr repository.CodeRepository
	ss sms.Service
}

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

func NewCodeService(cr repository.CodeRepository, ss sms.Service) CodeService {
	return &codeService{
		cr: cr,
		ss: ss,
	}
}

func (cs *codeService) Send(ctx context.Context, biz string, phone string) error {
	code := cs.generate()
	err := cs.cr.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	const codeTplId = "1877556"
	return cs.ss.Send(ctx, codeTplId, []string{code}, phone)
}

func (cs *codeService) Verify(ctx context.Context,
	biz string, phone string, inputCode string) (bool, error) {
	ok, err := cs.cr.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooMany {
		// 将这个错误屏蔽，单纯的告诉调用者有问题就好了
		return false, nil
	}
	return ok, err
}

func (cs *codeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
