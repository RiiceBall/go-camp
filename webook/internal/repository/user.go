package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrRecordNotFound
)

type UserRepository struct {
	ud *dao.UserDAO
}

func NewUserRepository(ud *dao.UserDAO) *UserRepository {
	return &UserRepository{
		ud: ud,
	}
}

func (ur *UserRepository) Create(ctx context.Context, user domain.User) error {
	return ur.ud.Insert(ctx, dao.User{
		Email:    user.Email,
		Password: user.Password,
	})
}

func (ur *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := ur.ud.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return ur.toDomain(u), nil
}

func (ur *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	u, err := ur.ud.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return ur.toDomain(u), nil
}

func (ur *UserRepository) UpdateUserById(ctx context.Context, userId int64, nickname string, birthday string, aboutMe string) error {
	return ur.ud.UpdateById(ctx, userId, nickname, birthday, aboutMe)
}

func (ur *UserRepository) toDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
		Nickname: user.Nickname,
		Birthday: user.Birthday,
		AboutMe:  user.AboutMe,
	}
}
