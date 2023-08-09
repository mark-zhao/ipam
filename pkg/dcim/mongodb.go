package dcim

import (
	"context"
	"errors"
	"fmt"
	modelv1 "ipam/model/v1"
	"ipam/utils/logging"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbIndex = `idcname`

type MongoConfig struct {
	DatabaseName       string
	CollectionName     string
	MongoClientOptions *options.ClientOptions
}

type mongodb struct {
	c    *mongo.Collection
	lock sync.RWMutex
}

func NewMongo(ctx context.Context, database *mongo.Client, conf modelv1.MongoConfig) (Storage, error) {
	return newMongo(ctx, database, conf)
}

func (m *mongodb) Name() string {
	return "mongodb"
}

func newMongo(ctx context.Context, m *mongo.Client, conf modelv1.MongoConfig) (*mongodb, error) {
	c := m.Database(conf.DatabaseName).Collection(conf.CollectionName)

	_, err := c.Indexes().CreateMany(ctx, []mongo.IndexModel{{
		Keys:    bson.M{dbIndex: 1},
		Options: options.Index().SetUnique(true),
	}})
	if err != nil {
		return nil, err
	}
	return &mongodb{c, sync.RWMutex{}}, nil
}

func (m *mongodb) CreateIDC(ctx context.Context, idc IDC) (IDC, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, cErr := m.c.InsertOne(ctx, &idc)
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return IDC{}, errors.New("数据录入数据库失败")
	}

	return idc, nil
}

func (m *mongodb) ReadIDC(ctx context.Context, idcname string) (IDC, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	filter := bson.M{"idcname": idcname}
	var iOld IDC
	if err := m.c.FindOne(ctx, filter).Decode(&iOld); err != nil {
		return IDC{}, fmt.Errorf(`不存在名字为%s的idc`, idcname)
	}
	return iOld, nil
}

func (m *mongodb) DeleteAllIDC(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	f := bson.D{{}} // match all documents
	_, err := m.c.DeleteMany(ctx, f)
	if err != nil {
		return fmt.Errorf(`error deleting all idc: %w`, err)
	}
	return nil
}

func (m *mongodb) ReadAllIDC(ctx context.Context) (idcs IDCS, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	f := bson.D{} // match all documents
	cur, err := m.c.Find(ctx, f)
	if err != nil {
		return nil, fmt.Errorf(`error reading all idcs: %w`, err)
	}
	defer cur.Close(ctx)
	if err = cur.All(ctx, &idcs); err != nil {
		logging.Error("获取数据失败")
		return
	}
	return
}

func (m *mongodb) UpdateIDC(ctx context.Context, idc IDC) (IDC, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	filter := bson.M{"idcname": idc.IDCName}

	o := options.Replace().SetUpsert(false)
	r, err := m.c.ReplaceOne(ctx, filter, idc, o)
	if err != nil {
		return IDC{}, fmt.Errorf("unable to update idc:%s, error: %w", idc.IDCName, err)
	}
	if r.MatchedCount == 0 {
		return IDC{}, fmt.Errorf("unable to update idc:%s", idc.IDCName)
	}
	if r.ModifiedCount == 0 {
		return IDC{}, fmt.Errorf("update did not effect any document:%s",
			idc.IDCName)
	}

	return idc, nil
}

func (m *mongodb) DeleteIDC(ctx context.Context, idcname string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	filter := bson.M{"idcname": idcname}
	r := m.c.FindOneAndDelete(ctx, filter)

	// ErrNoDocuments should be returned if the idc does not exist
	if r.Err() != nil && errors.Is(r.Err(), mongo.ErrNoDocuments) {
		return fmt.Errorf(`idc not found:%s, error:%w`, idcname, r.Err())
	} else if r.Err() != nil {
		return fmt.Errorf(`error while trying to find idc:%s, error:%w`, idcname, r.Err())
	}
	return nil
}
