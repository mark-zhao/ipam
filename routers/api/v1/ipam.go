package v1

import (
	"context"
	"errors"
	"fmt"
	"ipam/utils/cmd"
	"ipam/utils/logging"
	conf "ipam/utils/options"
	"ipam/utils/tools"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ipam/pkg/idc"
	goipam "ipam/pkg/ipam"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const modelIPAM string = "IPAM"

// 注册路由
func IPAMRouter() {
	APIs["/ipam"] = map[UriInterface]interface{}{
		NewUri("GET", "/CidrsList"):                    (&InstanceResource{}).CidrsList,
		NewUri("GET", "/CidrsInfo"):                    (&InstanceResource{}).CidrsInfo,
		NewUri("POST", "/GetPrefix"):                   (&InstanceResource{}).GetPrefix,
		NewUri("POST", "/CreatePrefix"):                (&InstanceResource{}).CreatePrefix,
		NewUri("POST", "/AcquireIP"):                   (&InstanceResource{}).AcquireIP,
		NewUri("POST", "/ReleaseIP"):                   (&InstanceResource{}).ReleaseIP,
		NewUri("POST", "/MarkIP"):                      (&InstanceResource{}).MarkIP,
		NewUri("POST", "/EditIPUserFromPrefix"):        (&InstanceResource{}).EditIPUserFromPrefix,
		NewUri("POST", "/EditIPDescriptionFromPrefix"): (&InstanceResource{}).EditIPDescriptionFromPrefix,
		NewUri("POST", "/DeletePrefix"):                (&InstanceResource{}).DeletePrefix,
		NewUri("POST", "/GetIP"):                       (&InstanceResource{}).GetIP,
	}
}

var ipam goipam.Ipamer

// const (
// 	timeFormart = "2006-01-02"
// )

type InstanceResource struct {
}

// 获取cidrs列表
type CidrsListRes struct {
	Cidrs map[string]map[string][]string `json:"cidrs"`
}

// 获取cidrs信息
type CidrsInfoRes struct {
	Cidrs []CidrInfo `json:"cidrs"`
}

type CidrInfo struct {
	Cidr         string `json:"cidr"`
	Gateway      string `json:"gateway"`
	ParentCidr   string `json:"parentcidr"`
	VlanID       int    `json:"vlanid"`
	VRF          string `json:"vrf"` //VRF
	IDC          string `json:"idc"` //IDC
	IsParent     bool   `json:"isparent"`
	AvailableIPs string `json:"availableips"`
}

// 创建prefix
type CreatePrefixReq struct {
	Cidr    string `json:"cidr"`
	Gateway string `json:"gateway"`
	VlanID  int    `json:"vlanid"`
	VRF     string `json:"vrf"` //VRF
	IDC     string `json:"idc"` //IDC
}
type CreatePrefixRes struct {
	OK int `json:"ok"`
}

// 申请ip
type AcquireIPReq struct {
	Cidr        string `json:"cidr"`
	Description string `json:"description"`
	Num         int    `json:"num"`
	User        string `json:"user"`
}

type AcquireIPRes struct {
	Prefix goipam.Prefix `json:"prefix"`
	Ips    []string      `json:"ips"`
}

// 获取prefix详细信息
type GetPrefixReq struct {
	Cidr string `json:"cidr"`
}

type GetPrefixRes struct {
	Prefix goipam.Prefix `json:"prefix"`
}

// 释放ip
type ReleaseIPReq struct {
	Cidr   string   `json:"cidr"`
	IPList []string `json:"iplist"`
}

// 删除Prefix
type DeletePrefixReq struct {
	Cidr string `json:"cidr"`
}

type DeletePrefixRes struct {
	OK int `json:"ok"`
}

// 获取ip 请求信息
type GetIPReq struct {
	User string `json:"user"`
}

// 获取ip回复数据
type GetIPRes struct {
	IPList []IPInfo `json:"iplist"`
}

type IPInfo struct {
	Cidr     string          `json:"cidr"`
	IP       string          `json:"ip"`
	IPDetail goipam.IPDetail `json:"ipdetail"`
	IDC      string          `json:"idc"`    //机房
	VRF      string          `json:"vrf"`    //VRF
	VlanID   int             `json:"vlanid"` //Vlan 号
}

// MarkIP请求信息
type MarkIPReq struct {
	Cidr        string   `json:"cidr"`
	Ips         []string `json:"ips"`
	User        string   `json:"user"`
	Description string   `json:"description"`
}

// 根据使用人获取ip
func (*InstanceResource) GetIP(c *gin.Context) {
	method := "GetIP"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req GetIPReq
	err := c.ShouldBind(&req)
	if err == nil {
		logging.Debug(req)
		if req.User == "" {
			resp.Render(c, 200, nil, errors.New("参数不能为空"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cidrs, err := ipam.ReadAllPrefixCidrs(ctx)
		if err != nil {
			logging.Error("获取cidrs 失败", err)
			resp.Render(c, 200, nil, errors.New("获取cidrs 失败"))
			return
		}
		var items []IPInfo
		for _, v := range cidrs {
			p := ipam.PrefixFrom(ctx, v)
			for k, v := range p.Ips {
				if v.User == req.User {
					items = append(items, IPInfo{
						p.Cidr,
						k,
						v,
						p.IDC,
						p.VRF,
						p.VlanID,
					})
				}
			}
		}
		resp.Render(c, 200, GetIPRes{items}, nil)
		return
	}
	resp.Render(c, 200, nil, err)
	return
}

// MarkIP
func (*InstanceResource) MarkIP(c *gin.Context) {
	method := "MarkIP"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIPAM, method)
	if !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req MarkIPReq
	err := c.ShouldBind(&req)
	if err == nil {
		if req.Cidr == "" || req.Description == "" || req.User == "" || req.Ips == nil {
			resp.Render(c, 200, nil, errors.New("参数不能为空"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res, err := ipam.MarkIP(ctx, req.Cidr, goipam.IPDetail{Operator: username, User: req.User, Description: req.Description, Date: tools.DateToString()}, req.Ips)
		logging.Error(err)
		resp.Render(c, 200, res, err)
		return
	}
	resp.Render(c, 200, nil, err)
	return
}

// 所有获取网段
func (*InstanceResource) CidrsList(c *gin.Context) {
	method := "CidrsList"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cidrs, err := ipam.ReadAllPrefixCidrs(ctx)
	m := make(map[string]map[string][]string)
	idcs := idc.IDCINFO
	for _, i := range idcs {
		v_n := map[string][]string{}
		for _, v := range i.VRF {
			v_n[v] = []string{}
		}
		m[i.IDCName] = v_n
	}
	for _, v := range cidrs {
		prefix := ipam.PrefixFrom(ctx, v)
		cs := m[prefix.IDC][prefix.VRF]
		m[prefix.IDC][prefix.VRF] = append(cs, v)
	}
	if err != nil {
		logging.Error("获取cidrs 失败", err)
		resp.Render(c, 200, nil, errors.New("获取cidrs 失败"))
		return
	}

	resp.Render(c, 200, CidrsListRes{m}, nil)
	return
}

// 获取网段信息
func (*InstanceResource) CidrsInfo(c *gin.Context) {
	method := "PrefixList"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cidrs, err := ipam.ReadAllPrefixCidrs(ctx)
	if err != nil {
		logging.Error("获取cidrs 失败", err)
		resp.Render(c, 200, nil, errors.New("获取cidrs 失败"))
		return
	}
	var items []CidrInfo
	for _, v := range cidrs {
		p := ipam.PrefixFrom(ctx, v)
		items = append(items, CidrInfo{
			p.Cidr,
			p.Gateway,
			p.ParentCidr,
			p.VlanID,
			p.VRF,
			p.IDC,
			p.IsParent,
			strconv.FormatUint(p.Usage().AcquiredIPs, 10),
		})
	}
	resp.Render(c, 200, CidrsInfoRes{items}, nil)
	return
}

// 获取网段详细信息
func (*InstanceResource) GetPrefix(c *gin.Context) {
	method := "GetPrefix"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req GetPrefixReq
	if c.ShouldBind(&req) == nil {
		if req.Cidr == "" {
			resp.Render(c, 200, nil, errors.New("网段不能为空"))
			return
		}
		logging.Debug(req)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p := ipam.PrefixFrom(ctx, req.Cidr)
		if p != nil {
			arp(req.Cidr, p.IDC, p.VlanID)
			a := ipam.PrefixFrom(ctx, req.Cidr)
			if a != nil {
				logging.Debug(*a)
				resp.Render(c, 200, GetPrefixRes{*a}, nil)
				return
			}
		}

	}
	resp.Render(c, 200, nil, errors.New("解析参数出错"))
	return
}

// 创建网段
func (*InstanceResource) CreatePrefix(c *gin.Context) {
	method := "CreatePrefix"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req CreatePrefixReq
	if c.ShouldBind(&req) == nil {
		logging.Debug(req)
		if req.Cidr == "" || req.IDC == "" || req.Gateway == "" || req.VlanID == 0 || req.VRF == "" {
			resp.Render(c, 200, nil, errors.New("参数不能为空"))
			return
		}
		ipnet, err := netip.ParsePrefix(req.Cidr)
		if err != nil {
			resp.Render(c, 200, nil, err)
			return
		}
		IP, err := netip.ParseAddr(req.Gateway)
		if err != nil || !ipnet.Contains(IP) {
			resp.Render(c, 200, nil, errors.New("gateway 输入有错误"))
			return
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := ipam.NewPrefix(ctx, req.Cidr, req.Gateway, "", req.VlanID, req.VRF, req.IDC, false)
			if err != nil {
				resp.Render(c, 200, nil, err)
				return
			}
		}
		resp.Render(c, 200, CreatePrefixRes{1}, nil)
		return
	}
}

// 申请ip
func (*InstanceResource) AcquireIP(c *gin.Context) {
	method := "AcquireIP"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIPAM, method)
	if !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req AcquireIPReq
	if c.Bind(&req) == nil {
		logging.Debug(req)
		if req.Description == "" || req.User == "" {
			logging.Error("用户或描述不能为空")
			resp.Render(c, 200, nil, errors.New("用户或描述不能为空"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p := ipam.PrefixFrom(ctx, req.Cidr)
		if p != nil {
			arp(req.Cidr, p.IDC, p.VlanID)
			ips, err := ipam.AcquireIP(ctx, req.Cidr, goipam.IPDetail{Operator: username, User: req.User, Description: req.Description, Date: tools.DateToString()}, req.Num)
			if err != nil {
				logging.Error(err)
				resp.Render(c, 200, nil, err)
				return
			}
			resp.Render(c, 200, AcquireIPRes{*p, ips}, nil)
			return
		}
		resp.Render(c, 200, nil, errors.New("网段不存在"))
		return
	}
}

// 释放ip
func (*InstanceResource) ReleaseIP(c *gin.Context) {
	method := "ReleaseIP"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req ReleaseIPReq

	if c.Bind(&req) == nil {
		logging.Debug(req)
		if req.Cidr == "" || req.IPList == nil {
			logging.Error("网段或ips不能为空")
			resp.Render(c, 200, nil, errors.New("网段或ips不能为空"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if res, err := ipam.ReleaseIPFromPrefix(ctx, req.Cidr, req.IPList); err != nil {
			logging.Error(err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			resp.Render(c, 200, res, nil)
			return
		}

	}
	resp.Render(c, 20, nil, fmt.Errorf("程序内部错误"))
	return
}

// 修改用户ip请求数据
type EditIPReq struct {
	User   string              `json:"user"`
	IPList map[string][]string `json:"iplist"`
}

// 修改描述ip请求数据
type EditDescriptionReq struct {
	Cidr        string `json:"cidr"`
	Description string `json:"description"`
	IP          string `json:"ip"`
}

// 修改ip用户属性
func (*InstanceResource) EditIPUserFromPrefix(c *gin.Context) {
	method := "EditIPUserFromPrefix"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req EditIPReq
	if c.Bind(&req) == nil {
		logging.Debug(req)
		if req.User == "" || req.IPList == nil {
			logging.Error("使用人不能为空和 iplist 都不能为空")
			resp.Render(c, 200, nil, errors.New("使用人不能为空和 iplist 都不能为空"))
			return
		}
		var err error
		for k, v := range req.IPList {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err = ipam.EditIPUserFromPrefix(ctx, k, req.User, v); err != nil {
				logging.Debug(err)
			}
		}
		if err != nil {
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, CreatePrefixRes{1}, nil)
	return
}

// 修改ip描述属性
func (*InstanceResource) EditIPDescriptionFromPrefix(c *gin.Context) {
	method := "EditIPDescriptionFromPrefix"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req EditDescriptionReq
	if c.Bind(&req) == nil {
		logging.Debug(req)
		if req.Description == "" || req.IP == "" {
			logging.Error("描述不能为空和iplist都不能为空")
			resp.Render(c, 200, nil, errors.New("描述不能为空和iplist都不能为空"))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := ipam.EditIPDescriptionFromPrefix(ctx, req.Cidr, req.Description, req.IP); err != nil {
			logging.Debug(err)
			resp.Render(c, 200, nil, err)
			return
		}
	}
	resp.Render(c, 200, CreatePrefixRes{1}, nil)
	return
}

// 删除网段
func (*InstanceResource) DeletePrefix(c *gin.Context) {
	method := "DeletePrefix"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req DeletePrefixReq
	if c.Bind(&req) == nil {
		logging.Debug(req)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := ipam.DeletePrefix(ctx, req.Cidr)
		if err != nil {
			logging.Error(err)
			resp.Render(c, 200, CreatePrefixRes{0}, errors.New("删除网段失败"))
			return
		}
	}
	resp.Render(c, 200, CreatePrefixRes{1}, nil)
	return
}

// mongo存储初始化
func init() {
	ctx := context.Background()
	opts := options.Client()
	opts.ApplyURI(fmt.Sprintf(`mongodb://%s:%s`, "192.168.152.92", "27017"))
	opts.Auth = &options.Credential{
		AuthMechanism: `SCRAM-SHA-1`,
		Username:      `ipam`,
		Password:      `123456`,
	}

	c := goipam.MongoConfig{
		DatabaseName:       `ipam`,
		CollectionName:     `prefixes`,
		MongoClientOptions: opts,
	}
	Storage, err := goipam.NewMongo(ctx, c)
	if err != nil {
		logging.Error("数据库连接失败")
	}
	ipam = goipam.NewWithStorage(Storage)
}

func arp(cidr string, idcname string, vlanid int) {
	if !conf.Conf.Arp.Onoff {
		return
	}
	cmd.PingNetwork(cidr)
	var ips []string
	if idc.IDCINFO != nil {
		for _, idc := range idc.IDCINFO {
			if idc.IDCName == idcname {
				if idc.Router != nil {
					for _, v := range idc.Router {
						if len(v.IP) == 0 || len(v.Password) == 0 || len(v.UserName) == 0 || len(v.RUNARPCmd) == 0 {
							continue
						}
						runarpcmd := fmt.Sprintf(v.RUNARPCmd, strconv.Itoa(vlanid))
						command := fmt.Sprintf("sshpass -p %s ssh %s@%s %s", v.Password, v.UserName, v.IP, runarpcmd)
						// command := fmt.Sprintf("sshpass -p '123456' ssh read-only@192.168.47.1 'dis arp vlan %s'", strconv.Itoa(vlanid))
						// command := fmt.Sprintf("sshpass -p '123456' ssh readonly@192.168.169.1 'dis arp vlan %s'", strconv.Itoa(vlanid))
						output, err := cmd.RunShell(command)
						if err != nil {
							logging.Error(err)
							continue
						}
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						zp := regexp.MustCompile(`\s+`)
						lines := strings.Split(output, "\n")
						prefix := ipam.PrefixFrom(ctx, cidr)
						for _, line := range lines {
							ll := len(line)
							if ll > 30 {
								lineData := zp.Split(line, -1)
								if _, err := netip.ParseAddr(lineData[0]); err == nil {
									_, ok := prefix.Ips[lineData[0]]
									if ok {
										continue
									} else {
										ips = append(ips, lineData[0])
									}
								}
							}
						}
					}
				}
				break
			}
		}
	}
	if ips == nil {
		return
	}
	ips = tools.RemoveDuplicateString(ips)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := ipam.MarkIP(ctx, cidr, goipam.IPDetail{Operator: "networkMan", User: "arp", Description: "arp scan", Date: tools.DateToString()}, ips); err != nil {
		logging.Debug("arp ", "标记失败", err)
		return
	}
}
