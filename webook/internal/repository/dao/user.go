package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound     = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (ud *UserDAO) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := ud.db.WithContext(ctx).Create(&user).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const uniqueIndexErrNo uint16 = 1062
		if me.Number == uniqueIndexErrNo {
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (ud *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (ud *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := ud.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (ud *UserDAO) UpdateById(ctx context.Context, id int64, nickname string, birthday string, aboutMe string) error {
	return ud.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(User{
		Nickname: nickname,
		Birthday: birthday,
		AboutMe:  aboutMe,
	}).Error
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 创建唯一索引
	Email    string `gorm:"unique"`
	Password string
	Nickname string `gorm:"type=varchar(16)"`
	Birthday string
	AboutMe  string `gorm:"type=varchar(1024)"`
	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
}
