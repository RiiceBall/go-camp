//go:build wireinject

package main

import (
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,

		// DAO
		dao.NewUserDAO,
		dao.NewArticleGORMDAO,

		// Cache
		cache.NewCodeCache, cache.NewUserCache,

		// Repository
		repository.NewUserRepository, repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		// Service
		ioc.InitSMSService,
		InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// Handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ijwt.NewRedisJWTHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
