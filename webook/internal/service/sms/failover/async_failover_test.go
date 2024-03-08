package failover

import (
	"context"
	"errors"
	"testing"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
	"webook/internal/service/sms"
	smsmocks "webook/internal/service/sms/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAsyncFailOverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (sms.Service, repository.SmsRepository)

		errorThreshold int32
		wantErr        error
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SmsRepository) {
				svc := smsmocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return svc, nil
			},
			errorThreshold: 2,
			wantErr:        nil,
		},
		{
			name: "发送失败",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SmsRepository) {
				svc := smsmocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("发送失败"))
				return svc, nil
			},
			errorThreshold: 2,
			wantErr:        errors.New("发送失败"),
		},
		{
			name: "错误率过高",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SmsRepository) {
				svc := smsmocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("发送失败"))
				smsRepo := repomocks.NewMockSmsRepository(ctrl)
				smsRepo.EXPECT().Create(gomock.Any(), domain.Sms{
					TplId:     "tplId",
					Args:      []string{"args"},
					Numbers:   []string{"number"},
					RetryLeft: 3,
				}).Return(nil)
				return svc, smsRepo
			},
			// 错误阈值设置为 0，表示只要有错误就认为错误率过高
			errorThreshold: 0,
			wantErr:        nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, smsRepo := tc.mock(ctrl)
			s := NewAsyncFailOverSMSService(svc, smsRepo, tc.errorThreshold, time.Second)
			err := s.Send(context.Background(), "tplId", []string{"args"}, "number")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
