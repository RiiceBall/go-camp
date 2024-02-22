package web

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/internal/domain"
	"webook/internal/service"
	svcmocks "webook/internal/service/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)

		// 预期的输入
		reqBuilder func(t *testing.T) *http.Request

		// 预期的输出
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`{
						"email": "123@qq.com",
						"password": "hello#world123",
						"confirmPassword": "hello#world123"
					}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "Bind出错",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`{
						"email": "123@qq.com",
						"password": "hello#world123",
						confirmPassword: "hello#world123"
					}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`{
						"email": "123",
						"password": "hello#world123",
						"confirmPassword": "hello#world123"
					}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "非法邮箱格式",
		},
		{
			name: "两次密码输入不同",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`{
						"email": "123@qq.com",
						"password": "hello#world12322",
						"confirmPassword": "hello#world123"
					}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "两次输入密码不对",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`{
						"email": "123@qq.com",
						"password": "hello",
						"confirmPassword": "hello"
					}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "密码必须包含字母、数字、特殊字符，并且不少于八位",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(errors.New("db错误"))
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`{
						"email": "123@qq.com",
						"password": "hello#world123",
						"confirmPassword": "hello#world123"
					}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(service.ErrDuplicateUser)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "/users/signup",
					bytes.NewReader([]byte(`{
						"email": "123@qq.com",
						"password": "hello#world123",
						"confirmPassword": "hello#world123"
					}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "重复邮箱，请换一个邮箱",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 构造 handler
			userSvc, codeSvc := tc.mock(ctrl)
			hdl := NewUserHandler(userSvc, codeSvc)

			// 准备服务器，注册路由
			server := gin.Default()
			hdl.RegisterRoutes(server)

			// 准备 req 和记录 recorder
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()

			// 执行测试
			server.ServeHTTP(recorder, req)

			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}

// TestEmailPattern 用来验证我们的邮箱正则表达式对不对
func TestEmailPattern(t *testing.T) {
	testCases := []struct {
		name  string
		email string
		match bool
	}{
		{
			name:  "不带@",
			email: "123456",
			match: false,
		},
		{
			name:  "带@ 但是没后缀",
			email: "123456@",
			match: false,
		},
		{
			name:  "合法邮箱",
			email: "123456@qq.com",
			match: true,
		},
	}

	h := NewUserHandler(nil, nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := h.emailRegexExp.MatchString(tc.email)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}

func TestPasswordPattern(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		match    bool
	}{
		{
			name:     "合法密码",
			password: "Hello#world123",
			match:    true,
		},
		{
			name:     "没有数字",
			password: "Hello#world",
			match:    false,
		},
		{
			name:     "没有特殊字符",
			password: "Helloworld123",
			match:    false,
		},
		{
			name:     "长度不足",
			password: "he!123",
			match:    false,
		},
	}

	h := NewUserHandler(nil, nil)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := h.passwordRegexExp.MatchString(tc.password)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}
