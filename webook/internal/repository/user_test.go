package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	cachemocks "webook/internal/repository/cache/mocks"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO)

		ctx context.Context
		uid int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存未命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{}, cache.ErrKeyNotExist)
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: "1999-02-02",
					AboutMe:  "自我介绍",
					Phone:    "123456789",
					Ctime:    time.UnixMilli(101),
				}).Return(nil)

				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{
					Id: uid,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "123456",
					Birthday: "1999-02-02",
					AboutMe:  "自我介绍",
					Phone: sql.NullString{
						String: "123456789",
						Valid:  true,
					},
					Ctime: 101,
				}, nil)

				return uc, ud
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: "1999-02-02",
				AboutMe:  "自我介绍",
				Phone:    "123456789",
				Ctime:    time.UnixMilli(101),
			},
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{
						Id:       123,
						Email:    "123@qq.com",
						Password: "123456",
						Birthday: "1999-02-02",
						AboutMe:  "自我介绍",
						Phone:    "123456789",
						Ctime:    time.UnixMilli(101),
					}, nil)
				ud := daomocks.NewMockUserDAO(ctrl)
				return uc, ud
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: "1999-02-02",
				AboutMe:  "自我介绍",
				Phone:    "123456789",
				Ctime:    time.UnixMilli(101),
			},
		},
		{
			name: "未找到用户",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{}, cache.ErrKeyNotExist)
				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{}, dao.ErrDataNotFound)

				return uc, ud
			},
			uid:      123,
			ctx:      context.Background(),
			wantUser: domain.User{},
			wantErr:  dao.ErrDataNotFound,
		},
		{
			name: "回写缓存失败",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{}, cache.ErrKeyNotExist)
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: "1999-02-02",
					AboutMe:  "自我介绍",
					Phone:    "123456789",
					Ctime:    time.UnixMilli(101),
				}).Return(errors.New("redis错误"))

				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().FindById(gomock.Any(), uid).Return(dao.User{
					Id: uid,
					Email: sql.NullString{
						String: "123@qq.com",
						Valid:  true,
					},
					Password: "123456",
					Birthday: "1999-02-02",
					AboutMe:  "自我介绍",
					Phone: sql.NullString{
						String: "123456789",
						Valid:  true,
					},
					Ctime: 101,
				}, nil)
				return uc, ud
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: "1999-02-02",
				AboutMe:  "自我介绍",
				Phone:    "123456789",
				Ctime:    time.UnixMilli(101),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc, ud := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)

			user, err := repo.FindById(tc.ctx, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
