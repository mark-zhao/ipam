package v1

import (
	"context"
	"errors"
	"ipam/pkg/audit"
	"ipam/pkg/note"
	"ipam/utils/logging"
	"ipam/utils/tools"
	"time"

	"github.com/gin-gonic/gin"
)

type NOTEResource struct {
}

const modelNOTE string = "NOTE"

// 注册路由
func NOTERouter() {
	APIs["/note"] = map[UriInterface]interface{}{
		NewUri("GET", "/NoteList"):    (&NOTEResource{}).NoteList,
		NewUri("POST", "/CreateNote"): (&NOTEResource{}).CreateNote,
		NewUri("POST", "/DeleteNote"): (&NOTEResource{}).DeleteNote,
	}
}

func (*NOTEResource) NoteList(c *gin.Context) {
	const method = "NoteList"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelNOTE, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	notes, err := noter.NoteList(ctx)
	if err == nil && notes != nil {
		resp.Render(c, 200, notes, nil)
		return
	}
	resp.Render(c, 200, nil, errors.New("没有note"))
}

func (*NOTEResource) CreateNote(c *gin.Context) {
	const method = "CreateNote"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelNOTE, method)
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := noter.CreateNote(ctx, r); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: req.Instance,
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
	}
	resp.Render(c, 200, nil, nil)
}

func (*NOTEResource) DeleteNote(c *gin.Context) {
	const method = "DeleteNote"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelNOTE, method)
	if !ok {
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := noter.DeleteNote(ctx, req.Instance); err != nil {
			logging.Info("录入数据库失败", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: req.Instance,
				Date:        tools.DateToString(),
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
		}
	}
	resp.Render(c, 200, nil, nil)
}
