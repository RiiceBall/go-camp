package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/internal/domain"

	"github.com/redis/go-redis/v9"
)

type UserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func (uc *UserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
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

func (uc *UserCache) Set(ctx context.Context, du domain.User) error {
	key := uc.key(du.Id)
	// 用 JSON 序列号
	data, err := json.Marshal(du)
	if err != nil {
		return err
	}
	return uc.cmd.Set(ctx, key, data, uc.expiration).Err()
}

func (c *UserCache) key(uid int64) string {
	return fmt.Sprintf("info:uid:%d", uid)
}

func NewUserCache(cmd redis.Cmdable) *UserCache {
	return &UserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}
