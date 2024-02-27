package failover

import (
	"context"
	"errors"
	"log"
	"webook/internal/service/sms"
)

type FailOverSMSService struct {
	svcs []sms.Service
}

func NewFailOverSMSService(svcs []sms.Service) *FailOverSMSService {
	return &FailOverSMSService{svcs: svcs}
}

func (f *FailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}
