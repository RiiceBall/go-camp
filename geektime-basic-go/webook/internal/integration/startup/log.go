package startup

import "gitee.com/geekbang/basic-go/webook/pkg/logger"

func InitLogger() logger.LoggerV1 {
	return logger.NewNopLogger()
}
