package dcim

import (
	"context"
	"fmt"
	"sync"
)

type memory struct {
	idcs map[string]IDC
	lock sync.RWMutex
}

// NewMemory create a memory storage for ipam
func NewMemory() Storage {
	idcs := make(map[string]IDC)
	return &memory{
		idcs: idcs,
		lock: sync.RWMutex{},
	}
}
func (m *memory) Name() string {
	return "memory"
}

func (m *memory) CreateIDC(ctx context.Context, idc IDC) (IDC, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, ok := m.idcs[idc.IDCName]
	if ok {
		return IDC{}, fmt.Errorf("idc already created:%v", idc.IDCName)
	}
	m.idcs[idc.IDCName] = *idc.deepCopy()
	return idc, nil
}

func (m *memory) ReadIDC(ctx context.Context, idcname string) (IDC, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	result, ok := m.idcs[idcname]
	if !ok {
		return IDC{}, fmt.Errorf("note %s not found", idcname)
	}
	return *result.deepCopy(), nil
}
func (m *memory) DeleteAllIDC(ctx context.Context) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	m.idcs = make(map[string]IDC)
	return nil

}
func (m *memory) ReadAllIDC(ctx context.Context) (IDCS, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	is := make(IDCS, 0, len(m.idcs))
	for _, v := range m.idcs {
		is = append(is, *v.deepCopy())
	}
	return is, nil

}
func (m *memory) UpdateIDC(ctx context.Context, idc IDC) (IDC, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	_, ok := m.idcs[idc.IDCName]
	if !ok {
		return IDC{}, fmt.Errorf("note %s not found", idc.IDCName)
	}
	m.idcs[idc.IDCName] = *idc.deepCopy()
	return idc, nil
}
func (m *memory) DeleteIDC(ctx context.Context, idcname string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.idcs, idcname)

	return nil
}
