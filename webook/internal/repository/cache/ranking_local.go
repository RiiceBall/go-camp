package cache

import (
	"context"
	"errors"
	"time"
	"webook/internal/domain"

	"github.com/ecodeclub/ekit/syncx/atomicx"
)

type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func (rc *RankingLocalCache) Set(ctx context.Context, arts []domain.Article) error {
	rc.topN.Store(arts)
	rc.ddl.Store(time.Now().Add(rc.expiration))
	return nil
}

func (rc *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := rc.ddl.Load()
	arts := rc.topN.Load()
	if len(arts) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}

func (rc *RankingLocalCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	arts := rc.topN.Load()
	if len(arts) == 0 {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}
