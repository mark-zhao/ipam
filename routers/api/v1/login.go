package v1

import (
	"errors"
	"ipam/component"
	modelv1 "ipam/model/v1"
	"ipam/utils/except"
	"ipam/utils/logging"
	"ipam/utils/options"
	"net/http"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

type LoginResource struct {
}

// 注册信息
type RegistInfo struct {
	// 手机号
	Name string `json:"name"`
	// 密码
	Pwd string `json:"pwd"`
}

// LoginResult 登录结果结构
type LoginResult struct {
	Token        string `json:"token"`
	options.User `json:"user"`
}

func (l *LoginResource) Login(c *gin.Context) {
	var loginReq modelv1.LoginReq
	if c.Bind(&loginReq) == nil {
		isPass, user := modelv1.LoginCheck(loginReq)
		if isPass {
			generateToken(c, user)
			logging.Info("生成token 成功")
		} else {
			resp.Render(c, except.ERROR_AUTH_USER, nil, errors.New(except.GetMsg(except.ERROR_AUTH_USER)))
			logging.Error(except.ERROR_AUTH_USER)
			return
		}
	} else {
		resp.Render(c, except.INVALID_PARAMS, nil, errors.New(except.GetMsg(except.INVALID_PARAMS)))
		logging.Error(except.INVALID_PARAMS)
		return
	}
}

// 生成令牌
func generateToken(c *gin.Context, user options.User) {
	j := &component.JWT{
		[]byte("Woshinibaba"),
	}
	claims := component.CustomClaims{
		user.Id,
		user.Name,
		user.Phone,
		user.Permission,
		jwtgo.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000), // 签名生效时间
			ExpiresAt: int64(time.Now().Unix() + 3600), // 过期时间 一小时
			Issuer:    "mark",                          //签名的发行者
		},
	}

	token, err := j.CreateToken(claims)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    err.Error(),
		})
		return
	}
	user.Pwd = ""
	data := LoginResult{
		User:  user,
		Token: token,
	}
	resp.Render(c, except.SUCCESS, data, nil)
	return
}
