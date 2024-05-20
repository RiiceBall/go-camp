package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/internal/domain"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExist = redis.Nil

//go:generate mockgen -source=./user.go -package=cachemocks -destination=./mocks/user.mock.go UserCache
type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, du domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (uc *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := uc.key(uid)
	data, err := uc.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	// 用 JSON 反序列化
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (uc *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := uc.key(du.Id)
	// 用 JSON 序列号
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	return uc.cmd.Set(ctx, key, data, uc.expiration).Err()
}

func (c *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("info:uid:%d", uid)
}

func NewUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}
