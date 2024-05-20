package repository

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"

	"github.com/ecodeclub/ekit/slice"
)

//go:generate mockgen -source=./article.go -package=repomocks -destination=./mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	ad dao.ArticleDAO
	ac cache.ArticleCache
	// 因为如果你直接访问 UserDAO，你就绕开了 repository，
	// repository 一般都有一些缓存机制
	ur UserRepository
}

func NewArticleRepository(ad dao.ArticleDAO,
	ur UserRepository,
	ac cache.ArticleCache) ArticleRepository {
	return &CachedArticleRepository{
		ad: ad,
		ur: ur,
		ac: ac,
	}
}

func (ar *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := ar.ad.Insert(ctx, ar.toEntity(art))
	if err == nil {
		er := ar.ac.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	return id, err
}

func (ar *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := ar.ad.UpdateById(ctx, ar.toEntity(art))
	if err == nil {
		er := ar.ac.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}

func (ar *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := ar.ad.Sync(ctx, ar.toEntity(art))
	if err == nil {
		er := ar.ac.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			// 也要记录日志
		}
	}
	// 在这里尝试，设置缓存
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 你可以灵活设置过期时间
		user, er := ar.ur.FindById(ctx, art.Author.Id)
		if er != nil {
			// 要记录日志
			return
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		er = ar.ac.SetPub(ctx, art)
		if er != nil {
			// 记录日志
		}
	}()
	return id, err
}

func (ar *CachedArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) error {
	err := ar.ad.SyncStatus(ctx, uid, id, status.ToUint8())
	if err == nil {
		er := ar.ac.DelFirstPage(ctx, uid)
		if er != nil {
			// 也要记录日志
		}
	}
	return err
}

func (ar *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 首先第一步，判定要不要查询缓存
	// 事实上， limit <= 100 都可以查询缓存
	if offset == 0 && limit == 100 {
		res, err := ar.ac.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			// 要考虑记录日志
			// 缓存未命中，你是可以忽略的
		}
	}
	arts, err := ar.ad.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return ar.toDomain(src)
	})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			// 缓存回写失败，不一定是大问题，但有可能是大问题
			err = ar.ac.SetFirstPage(ctx, uid, res)
			if err != nil {
				// 记录日志
				// 我需要监控这里
			}
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ar.preCache(ctx, res)
	}()
	return res, nil
}

func (ar *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := ar.ac.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := ar.ad.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = ar.toDomain(art)
	go func() {
		er := ar.ac.Set(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (ar *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := ar.ac.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := ar.ad.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 我现在要去查询对应的 User 信息，拿到创作者信息
	res = ar.toDomain(dao.Article(art))
	author, err := ar.ur.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
		// 要额外记录日志，因为你吞掉了错误信息
		//return res, nil
	}
	res.Author.Name = author.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := ar.ac.SetPub(ctx, res)
		if er != nil {
			// 记录日志
		}
	}()
	return res, nil
}

func (car *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	arts, err := car.ad.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.PublishedArticle, domain.Article](arts,
		func(idx int, src dao.PublishedArticle) domain.Article {
			return car.toDomain(dao.Article(src))
		}), nil
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

func (c *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}

func (ar *CachedArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := ar.ac.Set(ctx, arts[0])
		if err != nil {
			// 记录缓存
		}
	}
}
