package repository

import (
	"context"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCachedArticleRepository_Sync(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) dao.ArticleDAO
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "新建同步成功",
			mock: func(ctrl *gomock.Controller) dao.ArticleDAO {
				articleDao := daomocks.NewMockArticleDAO(ctrl)
				articleDao.EXPECT().Sync(gomock.Any(), dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(int64(1), nil)
				return articleDao
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			articleDao := tc.mock(ctrl)
			repo := NewCachedArticleRepository(articleDao)
			id, err := repo.Sync(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
