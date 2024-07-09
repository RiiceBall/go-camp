//go:build wireinject

package main

import (
	grpc2 "webook/follow/grpc"
	"webook/follow/ioc"
	"webook/follow/repository"
	"webook/follow/repository/dao"
	"webook/follow/service"

	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	dao.NewGORMFollowRelationDAO,
	repository.NewFollowRelationRepository,
	service.NewFollowRelationService,
	grpc2.NewFollowRelationServiceServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
