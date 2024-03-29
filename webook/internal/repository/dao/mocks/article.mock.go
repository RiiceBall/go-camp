// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/repository/dao/article.go
//
// Generated by this command:
//
//	mockgen -source=./internal/repository/dao/article.go -package=daomocks -destination=./internal/repository/dao/mocks/article.mock.go
//

// Package daomocks is a generated GoMock package.
package daomocks

import (
	context "context"
	reflect "reflect"
	dao "webook/internal/repository/dao"

	gomock "go.uber.org/mock/gomock"
)

// MockArticleDAO is a mock of ArticleDAO interface.
type MockArticleDAO struct {
	ctrl     *gomock.Controller
	recorder *MockArticleDAOMockRecorder
}

// MockArticleDAOMockRecorder is the mock recorder for MockArticleDAO.
type MockArticleDAOMockRecorder struct {
	mock *MockArticleDAO
}

// NewMockArticleDAO creates a new mock instance.
func NewMockArticleDAO(ctrl *gomock.Controller) *MockArticleDAO {
	mock := &MockArticleDAO{ctrl: ctrl}
	mock.recorder = &MockArticleDAOMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockArticleDAO) EXPECT() *MockArticleDAOMockRecorder {
	return m.recorder
}

// GetByAuthor mocks base method.
func (m *MockArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset, limit int) ([]dao.Article, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByAuthor", ctx, uid, offset, limit)
	ret0, _ := ret[0].([]dao.Article)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByAuthor indicates an expected call of GetByAuthor.
func (mr *MockArticleDAOMockRecorder) GetByAuthor(ctx, uid, offset, limit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByAuthor", reflect.TypeOf((*MockArticleDAO)(nil).GetByAuthor), ctx, uid, offset, limit)
}

// GetById mocks base method.
func (m *MockArticleDAO) GetById(ctx context.Context, id int64) (dao.Article, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetById", ctx, id)
	ret0, _ := ret[0].(dao.Article)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetById indicates an expected call of GetById.
func (mr *MockArticleDAOMockRecorder) GetById(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetById", reflect.TypeOf((*MockArticleDAO)(nil).GetById), ctx, id)
}

// GetPubById mocks base method.
func (m *MockArticleDAO) GetPubById(ctx context.Context, id int64) (dao.PublishedArticle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPubById", ctx, id)
	ret0, _ := ret[0].(dao.PublishedArticle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPubById indicates an expected call of GetPubById.
func (mr *MockArticleDAOMockRecorder) GetPubById(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPubById", reflect.TypeOf((*MockArticleDAO)(nil).GetPubById), ctx, id)
}

// Insert mocks base method.
func (m *MockArticleDAO) Insert(ctx context.Context, art dao.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, art)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Insert indicates an expected call of Insert.
func (mr *MockArticleDAOMockRecorder) Insert(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockArticleDAO)(nil).Insert), ctx, art)
}

// Sync mocks base method.
func (m *MockArticleDAO) Sync(ctx context.Context, entity dao.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync", ctx, entity)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Sync indicates an expected call of Sync.
func (mr *MockArticleDAOMockRecorder) Sync(ctx, entity any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockArticleDAO)(nil).Sync), ctx, entity)
}

// SyncStatus mocks base method.
func (m *MockArticleDAO) SyncStatus(ctx context.Context, uid, id int64, status uint8) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncStatus", ctx, uid, id, status)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncStatus indicates an expected call of SyncStatus.
func (mr *MockArticleDAOMockRecorder) SyncStatus(ctx, uid, id, status any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncStatus", reflect.TypeOf((*MockArticleDAO)(nil).SyncStatus), ctx, uid, id, status)
}

// UpdateById mocks base method.
func (m *MockArticleDAO) UpdateById(ctx context.Context, art dao.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateById", ctx, art)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateById indicates an expected call of UpdateById.
func (mr *MockArticleDAOMockRecorder) UpdateById(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateById", reflect.TypeOf((*MockArticleDAO)(nil).UpdateById), ctx, art)
}