package limiter

import "context"

//go:generate mockgen -source=./types.go -package=limitermocks -destination=./mocks/limiter.mock.go Limiter
type Limiter interface {
	// Limit 是否出发限流
	// 返回 true 表示触发限流
	Limit(ctx context.Context, key string) (bool, error)
}
