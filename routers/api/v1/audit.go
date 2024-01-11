package v1

import (
	"context"
	"errors"
	"fmt"
	Administrator "ipam/pkg/user"
	"ipam/utils/logging"
	"ipam/utils/tools"
	"time"

	"github.com/gin-gonic/gin"
)

const modelAudit string = "AUDIT"

type AuditResource struct {
}

// 注册路由
func AuditRouter() {
	p := Administrator.Permission{
		Id:    5,
		Label: modelAudit,
		Children: []Administrator.Permission2{
			{Id: 51, Label: "audit"},
		},
	}
	Permissions = append(Permissions, p)
	APIs["/audit"] = map[UriInterface]interface{}{
		NewUri("POST", "/AuditList"): (&AuditResource{}).AuditList,
	}
}

type AuditListReq struct {
	St string `json:"st"`
	Et string `json:"et"`
}

func (*AuditResource) AuditList(c *gin.Context) {
	layout := "2006-01-02 15:04:05" // 根据你的输入日期格式进行调整
	const method = "AuditList"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelAudit, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	//获取前端数据
	var req AuditListReq
	if c.ShouldBind(&req) == nil {
		logging.Debug(req)
		st, err := time.Parse(layout, req.St)
		if err != nil {
			fmt.Println("解析开始时间失败:", err)
			resp.Render(c, 200, nil, errors.New("解析开始时间失败"))
			return
		}
		et, err := time.Parse(layout, req.Et)
		if err != nil {
			fmt.Println("解析结束时间失败:", err)
			resp.Render(c, 200, nil, errors.New("解析开始时间失败"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		audits, err := auditer.List(ctx, st, et)
		if err == nil && audits != nil {
			resp.Render(c, 200, audits, nil)
			return
		}
	}
	resp.Render(c, 200, nil, errors.New("没有audits"))
}
