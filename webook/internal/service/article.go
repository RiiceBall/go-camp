package service

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
}

type articleService struct {
	ar repository.ArticleRepository
}

func NewArticleService(ar repository.ArticleRepository) ArticleService {
	return &articleService{
		ar: ar,
	}
}

func (as *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := as.ar.Update(ctx, art)
		return art.Id, err
	}
	return as.ar.Create(ctx, art)
}

func (as *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return as.ar.Sync(ctx, art)
}

func (as *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return as.ar.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}
