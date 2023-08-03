package note

import (
	"context"
	"errors"
	"fmt"
	"ipam/utils/logging"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbIndex = `instance`

type MongoConfig struct {
	DatabaseName       string
	CollectionName     string
	MongoClientOptions *options.ClientOptions
}

type mongodb struct {
	c    *mongo.Collection
	lock sync.RWMutex
}

func NewMongo(ctx context.Context, config MongoConfig) (Storage, error) {
	return newMongo(ctx, config)
}

func (m *mongodb) Name() string {
	return "mongodb"
}

func newMongo(ctx context.Context, config MongoConfig) (*mongodb, error) {
	m, err := mongo.NewClient(config.MongoClientOptions)
	if err != nil {
		return nil, err
	}
	err = m.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = m.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	c := m.Database(config.DatabaseName).Collection(config.CollectionName)

	_, err = c.Indexes().CreateMany(ctx, []mongo.IndexModel{{
		Keys:    bson.M{dbIndex: 1},
		Options: options.Index().SetUnique(true),
	}})
	if err != nil {
		return nil, err
	}
	return &mongodb{c, sync.RWMutex{}}, nil
}

func (m *mongodb) CreateNote(ctx context.Context, n *Note) (*Note, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	filter := bson.M{"instance": n.Instance}
	var nOld Note
	m.c.FindOne(ctx, filter).Decode(&nOld)
	if nOld.Instance == n.Instance {
		return nil, errors.New("实例重名了")
	}
	_, cErr := m.c.InsertOne(ctx, n)
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return nil, errors.New("数据录入数据库失败")
	}
	return n, nil
}

func (m *mongodb) DeleteAllNote(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	f := bson.D{{}} // match all documents
	_, err := m.c.DeleteMany(ctx, f)
	if err != nil {
		return fmt.Errorf(`error deleting all notes: %w`, err)
	}
	return nil
}

func (m *mongodb) ReadAllNote(ctx context.Context) (notes Notes, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	f := bson.D{{}} // match all documents
	cur, err := m.c.Find(ctx, f)
	if err != nil {
		return nil, fmt.Errorf(`error reading all notes: %w`, err)
	}
	defer cur.Close(ctx)
	if err = cur.All(ctx, &notes); err != nil {
		logging.Error("获取数据失败")
		return
	}
	return
}

func (m *mongodb) DeleteNote(ctx context.Context, instance string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	filter := bson.M{"instance": instance}
	r := m.c.FindOneAndDelete(ctx, filter)

	// ErrNoDocuments should be returned if the note does not exist
	if r.Err() != nil && errors.Is(r.Err(), mongo.ErrNoDocuments) {
		return fmt.Errorf(`note not found:%s, error:%w`, instance, r.Err())
	} else if r.Err() != nil {
		return fmt.Errorf(`error while trying to find note:%s, error:%w`, instance, r.Err())
	}
	return nil
}
