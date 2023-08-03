package audit

import (
	"context"
	"time"
)

// Auditer can be used to do Audit stuff.
type Auditer interface {
	Add(ctx context.Context, a *AuditInfo) error
	List(ctx context.Context, st, et time.Time) (Audits, error)
}

type audit struct {
	storage Storage
}

// NewWithStorage allows you to create a Dcim instance with your Storage implementation.
// The Storage interface must be implemented.
func NewWithStorage(storage Storage) Auditer {
	return &audit{storage: storage}
}
