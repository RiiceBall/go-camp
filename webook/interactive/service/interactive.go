package service

import (
	"context"
	"webook/interactive/domain"
	"webook/interactive/repository"

	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=./mocks/interactive.mock.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// Like 点赞
	Like(ctx context.Context, biz string, bizId int64, uid int64) error
	// CancelLike 取消点赞
	CancelLike(ctx context.Context, biz string, bizId int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	ir repository.InteractiveRepository
}

func NewInteractiveService(ir repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		ir: ir,
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
	return intr, eg.Wait()
}

func (is *interactiveService) GetByIds(ctx context.Context,
	biz string, ids []int64) (map[int64]domain.Interactive, error) {
	intrs, err := is.ir.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}
