package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	TopLike(ctx context.Context, biz string) ([]domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	ic cache.InteractiveCache
	id dao.InteractiveDAO
	l  logger.LoggerV1
}

func NewCachedInteractiveRepository(id dao.InteractiveDAO,
	ic cache.InteractiveCache, l logger.LoggerV1) InteractiveRepository {
	return &CachedInteractiveRepository{
		id: id,
		ic: ic,
		l:  l,
	}
}

func (ir *CachedInteractiveRepository) IncrReadCnt(ctx context.Context,
	biz string, bizId int64) error {
	err := ir.id.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	// 这边会有部分失败引起的不一致的问题，但是你其实不需要解决，
	// 因为阅读数不准确完全没有问题
	return ir.ic.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (ir *CachedInteractiveRepository) IncrLike(ctx context.Context,
	biz string, bizId int64, uid int64) error {
	err := ir.id.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return ir.ic.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (ir *CachedInteractiveRepository) DecrLike(ctx context.Context,
	biz string, bizId int64, uid int64) error {
	err := ir.id.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return ir.ic.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (ir *CachedInteractiveRepository) AddCollectionItem(ctx context.Context,
	biz string, bizId, cid, uid int64) error {
	err := ir.id.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Biz:   biz,
		Cid:   cid,
		BizId: bizId,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	return ir.ic.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (ir *CachedInteractiveRepository) Get(ctx context.Context,
	biz string, bizId int64) (domain.Interactive, error) {
	intr, err := ir.ic.Get(ctx, biz, bizId)
	if err == nil {
		// 缓存只缓存了具体的数字，但是没有缓存自身有没有点赞的信息
		// 因为一个人反复刷，重复刷一篇文章是小概率的事情
		// 也就是说，你缓存了某个用户是否点赞的数据，命中率会很低
		return intr, nil
	}
	ie, err := ir.id.Get(ctx, biz, bizId)
	if err == nil {
		res := ir.toDomain(ie)
		if er := ir.ic.Set(ctx, biz, bizId, res); er != nil {
			ir.l.Error("回写缓存失败",
				logger.Int64("bizId", bizId),
				logger.String("biz", biz),
				logger.Error(er))
		}
		return res, nil
	}
	if err == dao.ErrDataNotFound {
		return domain.Interactive{}, nil
	}
	return domain.Interactive{}, err
}

func (ir *CachedInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := ir.id.GetLikeInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrDataNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (ir *CachedInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := ir.id.GetCollectionInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrDataNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (ir *CachedInteractiveRepository) TopLike(ctx context.Context, biz string) ([]domain.Interactive, error) {
	interactives, err := ir.ic.GetTopLike(ctx, biz)
	if err == nil {
		return interactives, nil
	}
	ie, err := ir.id.TopLike(ctx, biz)
	if err == nil {
		res := ir.manyToDomain(ie)
		if er := ir.ic.SetTopLike(ctx, biz, res); er != nil {
			ir.l.Error("回写缓存失败",
				logger.String("biz", biz),
				logger.Error(er))
		}
		return res, nil
	}
	if err == dao.ErrDataNotFound {
		return []domain.Interactive{}, nil
	}
	return []domain.Interactive{}, err
}

func (ir *CachedInteractiveRepository) manyToDomain(intrs []dao.Interactive) []domain.Interactive {
	result := make([]domain.Interactive, len(intrs))
	for i, v := range intrs {
		result[i] = ir.toDomain(v)
	}
	return result
}

func (ir *CachedInteractiveRepository) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		ReadCnt:    intr.ReadCnt,
	}
}
