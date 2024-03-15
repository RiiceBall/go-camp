//go:build wireinject

package startup

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

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,

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
	)
	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		dao.NewArticleGORMDAO,
		service.NewArticleService,
		web.NewArticleHandler,
		repository.NewCachedArticleRepository)
	return &web.ArticleHandler{}
}
