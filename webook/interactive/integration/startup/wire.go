//go:build wireinject

package startup

import (
	"webook/interactive/grpc"
	"webook/interactive/repository"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/interactive/service"

	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger,
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
