package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"
	"webook/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
}

func (lmb *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if (path == "/users/signup") || (path == "/users/login") || (path == "/hello") {
			return
		}
		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			// 没登陆，没有 Authorization
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(authCode, " ")
		if len(segs) != 2 {
			// Authorization 是乱传的，格式不对
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		if err != nil {
			// token 有问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			// token 成功解析，但是非法或是过期了
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if uc.UserAgent != ctx.GetHeader("User-Agent") {
			// 能进入这里的大概率是攻击者，以后要埋点（记录日志）
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expireTime := uc.ExpiresAt
		now := time.Now()
		if expireTime.Sub(now) < time.Minute*10 {
			uc.ExpiresAt = jwt.NewNumericDate(now.Add(time.Minute * 30))
			println(tokenStr)
			newTokenStr, err := token.SignedString(web.JWTKey)
			println(newTokenStr)
			ctx.Header("x-jwt-token", newTokenStr)
			println("go")
			if err != nil {
				log.Println(err)
			}
		}
		ctx.Set("user", uc)
	}
}
