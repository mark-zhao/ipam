package note

import "context"

// Storage is a interface to store dcim objects.
type Storage interface {
	Name() string
	CreateNote(ctx context.Context, n *Note) (*Note, error)
	DeleteAllNote(ctx context.Context) error
	ReadAllNote(ctx context.Context) (Notes, error)
	DeleteNote(ctx context.Context, instance string) error
}
