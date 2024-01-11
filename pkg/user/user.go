package Administrator

import (
	"context"
	"encoding/hex"
	"ipam/utils/aeser"
	"ipam/utils/logging"
)

const hexKey = "6c1acf9ad6f12ff7e3c5b94df9f9ef329996b6ea7d148afafe76765d42d0a876"

type Permission2 struct {
	Id    int    `json:"id"`
	Label string `json:"label"`
}

type Permission struct {
	Id       int           `json:"id"`
	Label    string        `json:"label"`
	Children []Permission2 `json:"children"`
}

type User struct {
	// Id         string   `bson:"userId" json:"userId"`
	Name       string                 `bson:"userName" json:"userName"`
	Gender     string                 `bson:"gender" json:"gender"`
	Phone      int                    `bson:"userMobile" json:"userMobile"`
	Pwd        string                 `bson:"pwd" json:"pwd"`
	Permission map[string]Permission2 `bson:"permission" json:"permission"`
}

func (u User) toUser() User {
	// Legacy support only on reading from database, convert to isParent.
	// TODO remove this in the next release
	return User{
		Name:       u.Name,
		Gender:     u.Gender,
		Phone:      u.Phone,
		Pwd:        u.Pwd,
		Permission: u.Permission,
	}
}
func (u *user) List(ctx context.Context) ([]User, error) {
	return u.storage.List(ctx)
}

func (u *user) Get(ctx context.Context, username string) (User, error) {
	return u.storage.Get(ctx, username)
}

func (u *user) Add(ctx context.Context, a *User) error {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}
	encryptResult, err := aeser.AESEncrypt([]byte(a.Pwd), key)
	if err != nil {
		panic(err)
	}
	Pwresult := hex.EncodeToString(encryptResult)
	a.Pwd = Pwresult
	err = u.storage.CreateUser(ctx, a)
	return err
}

func (u *user) Del(ctx context.Context, username string) error {
	return u.storage.DelUser(ctx, username)
}

func (u *user) Changer(ctx context.Context, user *User) error {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}
	encryptResult, err := aeser.AESEncrypt([]byte(user.Pwd), key)
	if err != nil {
		panic(err)
	}
	Pwresult := hex.EncodeToString(encryptResult)
	user.Pwd = Pwresult
	return u.storage.Changer(ctx, user)
}

type LoginReq struct {
	Name string `json:"name" form:"name"`
	Pwd  string `json:"pwd" form:"pwd"`
}

func (u *user) LoginCheck(ctx context.Context, loginReq LoginReq) (bool, User) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return false, User{}
	}
	encryptResult, err := aeser.AESEncrypt([]byte(loginReq.Pwd), key)
	if err != nil {
		panic(err)
	}
	Pwdresult := hex.EncodeToString(encryptResult)
	user, err := u.storage.Get(ctx, loginReq.Name)
	if err == nil && user.Pwd == Pwdresult {
		logging.Info("登录成功")
		return true, user
	}
	logging.Error("用户名或密码错误")
	return false, User{}
}
