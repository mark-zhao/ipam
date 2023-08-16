package note

import (
	"context"
	"encoding/json"
	"fmt"
	modelv1 "ipam/model/v1"
	"log"

	"github.com/jmoiron/sqlx" // MySQL driver
)

const dbTable = "notes"

// mysqlDB is a note.Storage implementation backed by a MySQL database
type sql struct {
	db *sqlx.DB
}

// MySQL's implementation of the Storage interface methods

func NewMySQL(ctx context.Context, config modelv1.MySqlConfig) (Storage, error) {
	db, err := sqlx.Connect("mysql", config.DSN)
	if err != nil {
		log.Fatalln(err)
	}
	return newMySQL(ctx, db)
}
func newMySQL(ctx context.Context, db *sqlx.DB) (*sql, error) {
	return &sql{db: db}, nil
}

func (s *sql) Name() string {
	return "mysql"
}
func (s *sql) noteExists(ctx context.Context, n *Note) (*Note, bool) {
	Note, err := s.ReadNote(ctx, n.Instance)
	if err != nil {
		return nil, false
	}
	return Note, true
}

func (s *sql) CreateNote(ctx context.Context, n *Note) (*Note, error) {
	existingNote, exists := s.noteExists(ctx, n)
	if exists {
		return existingNote, nil
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return &Note{}, fmt.Errorf("unable to start transaction:%w", err)
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO notes (instance, description, operator, date) VALUES ($1, $2, $3, $4)", n.Instance, n.Description, n.Operator, n.Date)
	if err != nil {
		return &Note{}, fmt.Errorf("unable to insert notes:%w", err)
	}

	return n, nil
}

func (s *sql) DeleteAllNote(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM notes")
	return err
}

func (s *sql) ReadAllNote(ctx context.Context) (Notes, error) {
	var notes [][]byte
	err := s.db.SelectContext(ctx, &notes, "SELECT prefix FROM notes")
	if err != nil {
		return nil, fmt.Errorf("unable to read prefixes:%w", err)
	}
	result := Notes{}
	for _, v := range notes {
		note, err := fromJSON(v)
		if err != nil {
			return nil, err
		}
		result = append(result, *note)
	}
	return result, nil
}

func (s *sql) DeleteNote(ctx context.Context, instance string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("unable to start transaction:%w", err)
	}
	_, err = tx.ExecContext(ctx, "DELETE from prefixes WHERE instance=$1", instance)
	if err != nil {
		return fmt.Errorf("unable delete prefix:%w", err)
	}
	return tx.Commit()
}

func (s *sql) ReadNote(ctx context.Context, instance string) (*Note, error) {
	var result []byte
	err := s.db.GetContext(ctx, &result, "SELECT prefix FROM prefixes WHERE instance=$1", instance)
	if err != nil {
		return &Note{}, fmt.Errorf("unable to read prefix:%w", err)
	}
	return fromJSON(result)
}

func fromJSON(js []byte) (*Note, error) {
	var note Note
	err := json.Unmarshal(js, &note)
	if err != nil {
		return &Note{}, fmt.Errorf("unable to unmarshal note:%w", err)
	}
	return &note, nil
}
