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

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articlSvcProvider = wire.NewSet(
	repository.NewArticleRepository,
	cache.NewArticleRedisCache,
	dao.NewArticleGORMDAO,
	service.NewArticleService)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articlSvcProvider,
		interactiveSvcProvider,

		// Cache
		cache.NewCodeCache,

		// Repository
		repository.NewCodeRepository,

		// Service
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewCodeService,

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

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		interactiveSvcProvider,
		service.NewArticleService,
		cache.NewArticleRedisCache,
		web.NewArticleHandler,
		repository.NewArticleRepository)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdPartySet, interactiveSvcProvider)
	return service.NewInteractiveService(nil, nil)
}
