package v1

import (
	"ipam/utils/logging"
	"ipam/utils/options"
)

const (
	dbName     = "myBlog.db"
	userBucket = "user"
)

// User 用户类
//type User struct {
//	Id         string `json:"userId"`
//	Name       string `json:"userName"`
//	Gender     string `json:"gender"`
//	Phone      string `json:"userMobile"`
//	Pwd        string `json:"pwd"`
//	Permission string `json:"permission"`
//}

// LoginReq 登录请求参数类
type LoginReq struct {
	Name string `json:"name" form:"name"`
	Pwd  string `json:"pwd" form:"pwd"`
}

func LoginCheck(loginReq LoginReq) (bool, options.User) {
	if v, ok := options.Conf.UserList[loginReq.Name]; ok && v.Pwd == loginReq.Pwd {
		logging.Info("登录成功")
		return true, v
	}
	logging.Error("用户名或密码错误")
	return false, options.User{}
}
