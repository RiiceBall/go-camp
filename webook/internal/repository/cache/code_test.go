package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"webook/internal/repository/cache/redismocks"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRedisCodeCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}

	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) redis.Cmdable

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(0))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "123456789")}, []any{"123456"}).
					Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123456789",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "redis 返回 error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(errors.New("redis错误"))
				cmd.SetVal(int64(0))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "123456789")}, []any{"123456"}).
					Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123456789",
			code:    "123456",
			wantErr: errors.New("redis错误"),
		},
		{
			name: "验证码存在，但是没有过期时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(-1))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "123456789")}, []any{"123456"}).
					Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123456789",
			code:    "123456",
			wantErr: errors.New("验证码存在，但是没有过期时间"),
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := redismocks.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(-2))
				res.EXPECT().Eval(gomock.Any(), luaSetCode,
					[]string{keyFunc("test", "123456789")}, []any{"123456"}).
					Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123456789",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := NewCodeCache(tc.mock(ctrl))
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
