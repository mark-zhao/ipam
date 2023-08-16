package audit

import (
	"context"
	"time"
)

type AuditInfo struct {
	Operator    string `bson:"operator" json:"operator"`       //操作员
	Func        string `bson:"func" json:"func"`               //函数
	Description string `bson:"description" json:"description"` //描述
	Date        string `bson:"date" json:"date"`               //分配时间
}

type Audits []AuditInfo

func (ar *audit) Add(ctx context.Context, a *AuditInfo) error {
	_, err := ar.storage.CreateAudit(ctx, a)
	return err
}

func (ar *audit) List(ctx context.Context, st, et time.Time) (Audits, error) {
	return ar.storage.ReadAllAudit(ctx, st, et)
}

func (A AuditInfo) deepCopy() *AuditInfo {
	return &AuditInfo{
		Operator:    A.Operator,
		Func:        A.Func,
		Description: A.Description,
		Date:        A.Date,
	}
}
