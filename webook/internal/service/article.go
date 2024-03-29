package service

import (
	"context"
	"webook/internal/domain"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type articleService struct {
	ar       repository.ArticleRepository
	producer article.Producer
	l        logger.LoggerV1
}

func NewArticleService(ar repository.ArticleRepository,
	producer article.Producer,
	l logger.LoggerV1) ArticleService {
	return &articleService{
		ar:       ar,
		producer: producer,
		l:        l,
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

func (as *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return as.ar.GetByAuthor(ctx, uid, offset, limit)
}

func (as *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return as.ar.GetById(ctx, id)
}

func (as *articleService) GetPubById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	res, err := as.ar.GetPubById(ctx, id)
	go func() {
		if err == nil {
			er := as.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				as.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(er))
			}
		}
	}()
	return res, err
}
