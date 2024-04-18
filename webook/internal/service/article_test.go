package service

import (
	"context"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
	"webook/pkg/logger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestArticleService_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.ArticleRepository

		art domain.Article

		wantId  int64
		wantErr error
	}{
		{
			name: "发表成功",
			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				repo := repomocks.NewMockArticleRepository(ctrl)
				repo.EXPECT().Sync(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return repo
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: int64(1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewArticleService(tc.mock(ctrl), nil, &logger.NopLogger{})
			id, err := svc.Publish(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
