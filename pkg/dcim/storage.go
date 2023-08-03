package dcim

import "context"

// Storage is a interface to store dcim objects.
type Storage interface {
	Name() string
	CreateIDC(ctx context.Context, idc IDC) (IDC, error)
	ReadIDC(ctx context.Context, idcname string) (IDC, error)
	DeleteAllIDC(ctx context.Context) error
	ReadAllIDC(ctx context.Context) (IDCS, error)
	UpdateIDC(ctx context.Context, idc IDC) (IDC, error)
	DeleteIDC(ctx context.Context, idcname string) error
}
