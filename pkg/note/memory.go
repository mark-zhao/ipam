package note

import (
	"context"
	"fmt"
	"sync"
)

type memory struct {
	notes map[string]Note
	lock  sync.RWMutex
}

// NewMemory create a memory storage for ipam
func NewMemory() Storage {
	notes := make(map[string]Note)
	return &memory{
		notes: notes,
		lock:  sync.RWMutex{},
	}
}
func (m *memory) Name() string {
	return "memory"
}
func (m *memory) CreateNote(ctx context.Context, note *Note) (*Note, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, ok := m.notes[note.Instance]
	if ok {
		return nil, fmt.Errorf("note already created:%v", &note)
	}
	m.notes[note.Instance] = *note.deepCopy()
	return note, nil
}
func (m *memory) ReadNote(_ context.Context, instance string) (Note, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	result, ok := m.notes[instance]
	if !ok {
		return Note{}, fmt.Errorf("note %s not found", instance)
	}
	return *result.deepCopy(), nil
}
func (m *memory) DeleteAllNote(_ context.Context) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	m.notes = make(map[string]Note)
	return nil
}
func (m *memory) ReadAllNote(_ context.Context) (Notes, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ps := make(Notes, 0, len(m.notes))
	for _, v := range m.notes {
		ps = append(ps, *v.deepCopy())
	}
	return ps, nil
}

func (m *memory) DeleteNote(ctx context.Context, instance string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.notes, instance)

	return nil
}
