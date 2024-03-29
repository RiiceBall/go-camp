//go:build wireinject

package main

import (
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"

	"github.com/google/wire"
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,
		ioc.InitSaramaClient, ioc.InitSyncProducer,
		ioc.InitConsumers,

		// DAO
		dao.NewUserDAO,
		dao.NewArticleGORMDAO,

		interactiveSvcProvider,

		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventConsumer,

		// Cache
		cache.NewCodeCache, cache.NewUserCache,
		cache.NewArticleRedisCache,

		// Repository
		repository.NewUserRepository, repository.NewCodeRepository,
		repository.NewArticleRepository,

		// Service
		ioc.InitSMSService,
		ioc.InitWechatService,
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

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
