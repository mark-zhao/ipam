package Administrator

import (
	"context"
)

// Auditer can be used to do Audit stuff.
type Administrator interface {
	List(ctx context.Context) ([]User, error)
	Get(ctx context.Context, username string) (User, error)
	Add(ctx context.Context, user *User) error
	Del(ctx context.Context, username string) error
	Changer(ctx context.Context, user *User) error
	LoginCheck(ctx context.Context, loginReq LoginReq) (bool, User)
}

type user struct {
	storage Storage
}

// NewWithStorage allows you to create a Dcim instance with your Storage implementation.
// The Storage interface must be implemented.
func NewWithStorage(storage Storage) Administrator {
	return &user{storage: storage}
}
