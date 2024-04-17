package cache

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"webook/internal/domain"

	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/interative_incr_cnt.lua
	luaIncrCnt string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldLikeCnt    = "like_cnt"
	fieldCollectCnt = "collect_cnt"
)

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
	GetTopLike(ctx context.Context, biz string) ([]domain.Interactive, error)
	SetTopLike(ctx context.Context, biz string, res []domain.Interactive) error
}

type RedisInteractiveCache struct {
	client redis.Cmdable
}

func NewRedisInteractiveCache(client redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{
		client: client,
	}
}

func (ic *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return ic.client.Eval(ctx, luaIncrCnt,
		[]string{ic.key(biz, bizId)},
		fieldReadCnt, 1).Err()
}

func (ic *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return ic.client.Eval(ctx, luaIncrCnt,
		[]string{ic.key(biz, bizId)},
		fieldLikeCnt, 1).Err()
}

func (ic *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return ic.client.Eval(ctx, luaIncrCnt,
		[]string{ic.key(biz, bizId)},
		fieldLikeCnt, -1).Err()
}

func (ic *RedisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context,
	biz string, bizId int64) error {
	return ic.client.Eval(ctx, luaIncrCnt,
		[]string{ic.key(biz, bizId)},
		fieldCollectCnt, 1).Err()
}

func (ic *RedisInteractiveCache) Get(ctx context.Context,
	biz string, bizId int64) (domain.Interactive, error) {
	// 直接使用 HMGet，即便缓存中没有对应的 key，也不会返回 error
	data, err := ic.client.HGetAll(ctx, ic.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		// 缓存不存在
		return domain.Interactive{}, ErrKeyNotExist
	}

	// 理论上来说，这里不可能有 error
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)

	return domain.Interactive{
		// 懒惰的写法
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
		ReadCnt:    readCnt,
	}, err
}

func (ic *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	key := ic.key(biz, bizId)
	err := ic.client.HMSet(ctx, key,
		fieldLikeCnt, intr.LikeCnt,
		fieldCollectCnt, intr.CollectCnt,
		fieldReadCnt, intr.ReadCnt).Err()
	if err != nil {
		return err
	}
	return ic.client.Expire(ctx, key, time.Minute*15).Err()
}

func (ic *RedisInteractiveCache) GetTopLike(ctx context.Context, biz string) ([]domain.Interactive, error) {
	key := ic.topLikeKey(biz)
	result, err := ic.client.Get(ctx, key).Result()
	if err != nil {
		return []domain.Interactive{}, err
	}
	var data []domain.Interactive
	// 反序列化数据
	err = json.Unmarshal([]byte(result), &data)
	if err != nil {
		return []domain.Interactive{}, err
	}
	return data, nil
}

func (ic *RedisInteractiveCache) SetTopLike(ctx context.Context, biz string, res []domain.Interactive) error {
	key := ic.topLikeKey(biz)
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	// 仅保存 5 分钟，以保持数据的实时性
	err = ic.client.Set(ctx, key, data, time.Minute*5).Err()
	if err != nil {
		return err
	}
	return nil
}

func (ic *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}

func (ic *RedisInteractiveCache) topLikeKey(biz string) string {
	return fmt.Sprintf("topLike:%s", biz)
}
