package service

import (
	"context"
	"errors"
	"webook/internal/domain"
	"webook/internal/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateUser         = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户名或密码不正确")
)

type UserService interface {
	SignUp(ctx context.Context, user domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Edit(ctx context.Context, userId int64, nickname string, birthday string, aboutMe string) error
	GetProfile(ctx context.Context, userId int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type userService struct {
	ur repository.UserRepository
	// logger *zap.Logger
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		ur: repo,
		// logger: zap.L(),
	}
}

func (us *userService) SignUp(ctx context.Context, user domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return us.ur.Create(ctx, user)
}

func (us *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := us.ur.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (us *userService) Edit(ctx context.Context, userId int64, nickname string, birthday string, aboutMe string) error {
	return us.ur.UpdateUserById(ctx, userId, nickname, birthday, aboutMe)
}

func (us *userService) GetProfile(ctx context.Context, userId int64) (domain.User, error) {
	u, err := us.ur.FindById(ctx, userId)
	return u, err
}

func (us *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
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

func (us *userService) FindOrCreateByWechat(ctx context.Context,
	wechatInfo domain.WechatInfo) (domain.User, error) {
	u, err := us.ur.FindByWechat(ctx, wechatInfo.OpenId)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	// 这边意味着是一个新用户，所以记录一下
	// JSON 格式的 wechatInfo
	zap.L().Info("新用户", zap.Any("wechatInfo", wechatInfo))
	// us.logger.Info("新用户", zap.Any("wechatInfo", wechatInfo))
	err = us.ur.Create(ctx, domain.User{
		WechatInfo: wechatInfo,
	})
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	return us.ur.FindByWechat(ctx, wechatInfo.OpenId)
}
