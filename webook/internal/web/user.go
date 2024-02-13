package web

import (
	"errors"
	"net/http"
	"time"
	"unicode/utf8"
	"webook/internal/domain"
	"webook/internal/service"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	UserIdKey            = "userId"
	bizLogin             = "login"
)

var JWTKey = []byte("WiXLWadWG44Rr2qP6VUDLod0dzAnRI45")

type UserHandler struct {
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	userService      *service.UserService
	codeService      *service.CodeService
}

func NewUserHandler(userService *service.UserService, codeService *service.CodeService) *UserHandler {
	return &UserHandler{
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		userService:      userService,
		codeService:      codeService,
	}
}

func (uh *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", uh.SignUp)
	ug.POST("/login", uh.Login)
	ug.POST("/edit", uh.Edit)
	ug.GET("/profile", uh.Profile)

	// 手机验证码相关功能
	ug.POST("/login_sms/code/send", uh.SendSMSLoginCode)
	ug.POST("/login_sms", uh.LoginSMS)
}

func (uh *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "请输入手机号",
		})
	}
	err := uh.codeService.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
}

func (uh *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := uh.codeService.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码不对，请重新输入",
		})
		return
	}
	u, err := uh.userService.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	uh.setJWTToken(ctx, u.Id)
	ctx.JSON(http.StatusOK, Result{
		Msg: "登陆成功",
	})
}

func (uh *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	isEmail, err := uh.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := uh.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = uh.userService.SignUp(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateUser:
		ctx.String(http.StatusOK, "重复邮箱，请换一个邮箱")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (uh *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	user, err := uh.userService.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		uh.setJWTToken(ctx, user.Id)
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (uh *UserHandler) setJWTToken(ctx *gin.Context, uid int64) {
	uc := UserClaims{
		Uid:       uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
}

func (uh *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// 获取字符串真实字符长度
	nicknameLength := utf8.RuneCountInString(req.Nickname)
	if nicknameLength < 2 || nicknameLength > 16 {
		ctx.String(http.StatusOK, "昵称长度不可小于2位或是大于16位")
		return
	}

	isBirthdayValid, err := uh.checkIfBirthdayIsValid(req.Birthday)
	if !isBirthdayValid {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	// 获取字符串总长度
	aboutMeLength := utf8.RuneCountInString(req.AboutMe)
	if aboutMeLength > 1024 {
		ctx.String(http.StatusOK, "关于我的内容不可超过1024个字符")
		return
	}
	sess := sessions.Default(ctx)
	userId := sess.Get(UserIdKey)
	err = uh.userService.Edit(ctx, userId.(int64), req.Nickname, req.Birthday, req.AboutMe)
	if err != nil {
		ctx.String(http.StatusOK, "用户信息更新失败")
		return
	}
	ctx.String(http.StatusOK, "更新成功")
}

func (uh *UserHandler) Profile(ctx *gin.Context) {
	// sess := sessions.Default(ctx)
	// userId := sess.Get(UserIdKey)
	uc := ctx.MustGet("user").(UserClaims)
	user, err := uh.userService.GetProfile(ctx, uc.Uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.JSON(http.StatusOK, UserInfo{
		Nickname: user.Nickname,
		Email:    user.Email,
		Birthday: user.Birthday,
		AboutMe:  user.AboutMe,
	})
}

func (uh *UserHandler) checkIfBirthdayIsValid(birthday string) (bool, error) {
	// 以特定格式解析日期
	parsedDate, err := time.Parse(time.DateOnly, birthday)
	if err != nil {
		return false, errors.New("非法生日")
	}

	// 获取当前日期并去除时分秒
	currentDate := time.Now()
	currentDate = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())

	// 比较日期
	if parsedDate.After(currentDate) {
		return false, errors.New("非法生日")
	}
	return true, nil
}

type UserInfo struct {
	Nickname string
	Email    string
	Birthday string
	AboutMe  string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}
