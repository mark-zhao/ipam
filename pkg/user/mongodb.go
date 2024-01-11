package Administrator

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

const dbIndex = `userName`
const userIndex = `userName`

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
func (m *mongodb) List(ctx context.Context) (users []User, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	f := bson.D{} // match all documents
	cur, err := m.c.Find(ctx, f)
	if err != nil {
		return nil, fmt.Errorf(`error reading all idcs: %w`, err)
	}
	defer cur.Close(ctx)
	if err = cur.All(ctx, &users); err != nil {
		logging.Error("获取数据失败")
		return
	}
	return
}

func (m *mongodb) Get(ctx context.Context, username string) (User, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	f := bson.D{{Key: "userName", Value: username}}
	r := m.c.FindOne(ctx, f)

	// ErrNoDocuments should be returned if the prefix does not exist
	if r.Err() != nil && errors.Is(r.Err(), mongo.ErrNoDocuments) {
		return User{}, fmt.Errorf(`user not found:%s, error:%w`, username, r.Err())
	} else if r.Err() != nil {
		return User{}, fmt.Errorf(`error while trying to find user:%s, error:%w`, username, r.Err())
	}
	u := User{}
	err := r.Decode(&u)
	if err != nil {
		return User{}, fmt.Errorf("unable to read user:%w", err)
	}
	return u.toUser(), nil
}

func (m *mongodb) CreateUser(ctx context.Context, userinfo *User) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	filter := bson.M{"userName": userinfo.Name}
	r := m.c.FindOne(ctx, filter)
	if r.Err() == nil {
		return errors.New("用户已存在")
	}

	_, cErr := m.c.InsertOne(ctx, userinfo)
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return errors.New("数据录入数据库失败")
	}

	return nil
}

func (m *mongodb) DelUser(ctx context.Context, username string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	f := bson.M{dbIndex: username}
	r := m.c.FindOneAndDelete(ctx, f)
	// ErrNoDocuments should be returned if the idc does not exist
	if r.Err() != nil && errors.Is(r.Err(), mongo.ErrNoDocuments) {
		return fmt.Errorf(`user not found:%s, error:%w`, username, r.Err())
	} else if r.Err() != nil {
		return fmt.Errorf(`error while trying to find user:%s, error:%w`, username, r.Err())
	}
	return nil
}

func (m *mongodb) Changer(ctx context.Context, user *User) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	f := bson.M{userIndex: user.Name}
	o := options.Replace().SetUpsert(false)
	r, err := m.c.ReplaceOne(ctx, f, user, o)
	if err != nil {
		return fmt.Errorf("unable to update user:%s, error: %w", user.Name, err)
	}
	if r.MatchedCount == 0 {
		return fmt.Errorf("unable to update user:%s,error:没有匹配到用户", user.Name)
	}
	if r.ModifiedCount == 0 {
		return fmt.Errorf("update did not effect any document:%s",
			user.Name)
	}

	return nil
}
