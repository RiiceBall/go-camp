package tencent

import (
	"context"
	"fmt"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SetContext(ctx)
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.TemplateParamSet = toStringPtrSlice(args)
	req.PhoneNumberSet = toStringPtrSlice(numbers)
	resp, err := s.client.SendSms(req)
	// 这里的日志记录是为了方便调试
	zap.L().Debug("请求腾讯SendSMS接口", zap.Any("req", req), zap.Any("resp", resp))
	if err != nil {
		return err
	}
	for _, statusPtr := range resp.Response.SendStatusSet {
		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送失败，code：%s，原因：%s", *status.Code, *status.Message)
		}
	}
	return nil
}

func toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}

func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		client:   client,
		appId:    &appId,
		signName: &signName,
	}
}
