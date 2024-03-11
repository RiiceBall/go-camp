package middleware

import (
	"net/http"
	"webook/internal/web"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) > 0 {
			for _, e := range ctx.Errors {
				// 断言自定义错误类型
				if appErr, ok := e.Err.(web.ErrorResult); ok {
					// 根据环境决定是否记录敏感信息
					if gin.Mode() == gin.DebugMode {
						zap.L().Error(appErr.ErrorMsg, zap.Any("context", appErr.Context), zap.Error(appErr.Err))
					} else {
						zap.L().Error(appErr.ErrorMsg, zap.Error(appErr.Err))
					}
					// 发送定制的JSON响应
					ctx.JSON(http.StatusOK, appErr.Result)
					ctx.Abort() // 防止其他中间件运行
					return
				}
			}
		}
	}
}
