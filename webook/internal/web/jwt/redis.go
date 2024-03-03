package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisJWTHandler struct {
	signingMethod jwt.SigningMethod
	client        redis.Cmdable

	rcExpiration time.Duration
}

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		signingMethod: jwt.SigningMethodHS256,
		client:        client,
		rcExpiration:  time.Hour * 24 * 7,
	}
}

func (rh *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := rh.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New(("无效 Token"))
	}
	return nil
}

func (rh *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		// 没登陆，没有 Authorization
		return ""
	}
	segs := strings.SplitN(authCode, " ", 2)
	if len(segs) != 2 {
		// Authorization 是乱传的，格式不对
		return ""
	}
	return segs[1]
}

func (rh *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := rh.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = rh.SetJWTToken(ctx, uid, ssid)
	return err
}

func (rh *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(rh.signingMethod, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (rh *RedisJWTHandler) setRefreshToken(ctx *gin.Context,
	uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置为七天过期
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(rh.rcExpiration)),
		},
	}
	refreshToken := jwt.NewWithClaims(rh.signingMethod, rc)
	refreshTokenStr, err := refreshToken.SignedString(RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", refreshTokenStr)
	return nil
}

func (rh *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	return rh.client.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid),
		"", rh.rcExpiration).Err()
}

var JWTKey = []byte("WiXLWadWG44Rr2qP6VUDLod0dzAnRI45")
var RCJWTKey = []byte("WiXLWadWG44Rr2qP6VUDLod0dzAnRI34")

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}
