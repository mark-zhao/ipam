package v1

import (
	"context"
	"fmt"
	"ipam/component"
	"ipam/pkg/audit"
	"ipam/pkg/dcim"
	goipam "ipam/pkg/ipam"
	"ipam/pkg/note"
	"ipam/utils/logging"

	"go.mongodb.org/mongo-driver/mongo/options"
)

var resp component.GokuApiResponse

const (
	AuthMechanism = `SCRAM-SHA-1`
	Username      = `ipam`
	Password      = `123456`
	DatabaseName  = `ipam`
)

type Response struct {
	msg  string      `json:"msg"`
	data interface{} `json:"data"`
}

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
	initipam()
	initidc()
	initaudit()
	initnote()
}

// dcimermongo存储初始化
var dcimer dcim.Dcimer

func initidc() {
	ctx := context.Background()
	opts := options.Client()
	opts.ApplyURI(fmt.Sprintf(`mongodb://%s:%s`, "192.168.152.92", "27017"))
	opts.Auth = &options.Credential{
		AuthMechanism: AuthMechanism,
		Username:      Username,
		Password:      Password,
	}

	c := dcim.MongoConfig{
		DatabaseName:       DatabaseName,
		CollectionName:     `idc`,
		MongoClientOptions: opts,
	}
	Storage, err := dcim.NewMongo(ctx, c)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	dcimer = dcim.NewWithStorage(Storage)
	dcimer.GetIDCINFO(ctx)
}

// ipam mongo存储初始化
var ipam goipam.Ipamer

func initipam() {
	ctx := context.Background()
	opts := options.Client()
	opts.ApplyURI(fmt.Sprintf(`mongodb://%s:%s`, "192.168.152.92", "27017"))
	opts.Auth = &options.Credential{
		AuthMechanism: AuthMechanism,
		Username:      Username,
		Password:      Password,
	}

	c := goipam.MongoConfig{
		DatabaseName:       DatabaseName,
		CollectionName:     `prefixes`,
		MongoClientOptions: opts,
	}
	Storage, err := goipam.NewMongo(ctx, c)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	ipam = goipam.NewWithStorage(Storage)
}

// audit mongo存储初始化
var auditer audit.Auditer

func initaudit() {
	ctx := context.Background()
	opts := options.Client()
	opts.ApplyURI(fmt.Sprintf(`mongodb://%s:%s`, "192.168.152.92", "27017"))
	opts.Auth = &options.Credential{
		AuthMechanism: AuthMechanism,
		Username:      Username,
		Password:      Password,
	}

	c := audit.MongoConfig{
		DatabaseName:       DatabaseName,
		CollectionName:     `audit`,
		MongoClientOptions: opts,
	}
	Storage, err := audit.NewMongo(ctx, c)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	auditer = audit.NewWithStorage(Storage)
}

// noter mongo存储初始化
var noter note.Noter

func initnote() {
	ctx := context.Background()
	opts := options.Client()
	opts.ApplyURI(fmt.Sprintf(`mongodb://%s:%s`, "192.168.152.92", "27017"))
	opts.Auth = &options.Credential{
		AuthMechanism: AuthMechanism,
		Username:      Username,
		Password:      Password,
	}

	c := note.MongoConfig{
		DatabaseName:       DatabaseName,
		CollectionName:     `note`,
		MongoClientOptions: opts,
	}
	Storage, err := note.NewMongo(ctx, c)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	noter = note.NewWithStorage(Storage)
}
