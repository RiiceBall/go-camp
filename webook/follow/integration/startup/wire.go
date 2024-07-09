//go:build wireinject

package startup

import (
	"webook/follow/grpc"
	"webook/follow/repository"
	"webook/follow/repository/cache"
	"webook/follow/repository/dao"
	"webook/follow/service"

	"github.com/google/wire"
)

func InitServer() *grpc.FollowServiceServer {
	wire.Build(
		InitRedis,
		InitLog,
		InitTestDB,
		dao.NewGORMFollowRelationDAO,
		cache.NewRedisFollowCache,
		repository.NewFollowRelationRepository,
		service.NewFollowRelationService,
		grpc.NewFollowRelationServiceServer,
	)
	return new(grpc.FollowServiceServer)
}
