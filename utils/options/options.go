package options

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var Conf *Config
var configFile string

func InitConfig() *Config {
	flag.StringVar(&configFile, "conf", "./config.json", "define config file ")
	flag.Parse()
	Conf = Getconfig()
	return Conf
}

// User 用户类
type User struct {
	Id         string   `json:"userId"`
	Name       string   `json:"userName"`
	Gender     string   `json:"gender"`
	Phone      string   `json:"userMobile"`
	Pwd        string   `json:"pwd"`
	Permission []string `json:"permission"`
}

// 配置文件结构
type Http struct {
	Addr    string `json:"Addr"`
	RunMode string `json:"Run_mode"`
	//LogDir			string			`json:"log_Dir"`
	//ReadTimeout 	time.Duration	`json:"read_timeout"`
	//WriteTimeout 	time.Duration	`json:"write_timeout"`
}

type Log struct {
	LogLevel   string `json:"LogLevel"`
	LogDir     string `json:"LogDir"`
	LogFile    string `json:"LogFile"`
	LogFileExt string `json:"LogFileExt"`
	TimeFormat string `json:"TimeFormat"`
}

type Arp struct {
	Onoff bool `json:"onoff"`
}

type Config struct {
	Http     Http            `json:"http"`
	Log      Log             `json:"log"`
	UserList map[string]User `json:"userList"`
	Arp      Arp             `json:"arp"`
}

// json读取
type JsonStruct struct {
}

func Getconfig() *Config {
	JsonParse := NewJsonStruct()
	v := new(Config)
	JsonParse.Load(configFile, &v)
	return v
}

// 读json 文件
func (jst *JsonStruct) Load(filename string, v interface{}) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("error:", err)
		return
	}
	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, v)
	if err != nil {
		log.Fatal("error:", err)
		return
	}
}
func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
