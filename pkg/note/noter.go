package note

import (
	"context"
)

// Dcim can be used to do NOTE stuff.
type Noter interface {
	//新建机房
	CreateNote(ctx context.Context, n *Note) error
	//删除机房
	DeleteNote(ctx context.Context, instance string) error
	// 新建VRF
	NoteList(ctx context.Context) (Notes, error)
}

type note struct {
	storage Storage
}

// NewWithStorage allows you to create a Dcim instance with your Storage implementation.
// The Storage interface must be implemented.
func NewWithStorage(storage Storage) Noter {
	return &note{storage: storage}
}
