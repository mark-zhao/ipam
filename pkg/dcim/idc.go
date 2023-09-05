package dcim

import (
	"context"
	"errors"
	"fmt"
	"ipam/utils/logging"
	"ipam/utils/tools"
	"net/netip"
)

// 路由器信息
type Router struct {
	IP        string `bson:"ip" json:"ip"`
	UserName  string `bson:"username" json:"username"`
	Password  string `bson:"password" json:"password"`
	RUNARPCmd string `bson:"runarpcmd" json:"runarpcmd"`
	Brand     string `bson:"brand" json:"brand"`
}

// 机房信息
type IDC struct {
	IDCName     string   `bson:"idcname" json:"idcname"`
	Description string   `bson:"description" json:"description"`
	Router      []Router `bson:"router" json:"router"`
	VRF         []string `bson:"vrf" json:"vrf"`
}

type IDCS []IDC

var IDCs IDCS

func (I IDC) deepCopy() *IDC {
	return &IDC{
		IDCName:     I.IDCName,
		Description: I.Description,
		Router:      I.Router,
		VRF:         I.VRF,
	}
}

// NEWIDC create a new IDC from a string notation.
func (d *dcim) NewIDC(idcname, description string, Router []Router, vrf []string) (*IDC, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	I := &IDC{
		idcname,
		description,
		Router,
		vrf,
	}
	return I, nil
}

func GetIDCs() (idcs []IDC) {
	idcs = IDCs
	//遍历IDCINFO切片
	for i := range idcs {
		// 遍历每个IDC的Router切片
		for j := range idcs[i].Router {
			// 修改Password字段的值
			idcs[i].Router[j].Password = ""
		}
	}
	return idcs
}

func (d *dcim) GetIDCINFO(ctx context.Context) {
	if idcs, err := d.storage.ReadAllIDC(ctx); err == nil {
		IDCs = idcs
	} else {
		logging.Debug(err)
	}
}

// 新建机房
func (d *dcim) CreateIDC(ctx context.Context, i IDC) error {
	if _, err := d.storage.ReadIDC(ctx, i.IDCName); err == nil {
		return errors.New("机房重名")
	}
	_, err := d.storage.CreateIDC(ctx, i)
	if err != nil {
		return err
	}
	d.GetIDCINFO(ctx)
	return nil
}

// 删除机房
func (d *dcim) DeleteIDC(ctx context.Context, idcname string) error {
	err := d.storage.DeleteIDC(ctx, idcname)
	d.GetIDCINFO(ctx)
	return err
}

// 新建VRF
func (d *dcim) CreateVRF(ctx context.Context, i IDC) error {
	oi, err := d.storage.ReadIDC(ctx, i.IDCName)
	if err != nil {
		return err
	}
	if tools.IsExistItem(i.VRF[0], oi.VRF) {
		return errors.New("新建的VRF已经存在")
	}
	oi.VRF = append(oi.VRF, i.VRF[0])
	_, err = d.storage.UpdateIDC(ctx, oi)
	d.GetIDCINFO(ctx)
	return err
}

// 删除VRF
func (d *dcim) DeleteVRF(ctx context.Context, i IDC) error {
	oi, err := d.storage.ReadIDC(ctx, i.IDCName)
	if err != nil {
		return err
	}
	oi.VRF = tools.RemoveElement(oi.VRF, i.VRF[0])
	_, err = d.storage.UpdateIDC(ctx, oi)
	d.GetIDCINFO(ctx)
	return err
}

// 新建路由器
func (d *dcim) CreateRouter(ctx context.Context, i IDC) error {
	oi, err := d.storage.ReadIDC(ctx, i.IDCName)
	if err != nil {
		return err
	}
	if _, err := netip.ParseAddr(i.Router[0].IP); err != nil {
		return fmt.Errorf("路由器ip不是一个合适的ip")
	}
	for _, v := range oi.Router {
		if v.IP == i.Router[0].IP {
			return fmt.Errorf("新建的VRF已经存在")
		}
	}
	oi.Router = append(oi.Router, i.Router[0])
	_, err = d.storage.UpdateIDC(ctx, oi)
	d.GetIDCINFO(ctx)
	return err
}

// 删除路由器
func (d *dcim) DeleteRouter(ctx context.Context, i IDC) error {
	oi, err := d.storage.ReadIDC(ctx, i.IDCName)
	if err != nil {
		return err
	}
	for a, v := range oi.Router {
		if v.IP == i.Router[0].IP {
			oi.Router = append(oi.Router[:a], oi.Router[a+1:]...)
		}
	}
	_, err = d.storage.UpdateIDC(ctx, oi)
	d.GetIDCINFO(ctx)
	return err
}
