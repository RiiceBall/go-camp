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
)

const (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	UserIdKey            = "userId"
)

type UserHandler struct {
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	userService      *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		userService:      svc,
	}
}

func (uh *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", uh.SignUp)
	ug.POST("/login", uh.Login)
	ug.POST("/edit", uh.Edit)
	ug.GET("/profile", uh.Profile)
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
	case service.ErrUserDuplicateEmail:
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
		sess := sessions.Default(ctx)
		sess.Set(UserIdKey, user.Id)
		sess.Options(sessions.Options{
			// 900 seconds
			MaxAge: 900,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
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
	sess := sessions.Default(ctx)
	userId := sess.Get(UserIdKey)
	user, err := uh.userService.GetProfile(ctx, userId.(int64))
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
