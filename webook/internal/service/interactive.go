package service

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"

	"golang.org/x/sync/errgroup"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// Like 点赞
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	// CancelLike 取消点赞
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
}

type interactiveService struct {
	ir repository.InteractiveRepository
	l  logger.LoggerV1
}

func NewInteractiveService(ir repository.InteractiveRepository,
	l logger.LoggerV1) InteractiveService {
	return &interactiveService{
		ir: ir,
		l:  l,
	}
}

func (is *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return is.ir.IncrReadCnt(ctx, biz, bizId)
}

func (is *interactiveService) Like(ctx context.Context, biz string, bizId int64, uid int64) error {
	return is.ir.IncrLike(ctx, biz, bizId, uid)
}

func (is *interactiveService) CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	return is.ir.DecrLike(ctx, biz, bizId, uid)
}

func (is *interactiveService) Collect(ctx context.Context,
	biz string, bizId, cid, uid int64) error {
	return is.ir.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (is *interactiveService) Get(
	ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	// 你也可以考虑将分发的逻辑也下沉到 repository 里面
	intr, err := is.ir.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		intr.Liked, err = is.ir.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		intr.Collected, err = is.ir.Collected(ctx, biz, bizId, uid)
		return err
	})
	// 说明是登录过的，补充用户是否点赞或者
	err = eg.Wait()
	if err != nil {
		// 这个查询失败只需要记录日志就可以，不需要中断执行
		is.l.Error("查询用户是否点赞的信息失败",
			logger.String("biz", biz),
			logger.Int64("bizId", bizId),
			logger.Int64("uid", uid),
			logger.Error(err))
	}
	return intr, err
}
