package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("发送太频繁")
	ErrCodeNotSended     = errors.New("未发送验证码")
)

type CodeCache struct {
	lock       sync.Mutex
	cache      map[string]*CacheItem
	expiration time.Duration
}

type CacheItem struct {
	Code       string
	Count      int
	Expiration time.Time
}

func NewCodeCache(cache map[string]*CacheItem) *CodeCache {
	return &CodeCache{
		cache:      cache,
		expiration: time.Minute * 10,
	}
}

func (c *CodeCache) Set(ctx context.Context, biz string,
	phone string, code string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := c.key(biz, phone)
	item, found := c.cache[key]
	// 如果找到并且剩余时间超过 9 分钟，则返回错误
	if found && item.Expiration.Sub(time.Now()) > time.Minute*9 {
		return ErrCodeSendTooMany
	}
	c.cache[key] = &CacheItem{
		Code:       code,
		Count:      3,
		Expiration: time.Now().Add(c.expiration),
	}
	return nil
}

func (c *CodeCache) Verify(ctx context.Context, biz string,
	phone string, code string) (bool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	key := c.key(biz, phone)
	item, found := c.cache[key]
	// 如果没找到
	if !found {
		return false, ErrCodeNotSended
	}
	// 如果次数耗尽或是密码错误
	if item.Count <= 0 || item.Code != code {
		item.Count--
		return false, ErrCodeVerifyTooMany
	}
	// 密码正确，将次数清零
	item.Count = 0
	return true, nil
}

func (c *CodeCache) key(biz string, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
