package v1

import (
	"context"
	"fmt"
	"ipam/component"
	modelv1 "ipam/model/v1"
	"ipam/pkg/audit"
	"ipam/pkg/dcim"
	goipam "ipam/pkg/ipam"
	"ipam/pkg/note"
	"ipam/utils/logging"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var resp component.GokuApiResponse

const (
	AuthMechanism = `SCRAM-SHA-1`
	Username      = "ipam"
	Password      = "123456"
	DatabaseName  = "ipam"
	MongodbIP     = "192.168.152.92"
	DBPort        = "27017"
)

type UriInterface interface {
	GetModel() string
	GetUri() string
}

type Uri struct {
	model string
	uri   string
}

func NewUri(model, uri string) *Uri {
	if len(uri) != 0 && uri[0:1] != "/" {
		uri = ""
	}
	return &Uri{
		model: model,
		uri:   uri,
	}
}

func (u *Uri) GetModel() string {
	return u.model
}

func (u *Uri) GetUri() string {
	return u.uri
}

var APIs = make(map[string]map[UriInterface]interface{})

// 初始化数据库
func init() {
	ctx := context.Background()
	//opts.ApplyURI("mongodb://172.29.253.89:27001,172.29.253.90:27001/ipam?replicaSet=ipam_repl")
	opts := options.Client().ApplyURI(fmt.Sprintf(`mongodb://%s:%s`, MongodbIP, DBPort))
	// 设置连接池大小
	opts.SetMaxPoolSize(10)

	// 设置最大空闲连接数
	opts.SetMaxConnIdleTime(2)

	opts.Auth = &options.Credential{
		AuthMechanism: AuthMechanism,
		Username:      Username,
		Password:      Password,
	}

	c := modelv1.MongoOpts{
		MongoClientOptions: opts,
	}
	m, err := modelv1.NewMongo(ctx, c)
	if err != nil {
		logging.Error("连接数据库失败:", err)
		os.Exit(0)
	}
	initipam(ctx, m)
	initidc(ctx, m)
	initaudit(ctx, m)
	initnote(ctx, m)
}

// dcimermongo存储初始化
var dcimer dcim.Dcimer

func initidc(ctx context.Context, m *mongo.Client) {
	conf := modelv1.MongoConfig{DatabaseName: DatabaseName, CollectionName: "idc"}
	Storage, err := dcim.NewMongo(ctx, m, conf)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	dcimer = dcim.NewWithStorage(Storage)
	dcimer.GetIDCINFO(ctx)
}

// ipam mongo存储初始化
var ipam goipam.Ipamer

func initipam(ctx context.Context, m *mongo.Client) {
	conf := modelv1.MongoConfig{DatabaseName: DatabaseName, CollectionName: "prefixes"}
	Storage, err := goipam.NewMongo(ctx, m, conf)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	ipam = goipam.NewWithStorage(Storage)
}

// audit mongo存储初始化
var auditer audit.Auditer

func initaudit(ctx context.Context, m *mongo.Client) {
	conf := modelv1.MongoConfig{DatabaseName: DatabaseName, CollectionName: "audit"}
	Storage, err := audit.NewMongo(ctx, m, conf)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	auditer = audit.NewWithStorage(Storage)
}

// noter mongo存储初始化
var noter note.Noter

func initnote(ctx context.Context, m *mongo.Client) {
	conf := modelv1.MongoConfig{DatabaseName: DatabaseName, CollectionName: "note"}
	Storage, err := note.NewMongo(ctx, m, conf)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	noter = note.NewWithStorage(Storage)
}
