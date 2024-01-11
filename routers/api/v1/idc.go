package v1

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"ipam/utils/aeser"
	"ipam/utils/logging"
	"ipam/utils/tools"
	"time"

	"ipam/pkg/audit"
	"ipam/pkg/dcim"
	Administrator "ipam/pkg/user"

	"github.com/gin-gonic/gin"
)

const modelIDC string = "IDC"

type IDCResource struct {
}

// 注册路由
func IDCRouter() {
	p := Administrator.Permission{
		Id:    4,
		Label: modelIDC,
		Children: []Administrator.Permission2{
			{Id: 41, Label: "IdcList"},
			{Id: 42, Label: "CreateIDC"},
			{Id: 43, Label: "DeleteIDC"},
			{Id: 44, Label: "CreateVRF"},
			{Id: 45, Label: "DeleteVRF"},
			{Id: 46, Label: "CreateRouter"},
			{Id: 47, Label: "DeleteRouter"},
		},
	}
	Permissions = append(Permissions, p)
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
	const method = "IdcList"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	idcs := dcim.IDCs
	resp.Render(c, 200, GetIdcRes{IDCS: idcs}, nil)
}

// 新建机房
func (*IDCResource) CreateIDC(c *gin.Context) {
	const method = "CreateIDC"
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
}

// 删除机房
func (*IDCResource) DeleteIDC(c *gin.Context) {
	const method = "DeleteIDC"
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
}

// 新建VRF
func (*IDCResource) CreateVRF(c *gin.Context) {
	const method = "CreateVRF"
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
		idsInfo := req.IDC
		if idsInfo.IDCName == "" || idsInfo.VRF == nil || idsInfo.VRF[0] == "" {
			logging.Info("idcname 和 vrf 不能为空")
			resp.Render(c, 200, nil, fmt.Errorf("idcname 和 vrf 不能为空"))
			return
		}
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
}

// 删除VRF
func (*IDCResource) DeleteVRF(c *gin.Context) {
	const method = "DeleteVRF"
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
}

// 新建路由器
func (*IDCResource) CreateRouter(c *gin.Context) {
	const method = "CreateRouter"
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
		//加密
		if r.Router == nil && r.Router[0].Password == "" {
			resp.Render(c, 200, nil, fmt.Errorf("密码为空"))
			return
		}
		hexKey := "6c1acf9ad6f12ff7e3c5b94df9f9ef329996b6ea7d148afafe76765d42d0a876"
		key, err := hex.DecodeString(hexKey)
		if err != nil {
			panic(err)
		}
		encryptResult, err := aeser.AESEncrypt([]byte(r.Router[0].Password), key)
		if err != nil {
			panic(err)
		}
		Pwresult := hex.EncodeToString(encryptResult)
		r.Router[0].Password = Pwresult
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
}

// 删除路由器
func (*IDCResource) DeleteRouter(c *gin.Context) {
	const method = "DeleteRouter"
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
}
