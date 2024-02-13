package service

import (
	"context"
	"errors"
	"webook/internal/domain"
	"webook/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateUser         = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户名或密码不正确")
)

type UserService struct {
	ur *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		ur: repo,
	}
}

func (us *UserService) SignUp(ctx context.Context, user domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return us.ur.Create(ctx, user)
}

func (us *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := us.ur.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (us *UserService) Edit(ctx context.Context, userId int64, nickname string, birthday string, aboutMe string) error {
	return us.ur.UpdateUserById(ctx, userId, nickname, birthday, aboutMe)
}

func (us *UserService) GetProfile(ctx context.Context, userId int64) (domain.User, error) {
	u, err := us.ur.FindById(ctx, userId)
	return u, err
}

func (us *UserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := us.ur.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	err = us.ur.Create(ctx, domain.User{
		Phone: phone,
	})
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	return us.ur.FindByPhone(ctx, phone)
}
