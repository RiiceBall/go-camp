package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDuplicateUser  = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	UpdateById(ctx context.Context, id int64, nickname string, birthday string, aboutMe string) error
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (ud *GORMUserDAO) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := ud.db.WithContext(ctx).Create(&user).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const uniqueIndexErrNo uint16 = 1062
		if me.Number == uniqueIndexErrNo {
			return ErrDuplicateUser
		}
	}
	return err
}

func (ud *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (ud *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (ud *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

func (ud *GORMUserDAO) UpdateById(ctx context.Context, id int64, nickname string, birthday string, aboutMe string) error {
	return ud.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(User{
		Nickname: nickname,
		Birthday: birthday,
		AboutMe:  aboutMe,
	}).Error
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 创建唯一索引
	Email    sql.NullString `gorm:"unique"`
	Phone    sql.NullString `gorm:"unique"`
	Password string
	Nickname string `gorm:"type=varchar(16)"`
	Birthday string
	AboutMe  string `gorm:"type=varchar(1024)"`
	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
}
