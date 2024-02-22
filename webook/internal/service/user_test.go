package service

import (
	"context"
	"errors"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordEncrypt(t *testing.T) {
	pwd := []byte("1234#5678abC")
	encrypted, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	assert.NoError(t, err)
	println(string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, pwd)
	assert.NoError(t, err)
}

func TestUserService_Login(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 预期的输入
		ctx      context.Context
		email    string
		password string

		// 预期的输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email: "123@qq.com",
						// 这里拿到的密码应该是加密后的密码
						Password: "$2a$10$ZC7jb8P3CeHGho.W7jHnmulsK.qkaNkOil3UzJo5Y7BoVKgg4ioaa",
						Phone:    "12345678901",
					}, nil)
				return repo
			},
			email: "123@qq.com",
			// 这里的密码是明文
			password: "1234#5678abC",
			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$ZC7jb8P3CeHGho.W7jHnmulsK.qkaNkOil3UzJo5Y7BoVKgg4ioaa",
				Phone:    "12345678901",
			},
			wantErr: nil,
		},
		{
			name: "用户未找到",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email: "123@qq.com",
			// 这里的密码是明文
			password: "1234#5678abC",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("db错误"))
				return repo
			},
			email: "123@qq.com",
			// 这里的密码是明文
			password: "1234#5678abC",
			wantUser: domain.User{},
			wantErr:  errors.New("db错误"),
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, nil)
				return repo
			},
			email: "123@qq.com",
			// 这里的密码是明文
			password: "1234#5678abCoo",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := NewUserService(repo)
			user, err := svc.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
