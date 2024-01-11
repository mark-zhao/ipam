package audit

import (
	"context"
	"errors"
	"fmt"
	modelv1 "ipam/model/v1"
	"ipam/utils/logging"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbIndex = `date`

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
		Options: options.Index().SetUnique(false),
	}})
	if err != nil {
		return nil, err
	}
	return &mongodb{c, sync.RWMutex{}}, nil
}

func (m *mongodb) CreateAudit(ctx context.Context, audit *AuditInfo) (*AuditInfo, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, cErr := m.c.InsertOne(ctx, audit)
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return nil, errors.New("数据录入数据库失败")
	}

	return audit, nil
}

func (m *mongodb) ReadAllAudit(ctx context.Context, st, et time.Time) (audits Audits, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	query := bson.M{
		"date": bson.M{
			"$gte": st.Format("2006-01-02 15:04:05"), // 指定要过滤的日期
			"$lte": et.Format("2006-01-02 15:04:05"), // 指定要过滤的日期
		},
	}
	cur, err := m.c.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(`error reading all AUDITS: %w`, err)
	}
	defer cur.Close(ctx)
	if err = cur.All(ctx, &audits); err != nil {
		logging.Error("获取数据失败")
		return
	}
	return
}
