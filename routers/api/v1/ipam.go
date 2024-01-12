package v1

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"ipam/utils/aeser"
	"ipam/utils/cmd"
	"ipam/utils/logging"
	conf "ipam/utils/options"
	"ipam/utils/tools"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ipam/pkg/audit"
	idc "ipam/pkg/dcim"
	goipam "ipam/pkg/ipam"
	Administrator "ipam/pkg/user"

	"github.com/gin-gonic/gin"
)

const modelIPAM string = "IPAM"

// 注册路由
func IPAMRouter() {
	p := Administrator.Permission{
		Id:    3,
		Label: modelIPAM,
		Children: []Administrator.Permission2{
			{Id: 31, Label: "CidrsList"},
			{Id: 32, Label: "CidrsInfo"},
			{Id: 33, Label: "Cidrs"},
			{Id: 34, Label: "GetPrefix"},
			{Id: 35, Label: "CreatePrefix"},
			{Id: 36, Label: "AcquireIP"},
			{Id: 37, Label: "ReleaseIP"},
			{Id: 38, Label: "MarkIP"},
			{Id: 39, Label: "EditIPUserFromPrefix"},
			{Id: 310, Label: "EditIPDescriptionFromPrefix"},
			{Id: 311, Label: "DeletePrefix"},
			{Id: 312, Label: "GetIP"},
		},
	}
	Permissions = append(Permissions, p)
	APIs["/ipam"] = map[UriInterface]interface{}{
		NewUri("GET", "/CidrsList"):                    (&InstanceResource{}).CidrsList,
		NewUri("GET", "/CidrsInfo"):                    (&InstanceResource{}).CidrsInfo,
		NewUri("GET", "/Cidrs"):                        (&InstanceResource{}).Cidrs,
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
	Project     string `json:"project"` //项目
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
	Project     string   `json:"project"` //项目
	Description string   `json:"description"`
}

// 根据使用人获取ip
func (*InstanceResource) GetIP(c *gin.Context) {
	const method = "GetIP"
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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
}

// MarkIP
func (*InstanceResource) MarkIP(c *gin.Context) {
	const method = "MarkIP"
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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		res, err := ipam.MarkIP(ctx, req.Cidr, goipam.IPDetail{Operator: username, User: req.User, Project: req.Project, Description: req.Description, Date: tools.DateToString()}, req.Ips)
		if err != nil {
			logging.Error("IP标记失败:", err)
			resp.Render(c, 200, nil, err)
			return
		}
		a := &audit.AuditInfo{
			Operator:    username,
			Func:        method,
			Description: strings.Join(req.Ips, ","),
			Date:        tools.DateToString(),
		}
		if err := auditer.Add(ctx, a); err != nil {
			logging.Error("audit insert mongo error:", err)
		}
		resp.Render(c, 200, res, err)
		return
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// 所有获取网段
func (*InstanceResource) CidrsList(c *gin.Context) {
	const method = "CidrsList"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cidrs, err := ipam.ReadAllPrefixCidrs(ctx)
	if err != nil {
		logging.Error("获取cidrs 失败:", err)
		resp.Render(c, 200, nil, errors.New("获取cidrs 失败"))
		return
	}
	m := make(map[string]map[string][]string)
	idcs := idc.IDCs
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
	resp.Render(c, 200, CidrsListRes{m}, nil)
}

type CidrsRes struct {
	Cidrs []goipam.Prefix `json:"cidrs"`
}

// 获取所有网段信息
func (*InstanceResource) Cidrs(c *gin.Context) {
	const method = "Cidrs"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cidrs, err := ipam.ReadAllPrefixCidrs(ctx)
	if err != nil {
		logging.Error("获取cidrs 失败", err)
		resp.Render(c, 200, nil, errors.New("获取cidrs 失败"))
		return
	}
	var items []goipam.Prefix
	for _, v := range cidrs {
		p := ipam.PrefixFrom(ctx, v)
		items = append(items, *p)
	}
	resp.Render(c, 200, CidrsRes{items}, nil)
}

// 获取网段信息
func (*InstanceResource) CidrsInfo(c *gin.Context) {
	const method = "CidrsInfo"
	logging.Info("开始", method)
	if _, ok := tools.FunAuth(c, modelIPAM, method); !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
}

// 获取网段详细信息
func (*InstanceResource) GetPrefix(c *gin.Context) {
	const method = "GetPrefix"
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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		p := ipam.PrefixFrom(ctx, req.Cidr)
		if p != nil {
			if err := arp(req.Cidr, p.IDC, p.VlanID, p.VRF); err != nil {
				logging.Info(err)
				resp.Render(c, 200, nil, fmt.Errorf("arp scan 错误"))
				return
			}
			a := ipam.PrefixFrom(ctx, req.Cidr)
			if a != nil {
				logging.Debug(*a)
				resp.Render(c, 200, GetPrefixRes{*a}, nil)
				return
			}
		}

	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// 创建网段
func (*InstanceResource) CreatePrefix(c *gin.Context) {
	const method = "CreatePrefix"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIPAM, method)
	if !ok {
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
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			_, err := ipam.NewPrefix(ctx, req.Cidr, req.Gateway, "", req.VlanID, req.VRF, req.IDC, false)
			if err != nil {
				logging.Error("创建网段失败:", err)
				resp.Render(c, 200, nil, err)
				return
			}
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: req.Cidr,
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
			resp.Render(c, 200, Res{0}, err)
			return
		}
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// 申请ip
func (*InstanceResource) AcquireIP(c *gin.Context) {
	const method = "AcquireIP"
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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		p := ipam.PrefixFrom(ctx, req.Cidr)
		if p != nil {
			if err := arp(req.Cidr, p.IDC, p.VlanID, p.VRF); err != nil {
				logging.Debug(err)
			}
			ips, err := ipam.AcquireIP(ctx, req.Cidr, goipam.IPDetail{Operator: username, User: req.User, Project: req.Project, Description: req.Description, Date: tools.DateToString()}, req.Num)
			if err != nil {
				logging.Error("申请IP失败:", err)
				resp.Render(c, 200, nil, err)
				return
			}
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: strings.Join(ips, ","),
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
			resp.Render(c, 200, AcquireIPRes{*p, ips}, nil)
			return
		}
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// 释放ip
func (*InstanceResource) ReleaseIP(c *gin.Context) {
	const method = "ReleaseIP"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIPAM, method)
	if !ok {
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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if res, err := ipam.ReleaseIPFromPrefix(ctx, req.Cidr, req.IPList); err != nil {
			logging.Error("释放IP失败:", err)
			resp.Render(c, 200, nil, err)
			return
		} else {
			a := &audit.AuditInfo{
				Operator:    username,
				Func:        method,
				Description: strings.Join(req.IPList, ","),
				Date:        tools.DateToString(),
			}
			if err := auditer.Add(ctx, a); err != nil {
				logging.Error("audit insert mongo error:", err)
			}
			resp.Render(c, 200, res, nil)
			return
		}
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
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
	const method = "EditIPUserFromPrefix"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIPAM, method)
	if !ok {
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
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if err = ipam.EditIPUserFromPrefix(ctx, k, req.User, v); err != nil {
				logging.Debug(err)
			}
		}
		if err != nil {
			logging.Error("修改ip用户属性失败:", err)
			resp.Render(c, 200, nil, err)
			return
		}
		a := &audit.AuditInfo{
			Operator:    username,
			Func:        method,
			Description: req.User,
			Date:        tools.DateToString(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := auditer.Add(ctx, a); err != nil {
			logging.Error("audit insert mongo error:", err)
		}
		resp.Render(c, 200, Res{0}, err)
		return
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// 修改ip描述属性
func (*InstanceResource) EditIPDescriptionFromPrefix(c *gin.Context) {
	const method = "EditIPDescriptionFromPrefix"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIPAM, method)
	if !ok {
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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := ipam.EditIPDescriptionFromPrefix(ctx, req.Cidr, req.Description, req.IP); err != nil {
			logging.Error("修改ip描述属性:", err)
			resp.Render(c, 200, nil, err)
			return
		}
		a := &audit.AuditInfo{
			Operator:    username,
			Func:        method,
			Description: req.IP,
			Date:        tools.DateToString(),
		}
		if err := auditer.Add(ctx, a); err != nil {
			logging.Error("audit insert mongo error:", err)
		}
		resp.Render(c, 200, Res{0}, nil)
		return
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// 删除网段
func (*InstanceResource) DeletePrefix(c *gin.Context) {
	const method = "DeletePrefix"
	logging.Info("开始", method)
	username, ok := tools.FunAuth(c, modelIPAM, method)
	if !ok {
		resp.Render(c, 403, nil, errors.New("没有权限访问"))
		return
	}
	var req DeletePrefixReq
	if c.Bind(&req) == nil {
		logging.Debug(req)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, err := ipam.DeletePrefix(ctx, req.Cidr, false)
		if err != nil {
			logging.Error("删除网段:", err)
			resp.Render(c, 200, CreatePrefixRes{0}, errors.New("删除网段失败"))
			return
		}
		a := &audit.AuditInfo{
			Operator:    username,
			Func:        method,
			Description: req.Cidr,
			Date:        tools.DateToString(),
		}
		if err := auditer.Add(ctx, a); err != nil {
			logging.Error("audit insert mongo error:", err)
		}
		resp.Render(c, 200, Res{0}, nil)
		return
	}
	resp.Render(c, 200, Res{0}, errors.New("解析数据失败"))
}

// arp scan
func arp(cidr string, idcname string, vlanid int, vrf string) error {
	if !conf.Conf.Arp.Onoff {
		return nil
	}
	cmd.PingNetwork(cidr)
	var ips []string
	if idc.IDCs != nil {
		for _, i := range idc.IDCs {
			if i.IDCName == idcname && i.Router != nil {
				for _, v := range i.Router {
					if len(v.IP) == 0 || len(v.Password) == 0 || len(v.UserName) == 0 {
						continue
					}
					encryptResult, err := hex.DecodeString(v.Password)
					if err != nil {
						return err
					}
					hexKey := "6c1acf9ad6f12ff7e3c5b94df9f9ef329996b6ea7d148afafe76765d42d0a876"
					key, err := hex.DecodeString(hexKey)
					if err != nil {
						return err
					}
					Pwresult, err := aeser.AESDecrypt(encryptResult, key)
					if err != nil {
						return err
					}
					var output string
					if v.Brand == "华为" {
						cmds := []string{
							"screen-length 0 temporary",
							fmt.Sprintf("display arp interface Vlanif %s", strconv.Itoa(vlanid)),
							//    add more commands here...
						}
						output, err = cmd.RunShellHW(v.UserName, string(Pwresult), v.IP, cmds)
						if err != nil {
							logging.Error(err)
							return err
						}
					} else if v.Brand == "思科" {
						runarpcmd := fmt.Sprintf("show ip arp vlan %s  vrf %s", strconv.Itoa(vlanid), vrf)
						command := fmt.Sprintf("/usr/bin/sshpass -p '%s' ssh %s@%s '%s'", Pwresult, v.UserName, v.IP, runarpcmd)
						logging.Debug(command)
						output, err = cmd.RunShell(command)
						if err != nil {
							logging.Error(err)
							return err
						}
					} else if v.Brand == "路由表" {
						cmds := []string{
							"screen-length 0 temporary",
							fmt.Sprintf("display ip routing-table vpn-instance %s ip-prefix %s", vrf, cidr),
							//    add more commands here...
						}
						output, err = cmd.RunShellHW(v.UserName, string(Pwresult), v.IP, cmds)
						if err != nil {
							logging.Error(err)
							return err
						}
					} else {
						runarpcmd := fmt.Sprintf("display arp vlan %s", strconv.Itoa(vlanid))
						command := fmt.Sprintf("/usr/bin/sshpass -p '%s' ssh %s@%s '%s'", Pwresult, v.UserName, v.IP, runarpcmd)
						logging.Debug(command)
						output, err = cmd.RunShell(command)
						if err != nil {
							logging.Error(err)
							return err
						}
					}

					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					zp := regexp.MustCompile(`\s+`)
					lines := strings.Split(output, "\n")
					prefix := ipam.PrefixFrom(ctx, cidr)
					for _, line := range lines {
						line = strings.TrimLeft(line, " ")
						if len(line) > 30 {
							regex := regexp.MustCompile(`(?i)incomplete`)
							if regex.MatchString(strings.ToLower(line)) {
								continue
							}
							lineData := zp.Split(line, -1)
							var ip string
							if strings.HasSuffix(lineData[0], "/32") {
								ip = strings.TrimSuffix(lineData[0], "/32")
								logging.Info(ip)
							} else {
								ip = lineData[0]
							}
							if _, err := netip.ParseAddr(ip); err == nil {
								_, ok := prefix.Ips[ip]
								if !ok {
									ips = append(ips, ip)
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
		return nil
	}
	logging.Info(ips)
	ips = tools.RemoveDuplicateString(ips)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := ipam.MarkIP(ctx, cidr, goipam.IPDetail{Operator: "networkMan", User: "arp", Project: "arp scan", Description: "arp scan", Date: tools.DateToString()}, ips); err != nil {
		return fmt.Errorf("arp 标记失败: %s", err)
	}
	return nil
}
