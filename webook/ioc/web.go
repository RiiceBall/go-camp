package ioc

import (
	"context"
	"strings"
	"time"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/internal/web/middleware"
	"webook/pkg/ginx"
	"webook/pkg/ginx/middleware/prometheus"
	"webook/pkg/ginx/middleware/ratelimit"
	"webook/pkg/limiter"
	"webook/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler,
	artHdl *web.ArticleHandler,
	wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable,
	hdl ijwt.Handler, l logger.LoggerV1) []gin.HandlerFunc {
	pb := &prometheus.Builder{
		Namespace: "riiceball",
		Subsystem: "webook",
		Name:      "gin_http",
		Help:      "统计 GIN 的 HTTP 接口数据",
	}
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "riiceball",
		Subsystem: "webook",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
			AllowOriginFunc: func(origin string) bool {
				return strings.Contains(origin, "localhost")
			},
			MaxAge: 12 * time.Hour,
		}),
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),
		otelgin.Middleware("webook"),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 100)).Build(),
		middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("", logger.Field{Key: "req", Val: al})
		}).AllowReqBody().AllowRespBody().Build(),
		middleware.NewLoginJWTMiddlewareBuilder(hdl).CheckLogin(),
	}
}
