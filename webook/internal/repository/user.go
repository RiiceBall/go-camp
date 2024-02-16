package repository

import (
	"context"
	"database/sql"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateUser
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	UpdateUserById(ctx context.Context, userId int64, nickname string, birthday string, aboutMe string) error
}

type CachedUserRepository struct {
	ud dao.UserDAO
	uc cache.UserCache
}

func NewUserRepository(ud dao.UserDAO, uc cache.UserCache) UserRepository {
	return &CachedUserRepository{
		ud: ud,
		uc: uc,
	}
}

func (ur *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return ur.ud.Insert(ctx, dao.User{
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
	})
}

func (ur *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := ur.ud.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return ur.toDomain(u), nil
}

func (ur *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	du, err := ur.uc.Get(ctx, id)
	if err == nil {
		return du, nil
	}
	u, err := ur.ud.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	du = ur.toDomain(u)
	_ = ur.uc.Set(ctx, du)
	return du, nil
}

func (ur *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := ur.ud.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return ur.toDomain(u), nil
}

func (ur *CachedUserRepository) UpdateUserById(ctx context.Context, userId int64, nickname string, birthday string, aboutMe string) error {
	return ur.ud.UpdateById(ctx, userId, nickname, birthday, aboutMe)
}

func (ur *CachedUserRepository) toDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Phone:    user.Phone.String,
		Password: user.Password,
		Nickname: user.Nickname,
		Birthday: user.Birthday,
		AboutMe:  user.AboutMe,
	}
}

func (ur *CachedUserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		AboutMe:  u.AboutMe,
	}
}
