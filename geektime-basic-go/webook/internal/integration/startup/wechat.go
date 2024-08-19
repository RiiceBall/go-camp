package startup

import (
	"gitee.com/geekbang/basic-go/webook/internal/service/oauth2/wechat"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}
