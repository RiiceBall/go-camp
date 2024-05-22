package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate mockgen -source=./article.go -package=daomocks -destination=./mocks/article.mock.go ArticleDAO
type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, entity Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error)
}

type ArticleGORMDAO struct {
	db *gorm.DB
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{
		db: db,
	}
}

func (ad *ArticleGORMDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := ad.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (ad *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := ad.db.WithContext(ctx).Model(&Article{}).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("更新失败，ID不对或者作者不对")
	}
	return nil
}

func (ad *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
	id := art.Id
	err := ad.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			err error
		)
		dao := NewArticleGORMDAO(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		now := time.Now().UnixMilli()
		pubArt := PublishedArticle(art)
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			// 对MySQL不起效，但是可以兼容别的方言
			// INSERT xxx ON DUPLICATE KEY SET `title`=?
			// 别的方言：
			// sqlite INSERT XXX ON CONFLICT DO UPDATES WHERE
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"status":  pubArt.Status,
				"utime":   now,
			}),
		}).Create(&pubArt).Error
		return err
	})
	return id, err
}

func (ad *ArticleGORMDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return ad.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? and author_id = ?", uid, id).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID 不对或者创作者不对")
		}
		return tx.Model(&PublishedArticle{}).
			Where("id = ?", uid).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			}).Error
	})
}

func (ad *ArticleGORMDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := ad.db.WithContext(ctx).
		Where("author_id = ?", uid).
		Offset(offset).Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (ad *ArticleGORMDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := ad.db.WithContext(ctx).
		Where("id = ?", id).First(&art).Error
	return art, err
}

func (ad *ArticleGORMDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle
	err := ad.db.WithContext(ctx).
		Where("id = ?", id).
		First(&res).Error
	return res, err
}

func (ad *ArticleGORMDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var res []PublishedArticle
	const ArticleStatusPublished = 2
	err := ad.db.WithContext(ctx).
		Where("utime < ? AND status = ?",
			start.UnixMilli(), ArticleStatusPublished).
		Offset(offset).Limit(limit).
		Find(&res).Error
	return res, err
}

type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title   string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 我要根据创作者ID来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}

type PublishedArticle Article
