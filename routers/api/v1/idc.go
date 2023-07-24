package v1

import (
	"errors"
	"ipam/utils/logging"
	"ipam/utils/tools"

	"ipam/pkg/idc"

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
	idc.IDC
}

// 创建机房回复
type Res struct {
	OK int `json:"ok"`
}

// 获取idc回复
type GetIdcRes struct {
	IDCS []idc.IDC `json: "idcs"`
}

// 获取IDC
func (*IDCResource) IdcList(c *gin.Context) {
	method := "GetIDC"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	idcs := idc.GetIDC()
	resp.Render(c, 200, GetIdcRes{IDCS: idcs}, nil)
	return
}

// 新建机房
func (*IDCResource) CreateIDC(c *gin.Context) {
	method := "CreateIDC"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}

	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		r := &req.IDC
		if err := r.CreateIDC(); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 删除机房
func (*IDCResource) DeleteIDC(c *gin.Context) {
	method := "DeleteIDC"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, "admin"); !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		r := &req.IDC
		if err := r.DeleteIDC(); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 新建VRF
func (*IDCResource) CreateVRF(c *gin.Context) {
	method := "CreateVRF"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		r := &req.IDC
		if err := r.CreateVRF(); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 删除VRF
func (*IDCResource) DeleteVRF(c *gin.Context) {
	method := "DeleteVRF"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		r := &req.IDC
		if err := r.DeleteVRF(); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 新建路由器
func (*IDCResource) CreateRouter(c *gin.Context) {
	method := "CreateRouter"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		r := &req.IDC
		if err := r.CreateRouter(); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}

// 删除路由器
func (*IDCResource) DeleteRouter(c *gin.Context) {
	method := "DeleteRouter"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	//获取前端数据
	var req Req
	if c.ShouldBind(&req.IDC) == nil {
		logging.Debug(req)
		r := &req.IDC
		if err := r.DeleteRouter(); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, Res{0}, nil)
	return
}
