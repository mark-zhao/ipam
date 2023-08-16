package audit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type memory struct {
	audits map[string]AuditInfo
	lock   sync.RWMutex
}

// NewMemory create a memory storage for ipam
func NewMemory() Storage {
	audits := make(map[string]AuditInfo)
	return &memory{
		audits: audits,
		lock:   sync.RWMutex{},
	}
}
func (m *memory) Name() string {
	return "memory"
}

func (m *memory) CreateAudit(ctx context.Context, audit *AuditInfo) (*AuditInfo, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, ok := m.audits[audit.Date]
	if ok {
		return nil, fmt.Errorf("audit already created:%v", &audit)
	}
	m.audits[audit.Date] = *audit.deepCopy()
	return audit, nil
}

func (m *memory) ReadAllAudit(ctx context.Context, st, et time.Time) (Audits, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	as := make(Audits, 0, len(m.audits))
	for k, v := range m.audits {
		t, err := time.Parse("2006-01-02 15:04:05", k)
		if err != nil {
			return nil, err
		}
		if et.After(t) && t.After(st) {
			as = append(as, *v.deepCopy())
		}
	}
	return as, nil
}
