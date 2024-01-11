package Administrator

import (
	"context"
)

// Storage is a interface to store dcim objects.
type Storage interface {
	Name() string
	List(ctx context.Context) ([]User, error)
	Get(ctx context.Context, username string) (User, error)
	CreateUser(ctx context.Context, userinfo *User) error
	DelUser(ctx context.Context, username string) error
	Changer(ctx context.Context, user *User) error
}
