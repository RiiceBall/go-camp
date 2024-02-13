package repository

import (
	"context"
	"webook/internal/repository/cache"
)

var ErrCodeVerifyTooMany = cache.ErrCodeVerifyTooMany
var ErrCodeSendTooMany = cache.ErrCodeSendTooMany

type CodeRepository struct {
	cc *cache.CodeCache
}

func NewCodeRepository(cc *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		cc: cc,
	}
}

func (cr *CodeRepository) Set(ctx context.Context, biz string,
	phone string, code string) error {
	return cr.cc.Set(ctx, biz, phone, code)
}

func (cr *CodeRepository) Verify(ctx context.Context, biz string,
	phone string, code string) (bool, error) {
	return cr.cc.Verify(ctx, biz, phone, code)
}
