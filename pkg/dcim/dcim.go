package dcim

import (
	"context"
	"sync"
)

// Dcim can be used to do IDC stuff.
type Dcimer interface {
	// 获取idcinfo
	GetIDCINFO(ctx context.Context)
	// NewIDC create a new IDC from a string notation.
	NewIDC(idcname, description string, Router []Router, vrf []string) (*IDC, error)
	//新建机房
	CreateIDC(ctx context.Context, i IDC) error
	//删除机房
	DeleteIDC(ctx context.Context, idcname string) error
	// 新建VRF
	CreateVRF(ctx context.Context, i IDC) error
	// 删除VRF
	DeleteVRF(ctx context.Context, i IDC) error
	// 新建路由
	CreateRouter(ctx context.Context, i IDC) error
	// 删除路由
	DeleteRouter(ctx context.Context, i IDC) error
}

type dcim struct {
	mu      sync.Mutex
	storage Storage
}

// NewWithStorage allows you to create a Dcim instance with your Storage implementation.
// The Storage interface must be implemented.
func NewWithStorage(storage Storage) Dcimer {
	return &dcim{storage: storage}
}
