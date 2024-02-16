package ioc

import (
	"webook/internal/service/sms"
	"webook/internal/service/sms/localsms"
)

func InitSMSService() sms.Service {
	// 创建该方法是为了方便替换 sms 服务
	return localsms.NewService()
}
