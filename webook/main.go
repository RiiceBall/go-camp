package main

import (
	"net/http"
	"strings"
	"time"
	"webook/config"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/sms"
	"webook/internal/service/sms/localsms"
	"webook/internal/web"
	"webook/internal/web/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	server := initWebServer()
	codeService := initCodeService(redisClient)
	initUser(server, db, redisClient, codeService)

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})

	server.Run(":8080")
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token"},
		AllowOriginFunc: func(origin string) bool {
			return strings.Contains(origin, "localhost")
		},
		MaxAge: 12 * time.Hour,
	}))

	// redisClient := redis.NewClient(&redis.Options{
	// 	Addr: config.Config.Redis.Addr,
	// })
	// server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	useJWT(server)
	// useSession(server)

	return server
}

func initUser(server *gin.Engine, db *gorm.DB, cmd redis.Cmdable, cs *service.CodeService) {
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(cmd)
	ur := repository.NewUserRepository(ud, uc)
	us := service.NewUserService(ur)
	uh := web.NewUserHandler(us, cs)
	uh.RegisterRoutes(server)
}

func initCodeService(cmd redis.Cmdable) *service.CodeService {
	cc := cache.NewCodeCache(cmd)
	cr := repository.NewCodeRepository(cc)
	cs := service.NewCodeService(cr, initMemorySms())
	return cs
}

func initMemorySms() sms.Service {
	return localsms.NewService()
}

func useJWT(server *gin.Engine) {
	login := &middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

func useSession(server *gin.Engine) {
	store := cookie.NewStore([]byte("secret"))
	// store := memstore.NewStore([]byte("WiXLWadWG22RryqP6VUDLod0dzAnRI45"),
	// 	[]byte("zEefeFwntbeUBWanTU3GHopqjUnUQ0l4"))
	// store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	// 	[]byte("WiXLWadWG22RryqP6VUDLod0dzAnRI45"),
	// 	[]byte("zEefeFwntbeUBWanTU3GHopqjUnUQ0l4"))
	// if err != nil {
	// 	panic(err)
	// }
	server.Use(sessions.Sessions("ssid", store))

	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}
