package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
}

type CachedArticleRepository struct {
	ad dao.ArticleDAO
}

func NewCachedArticleRepository(ad dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		ad: ad,
	}
}

func (car *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return car.ad.Insert(ctx, car.toEntity(art))
}

func (car *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return car.ad.UpdateById(ctx, car.toEntity(art))
}

func (car *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return car.ad.Sync(ctx, car.toEntity(art))
}

func (car *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	return car.ad.SyncStatus(ctx, uid, id, status.ToUint8())
}

func (car *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}
