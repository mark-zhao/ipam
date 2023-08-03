package v1

import (
	"context"
	"errors"
	"fmt"
	"ipam/utils/logging"
	"ipam/utils/tools"
	"time"

	"ipam/pkg/audit"
	"ipam/pkg/dcim"
	idc "ipam/pkg/dcim"

	"github.com/gin-gonic/gin"
)

const modelIDC string = "IDC"

type IDCResource struct {
}

// 注册路由
func IDCRouter() {
	APIs["/idc"] = map[UriInterface]interface{}{
		NewUri("GET", "/IdcList"):       (&IDCResource{}).IdcList,
		NewUri("POST", "/CreateIDC"):    (&IDCResource{}).CreateIDC,
		NewUri("POST", "/DeleteIDC"):    (&IDCResource{}).DeleteIDC,
		NewUri("POST", "/CreateVRF"):    (&IDCResource{}).CreateVRF,
		NewUri("POST", "/DeleteVRF"):    (&IDCResource{}).DeleteVRF,
		NewUri("POST", "/CreateRouter"): (&IDCResource{}).CreateRouter,
		NewUri("POST", "/DeleteRouter"): (&IDCResource{}).DeleteRouter,
	}
}

// 创建机房请求
type Req struct {
	dcim.IDC
}

// 创建机房回复
type Res struct {
	OK int `json:"ok"`
}

// 获取idc回复
type GetIdcRes struct {
	IDCS []dcim.IDC `json:"idcs"`
}

// 获取IDC
func (*IDCResource) IdcList(c *gin.Context) {
	method := "IdcList"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	idcs := idc.IDCs
	resp.Render(c, 200, GetIdcRes{IDCS: idcs}, nil)
	return
}

// 新建机房
func (*IDCResource) CreateIDC(c *gin.Context) {
	method := "CreateIDC"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIDC, method)
	if !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		if err := dcimer.CreateIDC(ctx, req.IDC); err != nil {
			logging.Info("录入数据库失败:", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: req.IDC.IDCName,
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 删除机房
func (*IDCResource) DeleteIDC(c *gin.Context) {
	method := "DeleteIDC"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIDC, "admin")
	if !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		if err := dcimer.DeleteIDC(ctx, req.IDC.IDCName); err != nil {
			logging.Info("录入数据库失败:", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: req.IDC.IDCName,
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
			if _, err = ipam.DeletePrefix(ctx, req.IDC.IDCName, true); err != nil {
				logging.Error(err)
				resp.Render(c, 200, Res{0}, fmt.Errorf("delete prefix: %w", err))
				return
			}
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 新建VRF
func (*IDCResource) CreateVRF(c *gin.Context) {
	method := "CreateVRF"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIDC, method)
	if !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		if err := dcimer.CreateVRF(ctx, req.IDC); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: req.IDC.VRF[0],
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 删除VRF
func (*IDCResource) DeleteVRF(c *gin.Context) {
	method := "DeleteVRF"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIDC, method)
	if !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		if err := dcimer.DeleteVRF(ctx, req.IDC); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: req.IDC.VRF[0],
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 新建路由器
func (*IDCResource) CreateRouter(c *gin.Context) {
	method := "CreateRouter"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIDC, method)
	if !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		r := &req.IDC
		if err := dcimer.CreateRouter(ctx, req.IDC); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: r.Router[0].IP,
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 删除路由器
func (*IDCResource) DeleteRouter(c *gin.Context) {
	method := "DeleteRouter"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIDC, method)
	if !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		if err := dcimer.DeleteRouter(ctx, req.IDC); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: req.Router[0].IP,
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}
