package audit

import (
	"context"
	"time"
)

// Storage is a interface to store dcim objects.
type Storage interface {
	Name() string
	CreateAudit(ctx context.Context, audit *AuditInfo) (*AuditInfo, error)
	ReadAllAudit(ctx context.Context, st, et time.Time) (Audits, error)
}
