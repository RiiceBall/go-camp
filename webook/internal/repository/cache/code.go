package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("发送太频繁")
)

type CodeCache interface {
	Set(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz string, phone string, code string) (bool, error)
}

type RedisCodeCache struct {
	cmd redis.Cmdable
}

func NewCodeCache(cmd redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		cmd: cmd,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz string,
	phone string, code string) error {
	res, err := c.cmd.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case -1:
		return errors.New("验证码存在，但是没有过期时间")
	case -2:
		return ErrCodeSendTooMany
	default:
		return nil
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz string,
	phone string, code string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{biz, phone}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case -1:
		return false, ErrCodeVerifyTooMany
	case -2:
		return false, nil
	default:
		return true, nil
	}
}

func (c *RedisCodeCache) key(biz string, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
