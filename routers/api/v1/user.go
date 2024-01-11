package v1

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"ipam/pkg/audit"
	Administrator "ipam/pkg/user"
	"ipam/utils/aeser"
	"ipam/utils/logging"
	"ipam/utils/tools"
	"time"

	"github.com/gin-gonic/gin"
)

const modelUser string = "User"

type UserResource struct {
}

// 注册路由
func UserRouter() {
	p := Administrator.Permission{
		Id:    1,
		Label: modelUser,
		Children: []Administrator.Permission2{
			{Id: 11, Label: "ListUser"},
			{Id: 12, Label: "CreateUser"},
			{Id: 13, Label: "DeleteUser"},
			{Id: 14, Label: "Changer"},
		},
	}
	Permissions = append(Permissions, p)
	APIs["/user"] = map[UriInterface]interface{}{
		NewUri("GET", "/ListUser"):    (&UserResource{}).ListUser,
		NewUri("POST", "/CreateUser"): (&UserResource{}).CreateUser,
		NewUri("POST", "/DeleteUser"): (&UserResource{}).DeleteUser,
		NewUri("POST", "/Changer"):    (&UserResource{}).Changer,
	}
}

// 获取用户list
type GetUsersRes struct {
	Users       []Administrator.User       `json:"users"`
	Permissions []Administrator.Permission `json:"permissions"`
}

// 获取所有用户
func (*UserResource) ListUser(c *gin.Context) {
	const method = "ListUser"
	logging.Info("开始", method)
	username, _ := tools.FunAuth(c, modelUser, method)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if username == "admin" {
		if users, err := admin.List(ctx); err != nil {
			logging.Info("录入数据库失败:", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			resp.Render(c, 200, GetUsersRes{Users: users, Permissions: Permissions}, nil)
			return
		}
	} else {
		if user, err := admin.Get(ctx, username); err != nil {
			logging.Info("查询数据库失败:", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			resp.Render(c, 200, GetUsersRes{Users: []Administrator.User{user}, Permissions: Permissions}, nil)
			return
		}
	}
}

// 新建用户
func (*UserResource) CreateUser(c *gin.Context) {
	const method = "CreateUser"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelUser, method)
	if !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Administrator.User
	if c.ShouldBind(&req) == nil {
		if err := admin.Add(ctx, &req); err != nil {
			logging.Info("录入数据库失败:", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: fmt.Sprintf("创建用户%s成功", req.Name),
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
		resp.Render(c, 200, Res{0}, nil)
		return
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// 删除用户
func (*UserResource) DeleteUser(c *gin.Context) {
	const method = "DeleteUser"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelUser, method)
	if !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Administrator.User
	if c.ShouldBind(&req) == nil {
		if req.Name == "admin" || req.Name == username {
			resp.Render(c, 200, nil, errors.New("不能删除admin和自己用户"))
			return
		}
		if err := admin.Del(ctx, req.Name); err != nil {
			logging.Info("录入数据库失败:", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: fmt.Sprintf("删除用户%s成功", req.Name),
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// 修改用户密码
func (*UserResource) Changer(c *gin.Context) {
	const method = "Changer"
	logging.Info("开始", method)
	username, _ := tools.FunAuth(c, modelUser, method)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Administrator.User
	if c.ShouldBind(&req) == nil {
		if req.Phone == 0 || req.Name == "" {
			resp.Render(c, 200, nil, errors.New("手机号和用户名不能为空"))
			return
		}
		ouser, err := admin.Get(ctx, req.Name)
		if err != nil {
			resp.Render(c, 200, nil, fmt.Errorf("获取用户 %s 失败", req.Name))
			return
		}
		if username == "admin" {
			if req.Pwd == "111111" {
				encryptResult, err := hex.DecodeString(ouser.Pwd)
				if err != nil {
					resp.Render(c, 200, nil, fmt.Errorf("解密失败"))
					return
				}
				hexKey := "6c1acf9ad6f12ff7e3c5b94df9f9ef329996b6ea7d148afafe76765d42d0a876"
				key, err := hex.DecodeString(hexKey)
				if err != nil {
					resp.Render(c, 200, nil, fmt.Errorf("解密失败"))
					return
				}
				Pwresult, err := aeser.AESDecrypt(encryptResult, key)
				if err != nil {
					resp.Render(c, 200, nil, fmt.Errorf("解密失败"))
					return
				}
				req.Pwd = string(Pwresult)
			}
			if req.Permission == nil {
				req.Permission = ouser.Permission
			} else if ouser.Name == "admin" {
				req.Permission = ouser.Permission
			}
			if err := admin.Changer(ctx, &req); err != nil {
				logging.Info("录入数据库失败:", err)
				resp.Render(c, 200, nil, err)
				return
			} else {
				a := &audit.AuditInfo{
					Operator:    username,
					Func:        method,
					Description: fmt.Sprintf("修改用户%s成功", req.Name),
					Date:        tools.DateToString(),
				}
				if err := auditer.Add(ctx, a); err != nil {
					logging.Error("audit insert mongo error:", err)
				}
			}
		}
		if username == req.Name || username != "admin" {
			req.Permission = ouser.Permission
			if req.Pwd == "111111" {
				encryptResult, err := hex.DecodeString(ouser.Pwd)
				if err != nil {
					resp.Render(c, 200, nil, fmt.Errorf("解密失败"))
					return
				}
				hexKey := "6c1acf9ad6f12ff7e3c5b94df9f9ef329996b6ea7d148afafe76765d42d0a876"
				key, err := hex.DecodeString(hexKey)
				if err != nil {
					resp.Render(c, 200, nil, fmt.Errorf("解密失败"))
					return
				}
				Pwresult, err := aeser.AESDecrypt(encryptResult, key)
				if err != nil {
					resp.Render(c, 200, nil, fmt.Errorf("解密失败"))
					return
				}
				logging.Info(string(Pwresult))
				req.Pwd = string(Pwresult)
			}
			if err := admin.Changer(ctx, &req); err != nil {
				logging.Info("录入数据库失败:", err)
				resp.Render(c, 200, nil, err)
				return
			} else {
				a := &audit.AuditInfo{
					Operator:    username,
					Func:        method,
					Description: fmt.Sprintf("修改用户%s成功", req.Name),
					Date:        tools.DateToString(),
				}
				if err := auditer.Add(ctx, a); err != nil {
					logging.Error("audit insert mongo error:", err)
				}
			}
		} else {
			resp.Render(c, 200, Res{1}, nil)
			return
		}
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}
