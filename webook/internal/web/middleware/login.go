package middleware

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"
	"webook/internal/web"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
}

func (lmb *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if (path == "/users/signup") || (path == "/users/login") {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get(web.UserIdKey)
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()

		const updateTimeKey = "update_time"
		val := sess.Get(updateTimeKey)
		// 检查 val 是否为时间类型
		lastUpdateTime, ok := val.(time.Time)
		// 如果已经超过1分钟则刷新
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Minute {
			sess.Set(web.UserIdKey, userId)
			sess.Set(updateTimeKey, now)
			sess.Options(sessions.Options{
				// 900 seconds
				MaxAge: 900,
			})
			err := sess.Save()
			if err != nil {
				// 打印日志
				fmt.Println(err)
			}
		}
	}
}
