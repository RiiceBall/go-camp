package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	rc cache.RankingCache

	// 下面是给 v1 用的
	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{rc: cache}
}

func NewCachedRankingRepositoryV1(redisCache *cache.RankingRedisCache, localCache *cache.RankingLocalCache) *CachedRankingRepository {
	return &CachedRankingRepository{redisCache: redisCache, localCache: localCache}
}

func (rr *CachedRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	res, err := rr.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}
	res, err = rr.redisCache.Get(ctx)
	if err != nil {
		return rr.localCache.ForceGet(ctx)
	}
	_ = rr.localCache.Set(ctx, res)
	return res, nil
}

func (rr *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return rr.rc.Get(ctx)
}

func (rr *CachedRankingRepository) ReplaceTopNV1(ctx context.Context, arts []domain.Article) error {
	_ = rr.localCache.Set(ctx, arts)
	return rr.redisCache.Set(ctx, arts)
}

func (rr *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return rr.rc.Set(ctx, arts)
}
