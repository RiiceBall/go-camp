package middleware

import (
	"net/http"
	"webook/internal/web"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
}

func (lmb *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if (path == "/users/signup") || (path == "/users/login") {
			return
		}
		sess := sessions.Default(ctx)
		if sess.Get(web.UserIdKey) == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
