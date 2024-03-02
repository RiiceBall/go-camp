//go:build wireinject

package main

import (
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRedis,

		// DAO
		dao.NewUserDAO,

		// Cache
		cache.NewCodeCache, cache.NewUserCache,

		// Repository
		repository.NewUserRepository, repository.NewCodeRepository,

		// Service
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,

		// Handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
