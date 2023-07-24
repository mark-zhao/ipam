package v1

import (
	"errors"
	"ipam/pkg/note"
	"ipam/utils/logging"
	"ipam/utils/tools"

	"github.com/gin-gonic/gin"
)

type NOTEResource struct {
}

// 注册路由
func NOTERouter() {
	APIs["/note"] = map[UriInterface]interface{}{
		NewUri("GET", "/NoteList"):    (&NOTEResource{}).NoteList,
		NewUri("POST", "/CreateNote"): (&NOTEResource{}).CreateNote,
		NewUri("POST", "/DeleteNote"): (&NOTEResource{}).DeleteNote,
	}
}

func (*NOTEResource) NoteList(c *gin.Context) {
	method := "NoteList"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	n := note.Note{}
	notes, err := n.NoteList()
	if err == nil && notes != nil {
		resp.Render(c, 200, notes, nil)
		return
	}
	resp.Render(c, 200, nil, errors.New("没有note"))
	return
}

func (*NOTEResource) CreateNote(c *gin.Context) {
	method := "CreateNote"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIDC, method)
	if !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	//获取前端数据
	var req note.Note
	if c.ShouldBind(&req) == nil {
		if req.Instance == "" || req.Description == "" {
			logging.Error("实例或描述不能为空")
			resp.Render(c, 200, nil, errors.New("实例或描述不能为空"))
			return
		}
		r := &req
		r.Operator = username
		r.Date = tools.DateToString()
		if err := r.CreateNote(); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, nil, nil)
	return
}

func (*NOTEResource) DeleteNote(c *gin.Context) {
	method := "DeleteNote"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIDC, method); !ok {
		logging.Info("没有权限访问")
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	//获取前端数据
	var req note.Note
	if c.ShouldBind(&req) == nil {
		if req.Instance == "" {
			logging.Error("实例不能为空")
			resp.Render(c, 200, nil, errors.New("实例不能为空"))
			return
		}
		r := &req
		if err := r.DeleteNote(); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, nil, nil)
	return
}
