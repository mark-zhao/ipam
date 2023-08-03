package note

import (
	"context"
	"errors"
	"ipam/utils/logging"
)

type Notes []Note

// 机房信息
type Note struct {
	Instance    string `bson:"instance" json:"instance"`       //实例
	Description string `bson:"description" json:"description"` //描述
	Operator    string `bson:"operator" json:"operator"`       //操作员
	Date        string `bson:"date" json:"date"`               //分配时间
}

func (nr *note) CreateNote(ctx context.Context, n *Note) error {
	_, cErr := nr.storage.CreateNote(ctx, n)
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return errors.New("数据录入数据库失败")
	}
	return nil
}

func (nr *note) DeleteNote(ctx context.Context, instance string) error {
	cErr := nr.storage.DeleteNote(ctx, instance)
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return errors.New("数据录入数据库失败")
	}
	return nil
}

func (nr *note) NoteList(ctx context.Context) (notes Notes, cErr error) {
	notes, cErr = nr.storage.ReadAllNote(ctx)
	if cErr != nil {
		logging.Error("获取数据失败", cErr)
		return nil, cErr
	}
	return notes, nil
}
