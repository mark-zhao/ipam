package note

import (
	"context"
	"errors"
	"fmt"
	goipam "ipam/pkg/ipam"
	"ipam/utils/logging"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongo 连接
type mongodb struct {
	c    *mongo.Collection
	lock sync.RWMutex
}

var cli *mongodb

// 机房信息
type Note struct {
	Instance    string `bson:"instance" json:"instance"`       //实例
	Description string `bson:"description" json:"description"` //描述
	Operator    string `bson:"operator" json:"operator"`       //操作员
	Date        string `bson:"date" json:"date"`               //分配时间
}

func (n *Note) CreateNote() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"instance": n.Instance}
	var nOld Note
	cli.lock.Lock()
	cli.c.FindOne(ctx, filter).Decode(&nOld)
	cli.lock.Unlock()
	if nOld.Instance == n.Instance {
		return errors.New("实例重名")
	}
	cli.lock.Lock()
	_, cErr := cli.c.InsertOne(ctx, &n)
	cli.lock.Unlock()
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return errors.New("数据录入数据库失败")
	}
	return nil
}

func (n *Note) DeleteNote() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"instance": n.Instance}
	cli.lock.Lock()
	_, cErr := cli.c.DeleteOne(ctx, filter)
	cli.lock.Unlock()
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return errors.New("数据录入数据库失败")
	}
	return nil
}

func (n *Note) NoteList() ([]Note, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cur, cErr := cli.c.Find(ctx, bson.D{})
	if cErr != nil {
		logging.Error("获取数据失败", cErr)
		return nil, cErr
	}
	defer cur.Close(ctx)
	var notes []Note
	if err := cur.All(ctx, &notes); err != nil {
		logging.Error("获取数据失败", err)
		return nil, err
	}
	return notes, nil
}
func init() {
	ctx := context.Background()
	opts := options.Client()
	opts.ApplyURI(fmt.Sprintf(`mongodb://%s:%s`, "192.168.152.92", "27017"))
	opts.Auth = &options.Credential{
		AuthMechanism: `SCRAM-SHA-1`,
		Username:      `ipam`,
		Password:      `123456`,
	}

	conf := goipam.MongoConfig{
		DatabaseName:       `ipam`,
		CollectionName:     `note`,
		MongoClientOptions: opts,
	}
	cli, _ = newMongo(ctx, conf)
}

func newMongo(ctx context.Context, config goipam.MongoConfig) (*mongodb, error) {
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
	if err != nil {
		logging.Info("连接数据库失败")
	}
	return &mongodb{c, sync.RWMutex{}}, nil
}
