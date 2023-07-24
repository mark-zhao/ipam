package idc

import (
	"context"
	"errors"
	"fmt"
	goipam "ipam/pkg/ipam"
	"ipam/utils/logging"
	"ipam/utils/tools"
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
var IDCINFO []IDC

// 路由器信息
type Router struct {
	IP         string `bson:"ip" json:"ip"`
	UserName   string `bson:"username" json:"username"`
	Password   string `bson:"password" json:"password"`
	RUNARPCmd  string `bson:"runarpcmd" json:"runarpcmd"`
	RUNPINGCmd string `bson:"runpingcmd" json:"runpingcmd"`
}

// 机房信息
type IDC struct {
	IDCName     string   `bson:"idcname" json:"idcname"`
	Description string   `bson:"description" json:"description"`
	Router      []Router `bson:"router" json:"router"`
	VRF         []string `bson:"vrf" json:"vrf"`
}

func GetIDC() (idcs []IDC) {
	return IDCINFO
}

// 获取IDC
func GetIDCINFO() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cur, cErr := cli.c.Find(ctx, bson.D{})
	if cErr != nil {
		logging.Error("获取数据失败")
		return
	}
	defer cur.Close(ctx)
	if err := cur.All(ctx, &IDCINFO); err != nil {
		logging.Error("获取数据失败")
		return
	}
	//遍历IDCINFO切片
	for i := range IDCINFO {
		// 遍历每个IDC的Router切片
		for j := range IDCINFO[i].Router {
			// 修改Password字段的值
			IDCINFO[i].Router[j].Password = ""
		}
	}
	return
}

// 新建机房
func (i *IDC) CreateIDC() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	filter := bson.M{"idcname": i.IDCName}
	var iOld IDC
	cli.lock.Lock()
	cli.c.FindOne(ctx, filter).Decode(&iOld)
	cli.lock.Unlock()
	if iOld.IDCName == i.IDCName {
		return errors.New("机房重名")
	}
	cli.lock.Lock()
	_, cErr := cli.c.InsertOne(ctx, &i)
	cli.lock.Unlock()
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return errors.New("数据录入数据库失败")
	}
	GetIDCINFO()
	return nil
}

// 删除机房
func (i *IDC) DeleteIDC() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	filter := bson.M{"idcname": i.IDCName}
	cli.lock.Lock()
	_, cErr := cli.c.DeleteOne(ctx, filter)
	cli.lock.Unlock()
	if cErr != nil {
		logging.Error("insert mongo error:", cErr)
		return errors.New("数据录入数据库失败")
	}
	GetIDCINFO()
	return nil
}

// 新建VRF
func (i *IDC) CreateVRF() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	filter := bson.M{"idcname": i.IDCName}
	var iOld IDC
	cli.lock.Lock()
	if err := cli.c.FindOne(ctx, filter).Decode(&iOld); err == nil {
		oldvrf := iOld.VRF
		if tools.IsExistItem(i.VRF[0], oldvrf) {
			return errors.New("新建的VRF已经存在")
		}
		newVrf := append(oldvrf, i.VRF...)
		update := bson.M{"$set": bson.M{"vrf": newVrf}}
		_, cErr := cli.c.UpdateOne(ctx, filter, update)
		if cErr != nil {
			logging.Error("insert mongo error:", cErr)
			return errors.New("数据录入数据库失败")
		}
	} else {
		return err
	}
	cli.lock.Unlock()
	GetIDCINFO()
	return nil
}

// 删除VRF
func (i *IDC) DeleteVRF() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	filter := bson.M{"idcname": i.IDCName}
	var iOld IDC
	cli.lock.Lock()
	if err := cli.c.FindOne(ctx, filter).Decode(&iOld); err == nil {
		newVrf := tools.RemoveElement(iOld.VRF, i.VRF[0])
		update := bson.M{"$set": bson.M{"vrf": newVrf}}
		_, cErr := cli.c.UpdateOne(ctx, filter, update)
		if cErr != nil {
			logging.Error("insert mongo error:", cErr)
			return errors.New("数据录入数据库失败")
		}
	} else {
		return err
	}
	cli.lock.Unlock()
	GetIDCINFO()
	return nil
}

// 新建路由器
func (i *IDC) CreateRouter() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	filter := bson.M{"idcname": i.IDCName}
	var iOld IDC
	cli.lock.Lock()
	if err := cli.c.FindOne(ctx, filter).Decode(&iOld); err == nil {
		oldRouter := iOld.Router
		var ips []string
		for _, value := range oldRouter {
			ips = append(ips, value.IP)
		}
		if tools.IsExistItem(i.Router[0].IP, ips) {
			return errors.New("新建的Router已经存在")
		}
		newRouter := append(oldRouter, i.Router...)
		update := bson.M{"$set": bson.M{"router": newRouter}}
		_, cErr := cli.c.UpdateOne(ctx, filter, update)
		if cErr != nil {
			logging.Error("insert mongo error:", cErr)
			return errors.New("数据录入数据库失败")
		}
	} else {
		return err
	}
	cli.lock.Unlock()
	GetIDCINFO()
	return nil
}

// 删除VRF
func (i *IDC) DeleteRouter() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	filter := bson.M{"idcname": i.IDCName}
	var iOld IDC
	cli.lock.Lock()
	if err := cli.c.FindOne(ctx, filter).Decode(&iOld); err == nil {
		var newRouter []Router
		for k, v := range iOld.Router {
			if v.IP == i.Router[0].IP {
				newRouter = append(iOld.Router[:k], iOld.Router[1+k:]...)
			}
		}
		update := bson.M{"$set": bson.M{"router": newRouter}}
		_, cErr := cli.c.UpdateOne(ctx, filter, update)
		if cErr != nil {
			logging.Error("insert mongo error:", cErr)
			return errors.New("数据录入数据库失败")
		}
	} else {
		return err
	}
	cli.lock.Unlock()
	GetIDCINFO()
	return nil
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
		CollectionName:     `idc`,
		MongoClientOptions: opts,
	}
	cli, _ = newMongo(ctx, conf)
	GetIDCINFO()
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
