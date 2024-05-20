package repository

import (
	"context"
	"webook/internal/repository/cache"
)

var ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany
var ErrCodeSendTooMany = cache.ErrCodeSendTooMany

//go:generate mockgen -source=./code.go -package=repomocks -destination=./mocks/code.mock.go CodeRepository
type CodeRepository interface {
	Set(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, code string) (bool, error)
}

type CacheCodeRepository struct {
	cc cache.CodeCache
}

func NewCodeRepository(cc cache.CodeCache) CodeRepository {
	return &CacheCodeRepository{
		cc: cc,
	}
}

func (cr *CacheCodeRepository) Set(ctx context.Context, biz string,
	phone string, code string) error {
	return cr.cc.Set(ctx, biz, phone, code)
}

func (cr *CacheCodeRepository) Verify(ctx context.Context, biz string,
	phone string, code string) (bool, error) {
	return cr.cc.Verify(ctx, biz, phone, code)
}
