package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type jwtHandler struct {
}

func (jh *jwtHandler) setJWTToken(ctx *gin.Context, uid int64) {
	uc := UserClaims{
		Uid:       uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
}

var JWTKey = []byte("WiXLWadWG44Rr2qP6VUDLod0dzAnRI45")

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
