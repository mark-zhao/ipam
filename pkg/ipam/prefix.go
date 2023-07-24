package ipam

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"ipam/utils/tools"
	"math"
	"net/netip"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"go4.org/netipx"
)

var (
	// ErrNotFound is returned if prefix or cidr was not found
	ErrNotFound NotFoundError
	// ErrNoIPAvailable is returned if no IP is available anymore
	ErrNoIPAvailable NoIPAvailableError
	// ErrAlreadyAllocated is returned if the requested address is not available
	ErrAlreadyAllocated AlreadyAllocatedError
	// ErrOptimisticLockError is returned if insert or update conflicts with the existing data
	ErrOptimisticLockError OptimisticLockError
)

type IPDetail struct {
	Operator    string `json:"operator"`    //操作员
	User        string `json:"user"`        //使用人
	Description string `json:"description"` //描述
	Date        string `json:"date"`        //分配时间
}

// Prefix is a expression of a ip with length and forms a classless network.
type Prefix struct {
	Gateway                string          `json:"gateway"`                //网关
	VlanID                 int             `json:"vlanid"`                 //Vlan 号
	VRF                    string          `json:"vrf"`                    //VRF
	IDC                    string          `json:"idc"`                    //IDC
	Cidr                   string          `json:"cidr"`                   // The Cidr of this prefix
	ParentCidr             string          `json:"parentcidr"`             // if this prefix is a child this is a pointer back
	IsParent               bool            `json:"isparent"`               // if this Prefix has child prefixes, this is set to true
	availableChildPrefixes map[string]bool `json:"availablechildprefixes"` // available child prefixes of this prefix
	// TODO remove this in the next release
	childPrefixLength int                 `json:"childprefixlength"` // the length of the child prefixes
	Ips               map[string]IPDetail `json:"ips"`               // The ips contained in this prefix
	version           int64               `json:"version"`           // version is used for optimistic locking
}

type Prefixes []Prefix

// deepCopy to a new Prefix
func (p Prefix) deepCopy() *Prefix {
	return &Prefix{
		Gateway:                p.Gateway,
		VlanID:                 p.VlanID,
		VRF:                    p.VRF,
		IDC:                    p.IDC,
		Cidr:                   p.Cidr,
		ParentCidr:             p.ParentCidr,
		IsParent:               p.IsParent,
		childPrefixLength:      p.childPrefixLength,
		availableChildPrefixes: copyMap(p.availableChildPrefixes),
		Ips:                    copyStruct(p.Ips),
		version:                p.version,
	}
}

// GobEncode implements GobEncode for Prefix
func (p *Prefix) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(p.availableChildPrefixes)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(p.childPrefixLength)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(p.IsParent)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(p.Ips)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(p.version)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(p.Cidr)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(p.ParentCidr)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// GobDecode implements GobDecode for Prefix
func (p *Prefix) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&p.availableChildPrefixes)
	if err != nil {
		return err
	}
	err = decoder.Decode(&p.childPrefixLength)
	if err != nil {
		return err
	}
	err = decoder.Decode(&p.IsParent)
	if err != nil {
		return err
	}
	err = decoder.Decode(&p.Ips)
	if err != nil {
		return err
	}
	err = decoder.Decode(&p.version)
	if err != nil {
		return err
	}
	err = decoder.Decode(&p.Cidr)
	if err != nil {
		return err
	}
	return decoder.Decode(&p.ParentCidr)
}

func copyMap(m map[string]bool) map[string]bool {
	cm := make(map[string]bool, len(m))
	for k, v := range m {
		cm[k] = v
	}
	return cm
}

func copyStruct(m map[string]IPDetail) map[string]IPDetail {
	cm := make(map[string]IPDetail, len(m))
	for k, v := range m {
		cm[k] = v
	}
	return cm
}

// Usage of ips and child Prefixes of a Prefix
type Usage struct {
	// AvailableIPs the number of available IPs if this is not a parent prefix
	// No more than 2^31 available IPs are reported
	AvailableIPs uint64
	// AcquiredIPs the number of acquired IPs if this is not a parent prefix
	AcquiredIPs uint64
	// AvailableSmallestPrefixes is the count of available Prefixes with 2 countable Bits
	// No more than 2^31 available Prefixes are reported
	AvailableSmallestPrefixes uint64
	// AvailablePrefixes is a list of prefixes which are available
	AvailablePrefixes []string
	// AcquiredPrefixes the number of acquired prefixes if this is a parent prefix
	AcquiredPrefixes uint64
}

func (i *ipamer) NewPrefix(ctx context.Context, cidr, gateway, parentCidr string, vlanId int, vrf, idc string, isParent bool) (*Prefix, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	existingPrefixes, err := i.storage.ReadAllPrefixCidrs(ctx)
	if err != nil {
		return nil, err
	}
	p, err := i.newPrefix(cidr, gateway, vlanId, vrf, idc, parentCidr, isParent)
	if err != nil {
		return nil, err
	}
	err = PrefixesOverlapping(i, existingPrefixes, []string{p.Cidr})
	if err != nil {
		return nil, err
	}
	newPrefix, err := i.storage.CreatePrefix(ctx, *p)
	if err != nil {
		return nil, err
	}

	return &newPrefix, nil
}

func (i *ipamer) DeletePrefix(ctx context.Context, cidr string) (*Prefix, error) {
	p := i.PrefixFrom(ctx, cidr)
	if p == nil {
		return nil, fmt.Errorf("%w: delete prefix:%s", ErrNotFound, cidr)
	}
	// if p.hasIPs() {
	// 	return nil, fmt.Errorf("prefix %s has ips, delete prefix not possible", p.Cidr)
	// }
	prefix, err := i.storage.DeletePrefix(ctx, *p)
	if err != nil {
		return nil, fmt.Errorf("delete prefix:%s %w", cidr, err)
	}

	return &prefix, nil
}

func (i *ipamer) AcquireChildPrefix(ctx context.Context, parentCidr string, length uint8) (*Prefix, error) {
	var prefix *Prefix
	return prefix, retryOnOptimisticLock(func() error {
		var err error
		prefix, err = i.acquireChildPrefixInternal(ctx, parentCidr, "", int(length))
		return err
	})
}

func (i *ipamer) AcquireSpecificChildPrefix(ctx context.Context, parentCidr, childCidr string) (*Prefix, error) {
	var prefix *Prefix
	return prefix, retryOnOptimisticLock(func() error {
		var err error
		prefix, err = i.acquireChildPrefixInternal(ctx, parentCidr, childCidr, 0)
		return err
	})
}

// acquireChildPrefixInternal will return a Prefix with a smaller length from the given Prefix.
func (i *ipamer) acquireChildPrefixInternal(ctx context.Context, parentCidr, childCidr string, length int) (*Prefix, error) {
	specificChildRequest := childCidr != ""
	var childprefix netip.Prefix
	parent := i.PrefixFrom(ctx, parentCidr)
	if parent == nil {
		return nil, fmt.Errorf("unable to find prefix for cidr:%s", parentCidr)
	}
	ipprefix, err := netip.ParsePrefix(parent.Cidr)
	if err != nil {
		return nil, err
	}
	if specificChildRequest {
		childprefix, err = netip.ParsePrefix(childCidr)
		if err != nil {
			return nil, err
		}
		length = childprefix.Bits()
	}
	if ipprefix.Bits() >= length {
		return nil, fmt.Errorf("given length:%d must be greater than prefix length:%d", length, ipprefix.Bits())
	}
	if parent.hasIPs() {
		return nil, fmt.Errorf("prefix %s has ips, acquire child prefix not possible", parent.Cidr)
	}

	var ipsetBuilder netipx.IPSetBuilder
	ipsetBuilder.AddPrefix(ipprefix)
	for cp, available := range parent.availableChildPrefixes {
		if available {
			continue
		}
		cpipprefix, err := netip.ParsePrefix(cp)
		if err != nil {
			return nil, err
		}
		ipsetBuilder.RemovePrefix(cpipprefix)
	}

	ipset, err := ipsetBuilder.IPSet()
	if err != nil {
		return nil, fmt.Errorf("error constructing ipset:%w", err)
	}

	var cp netip.Prefix
	if !specificChildRequest {
		var ok bool
		cp, _, ok = ipset.RemoveFreePrefix(uint8(length))
		if !ok {
			pfxs := ipset.Prefixes()
			if len(pfxs) == 0 {
				return nil, fmt.Errorf("no prefix found in %s with length:%d", parentCidr, length)
			}

			var availablePrefixes []string
			for _, p := range pfxs {
				availablePrefixes = append(availablePrefixes, p.String())
			}
			adj := "are"
			if len(availablePrefixes) == 1 {
				adj = "is"
			}

			return nil, fmt.Errorf("no prefix found in %s with length:%d, but %s %s available", parentCidr, length, strings.Join(availablePrefixes, ","), adj)
		}
	} else {
		if ok := ipset.ContainsPrefix(childprefix); !ok {
			// Parent prefix does not contain specific child prefix
			return nil, fmt.Errorf("specific prefix %s is not available in prefix %s", childCidr, parentCidr)
		}
		cp = childprefix
	}

	child := &Prefix{
		Cidr:       cp.String(),
		ParentCidr: parentCidr,
	}

	parent.availableChildPrefixes[child.Cidr] = false
	parent.IsParent = true

	_, err = i.storage.UpdatePrefix(ctx, *parent)
	if err != nil {
		return nil, fmt.Errorf("unable to update parent prefix:%v error:%w", parent, err)
	}

	//待修改
	//child, err = i.newPrefix(child.Cidr, parentCidr)
	if err != nil {
		return nil, fmt.Errorf("unable to persist created child:%w", err)
	}
	_, err = i.storage.CreatePrefix(ctx, *child)
	if err != nil {
		return nil, fmt.Errorf("unable to update parent prefix:%v error:%w", child, err)
	}

	return child, nil
}

func (i *ipamer) ReleaseChildPrefix(ctx context.Context, child *Prefix) error {
	return retryOnOptimisticLock(func() error {
		return i.releaseChildPrefixInternal(ctx, child)
	})
}

// releaseChildPrefixInternal will mark this child Prefix as available again.
func (i *ipamer) releaseChildPrefixInternal(ctx context.Context, child *Prefix) error {
	parent := i.PrefixFrom(ctx, child.ParentCidr)

	if parent == nil {
		return fmt.Errorf("prefix %s is no child prefix", child.Cidr)
	}
	if len(child.Ips) > 2 {
		return fmt.Errorf("prefix %s has ips, deletion not possible", child.Cidr)
	}

	parent.availableChildPrefixes[child.Cidr] = true
	_, err := i.DeletePrefix(ctx, child.Cidr)
	if err != nil {
		return fmt.Errorf("unable to release prefix %v:%w", child, err)
	}
	_, err = i.storage.UpdatePrefix(ctx, *parent)
	if err != nil {
		return fmt.Errorf("unable to release prefix %v:%w", child, err)
	}
	return nil
}

func (i *ipamer) PrefixFrom(ctx context.Context, cidr string) *Prefix {
	ipprefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil
	}
	prefix, err := i.storage.ReadPrefix(ctx, ipprefix.Masked().String())
	if err != nil {
		return nil
	}
	return &prefix
}

func (i *ipamer) MarkIP(ctx context.Context, prefixCidr string, ipDetail IPDetail, ips []string) (*ReleaseIPRes, error) {
	ipnet, err := netip.ParsePrefix(prefixCidr)
	if err != nil {
		return nil, fmt.Errorf(" %s, This is not a network segment", prefixCidr)
	}
	res := &ReleaseIPRes{}
	var IPS []netip.Addr
	for _, ip := range ips {
		IP, err := netip.ParseAddr(ip)
		if err != nil || !ipnet.Contains(IP) {
			res.Result = append(res.Result, ReleaseResult{ip, "标记IP失败,ErrNotFound IP"})
		} else {
			res.Result = append(res.Result, ReleaseResult{ip, "标记IP成功"})
			IPS = append(IPS, IP)
		}
	}
	prefix := i.PrefixFrom(ctx, prefixCidr)
	if prefix == nil {
		return nil, fmt.Errorf("%w: unable to find prefix for cidr:%s", ErrNotFound, prefixCidr)
	}
	for _, v := range IPS {
		ipstring := v.String()
		prefix.Ips[ipstring] = ipDetail
	}
	_, err = i.storage.UpdatePrefix(ctx, *prefix)
	if err != nil {
		return nil, fmt.Errorf("Failed to write to database")
	}
	return res, nil
}

// 修改ip用户
func (i *ipamer) EditIPUserFromPrefix(ctx context.Context, prefixCidr, user string, ips []string) error {
	prefix := i.PrefixFrom(ctx, prefixCidr)
	if prefix == nil {
		return fmt.Errorf("%w: unable to find prefix for cidr:%s", ErrNotFound, prefixCidr)
	}
	for _, v := range ips {
		if prefix.IsParent {
			return fmt.Errorf("prefix %s has childprefixes, acquire ip not possible", prefix.Cidr)
		}
		_, ok := prefix.Ips[v]
		if ok {
			IPDetail := prefix.Ips[v]
			IPDetail.User = user
			prefix.Ips[v] = IPDetail
		}
	}
	_, err := i.storage.UpdatePrefix(ctx, *prefix)
	if err != nil {
		return fmt.Errorf("unable to EditIPUserFromPrefix ip:%v error:%w", prefix, err)
	}
	return nil
}

func (i *ipamer) EditIPDescriptionFromPrefix(ctx context.Context, prefixCidr, description string, ip string) error {
	prefix := i.PrefixFrom(ctx, prefixCidr)
	if prefix == nil {
		return fmt.Errorf("%w: unable to find prefix for cidr:%s", ErrNotFound, prefixCidr)
	}
	if prefix.IsParent {
		return fmt.Errorf("prefix %s has childprefixes, acquire ip not possible", prefix.Cidr)
	}
	_, ok := prefix.Ips[ip]
	if ok {
		IPDetail := prefix.Ips[ip]
		IPDetail.Description = description
		prefix.Ips[ip] = IPDetail
		_, err := i.storage.UpdatePrefix(ctx, *prefix)
		if err != nil {
			return fmt.Errorf("1unable to EditIPDescriptionFromPrefix ip:%v error:%w", prefix, err)
		}
		return nil
	}
	return fmt.Errorf("2unable to EditIPDescriptionFromPrefix ip:%v", prefix)

}

func (i *ipamer) AcquireSpecificIP(ctx context.Context, prefixCidr string, ipDetail IPDetail, specificIP string, num int) ([]string, error) {
	var ips []string
	return ips, retryOnOptimisticLock(func() error {
		var err error
		ips, err = i.acquireSpecificIPInternal(ctx, prefixCidr, ipDetail, specificIP, num)
		return err
	})
}

// acquireSpecificIPInternal will acquire given IP and mark this IP as used, if already in use, return nil.
// If specificIP is empty, the next free IP is returned.
// If there is no free IP an NoIPAvailableError is returned.
// If the Prefix is not found an NotFoundError is returned.
func (i *ipamer) acquireSpecificIPInternal(ctx context.Context, prefixCidr string, ipDetail IPDetail, specificIP string, num int) ([]string, error) {
	prefix := i.PrefixFrom(ctx, prefixCidr)
	if prefix == nil {
		return nil, fmt.Errorf("%w: unable to find prefix for cidr:%s", ErrNotFound, prefixCidr)
	}
	if prefix.IsParent {
		return nil, fmt.Errorf("prefix %s has childprefixes, acquire ip not possible", prefix.Cidr)
	}
	ipnet, err := netip.ParsePrefix(prefix.Cidr)
	if err != nil {
		return nil, err
	}

	var specificIPnet netip.Addr
	if specificIP != "" {
		specificIPnet, err = netip.ParseAddr(specificIP)
		if err != nil {
			return nil, fmt.Errorf("given ip:%s in not valid", specificIP)
		}
		if !ipnet.Contains(specificIPnet) {
			return nil, fmt.Errorf("given ip:%s is not in %s", specificIP, prefixCidr)
		}
		_, ok := prefix.Ips[specificIPnet.String()]
		if ok {
			return nil, fmt.Errorf("%w: given ip:%s is already allocated", ErrAlreadyAllocated, specificIPnet)
		}
	}

	iprange := netipx.RangeOfPrefix(ipnet)
	anum := 0
	ips := []string{}
	for ip := iprange.From(); ipnet.Contains(ip); ip = ip.Next() {
		ipstring := ip.String()
		_, ok := prefix.Ips[ipstring]
		if ok {
			continue
		}
		if specificIP == "" || specificIPnet.Compare(ip) == 0 {
			prefix.Ips[ipstring] = ipDetail
			ips = append(ips, ipstring)
			anum += 1
			if anum == num {
				break
			}
		}
	}
	if anum < num {
		return nil, fmt.Errorf("%s 当前只能分配出%d", prefixCidr, anum)
	}
	_, err = i.storage.UpdatePrefix(ctx, *prefix)
	if err != nil {
		return nil, fmt.Errorf("unable to persist acquired")
	}
	return ips, nil
}

func (i *ipamer) AcquireIP(ctx context.Context, prefixCidr string, ipDetail IPDetail, num int) (ips []string, err error) {
	ips, err = i.AcquireSpecificIP(ctx, prefixCidr, ipDetail, "", num)
	return
}

func (i *ipamer) ReleaseIP(ctx context.Context, ip *IP) (*Prefix, error) {
	_, err := i.ReleaseIPFromPrefix(ctx, ip.ParentPrefix, []string{ip.IP.String()})
	prefix := i.PrefixFrom(ctx, ip.ParentPrefix)
	return prefix, err
}

func (i *ipamer) ReleaseIPFromPrefix(ctx context.Context, prefixCidr string, ips []string) (ReleaseIPRes *ReleaseIPRes, err error) {
	// return retryOnOptimisticLock(func() error {
	// 	return i.releaseIPFromPrefixInternal(ctx, prefixCidr, ips)
	// })
	err = retryOnOptimisticLock(func() error {
		var innerErr error
		ReleaseIPRes, innerErr = i.releaseIPFromPrefixInternal(ctx, prefixCidr, ips)
		return innerErr
	})

	return ReleaseIPRes, err
}

type ReleaseResult struct {
	IP     string `json:"ip"`
	Result string `json:"result"`
}
type ReleaseIPRes struct {
	Result []ReleaseResult `json:"result"`
}

// releaseIPFromPrefixInternal will release the given IP for later usage.
func (i *ipamer) releaseIPFromPrefixInternal(ctx context.Context, prefixCidr string, ips []string) (res *ReleaseIPRes, error error) {
	prefix := i.PrefixFrom(ctx, prefixCidr)
	if prefix == nil {
		return nil, fmt.Errorf("%w: unable to find prefix for cidr:%s", ErrNotFound, prefixCidr)
	}
	res = &ReleaseIPRes{}
	for _, v := range ips {
		_, ok := prefix.Ips[v]
		if !ok {
			res.Result = append(res.Result, ReleaseResult{v, "释放失败,ErrNotFound IP"})
		} else {
			delete(prefix.Ips, v)
			res.Result = append(res.Result, ReleaseResult{v, "释放成功"})
		}
	}
	_, err := i.storage.UpdatePrefix(ctx, *prefix)
	if err != nil {
		return nil, fmt.Errorf("unable to release ip %w", err)
	}
	return res, nil
}

// PrefixesOverlapping will check if one ore more prefix of newPrefixes is overlapping
// with one of existingPrefixes
func PrefixesOverlapping(i *ipamer, existingPrefixes []string, newPrefixes []string) error {
	for _, np := range newPrefixes {
		nip, err := netip.ParsePrefix(np)
		if err != nil {
			return fmt.Errorf("parsing prefix %s failed:%w", np, err)
		}
		for _, ep := range existingPrefixes {
			eip, err := netip.ParsePrefix(ep)
			if err != nil {
				return fmt.Errorf("parsing prefix %s failed:%w", np, err)
			}
			iprange := netipx.RangeOfPrefix(eip)
			if nip.Contains(iprange.From()) {
				return fmt.Errorf("parsing prefix %s failed,Error already has subnet", np)
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, ep := range existingPrefixes {
		eip, err := netip.ParsePrefix(ep)
		if err != nil {
			return fmt.Errorf("parsing prefix %s failed:%w", ep, err)
		}
		for _, np := range newPrefixes {
			nip, err := netip.ParsePrefix(np)
			if err != nil {
				return fmt.Errorf("parsing prefix %s failed:%w", np, err)
			}
			iprange := netipx.RangeOfPrefix(nip)
			if eip.Contains(iprange.From()) {
				prefix := i.PrefixFrom(ctx, ep)
				if prefix == nil {
					continue
				}
				for ip := iprange.From(); eip.Contains(ip); ip = ip.Next() {
					ipstring := ip.String()
					prefix.Ips[ipstring] = IPDetail{"networkman", np, "子网段使用", tools.DateToString()}
				}
				_, err := i.storage.UpdatePrefix(ctx, *prefix)
				if err != nil {
					return err
				}
			}
			// if eip.Overlaps(nip) || nip.Overlaps(eip) {
			// 	return fmt.Errorf("%s overlaps %s", nip, eip)
			// }
		}
	}
	return nil
}

// newPrefix create a new Prefix from a string notation.
func (i *ipamer) newPrefix(cidr, gateway string, vlanId int, vrf, idc, parentCidr string, isParent bool) (*Prefix, error) {
	ipnet, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, fmt.Errorf("cidr errors")
	}
	// parentnet, err := netip.ParsePrefix(parentCidr)
	// if err != nil {
	// 	return nil, fmt.Errorf("parentCidr errors")
	// }
	GW, err := netip.ParseAddr(gateway)
	if err != nil || !ipnet.Contains(GW) {
		return nil, fmt.Errorf("gateway errors")
	}
	if parentCidr != "" {
		ipnetParent, err := netip.ParsePrefix(parentCidr)
		if err != nil {
			return nil, fmt.Errorf("unable to parse parent cidr:%s %w", cidr, err)
		}
		parentCidr = ipnetParent.Masked().String()
	}
	p := &Prefix{
		Gateway:                gateway,
		VlanID:                 vlanId,
		VRF:                    vrf,
		IDC:                    idc,
		Cidr:                   ipnet.Masked().String(),
		ParentCidr:             parentCidr,
		Ips:                    make(map[string]IPDetail),
		availableChildPrefixes: make(map[string]bool),
		IsParent:               isParent,
	}

	// FIXME: should this be done by the user ?
	// First ip in the prefix and broadcast is blocked.
	iprange := netipx.RangeOfPrefix(ipnet)
	p.Ips[gateway] = IPDetail{"networkman", "networkman", "网关地址", tools.DateToString()}
	p.Ips[iprange.From().String()] = IPDetail{"networkman", "networkman", "网络地址", tools.DateToString()}
	if ipnet.Addr().Is4() {
		// broadcast is ipv4 only
		p.Ips[iprange.To().String()] = IPDetail{"networkman", "networkman", "广播地址", tools.DateToString()}
	}

	return p, nil
}

func (i *ipamer) Dump(ctx context.Context) (string, error) {
	pfxs, err := i.storage.ReadAllPrefixes(ctx)
	if err != nil {
		return "", err
	}
	js, err := pfxs.toJSON()
	if err != nil {
		return "", err
	}
	return string(js), nil
}

func (i *ipamer) Load(ctx context.Context, dump string) error {
	existingpfxs, err := i.storage.ReadAllPrefixes(ctx)
	if err != nil {
		return err
	}
	if len(existingpfxs) > 0 {
		return fmt.Errorf("prefixes exist, please drop existing data before loading")
	}
	pfxs, err := fromJSONs([]byte(dump))
	if err != nil {
		return err
	}
	err = i.storage.DeleteAllPrefixes(ctx)
	if err != nil {
		return err
	}
	for _, pfx := range pfxs {
		_, err = i.storage.CreatePrefix(ctx, pfx)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadAllPrefixCidrs retrieves all existing Prefix CIDRs from the underlying storage
func (i *ipamer) ReadAllPrefixCidrs(ctx context.Context) ([]string, error) {
	return i.storage.ReadAllPrefixCidrs(ctx)
}

func (p *Prefix) String() string {
	return p.Cidr
}

func (u *Usage) String() string {
	if u.AcquiredPrefixes == 0 {
		return fmt.Sprintf("ip:%d/%d", u.AcquiredIPs, u.AvailableIPs)
	}
	return fmt.Sprintf("ip:%d/%d prefixes alloc:%d avail:%d", u.AcquiredIPs, u.AvailableIPs, u.AcquiredPrefixes, u.AvailableSmallestPrefixes)
}

// Network return the net.IP part of the Prefix
func (p *Prefix) Network() (netip.Addr, error) {
	ipprefix, err := netip.ParsePrefix(p.Cidr)
	if err != nil {
		return netip.Addr{}, err
	}
	return ipprefix.Addr(), nil
}

// hasIPs will return true if there are allocated IPs
func (p *Prefix) hasIPs() bool {
	ipprefix, err := netip.ParsePrefix(p.Cidr)
	if err != nil {
		return false
	}
	if ipprefix.Addr().Is4() && len(p.Ips) > 2 {
		return true
	}
	if ipprefix.Addr().Is6() && len(p.Ips) > 1 {
		return true
	}
	return false
}

// availableips return the number of ips available in this Prefix
func (p *Prefix) availableips() uint64 {
	ipprefix, err := netip.ParsePrefix(p.Cidr)
	if err != nil {
		return 0
	}
	// We don't report more than 2^31 available IPs by design
	if (ipprefix.Addr().BitLen() - ipprefix.Bits()) > 31 {
		return math.MaxInt32
	}
	return 1 << (ipprefix.Addr().BitLen() - ipprefix.Bits())
}

// acquiredips return the number of ips acquired in this Prefix
func (p *Prefix) acquiredips() uint64 {
	return uint64(len(p.Ips))
}

// availablePrefixes will return the amount of prefixes allocatable and the amount of smallest 2 bit prefixes
func (p *Prefix) availablePrefixes() (uint64, []string) {
	prefix, err := netip.ParsePrefix(p.Cidr)
	if err != nil {
		return 0, nil
	}
	var ipsetBuilder netipx.IPSetBuilder
	ipsetBuilder.AddPrefix(prefix)
	for cp, available := range p.availableChildPrefixes {
		if available {
			continue
		}
		ipprefix, err := netip.ParsePrefix(cp)
		if err != nil {
			continue
		}
		ipsetBuilder.RemovePrefix(ipprefix)
	}

	ipset, err := ipsetBuilder.IPSet()
	if err != nil {
		return 0, []string{}
	}

	// Only 2 Bit Prefixes are usable, set max bits available 2 less than max in family
	maxBits := prefix.Addr().BitLen() - 2
	pfxs := ipset.Prefixes()
	totalAvailable := uint64(0)
	availablePrefixes := []string{}
	for _, pfx := range pfxs {
		// same as: totalAvailable += uint64(math.Pow(float64(2), float64(maxBits-pfx.Bits)))
		totalAvailable += 1 << (maxBits - pfx.Bits())
		availablePrefixes = append(availablePrefixes, pfx.String())
	}
	// we are not reporting more that 2^31 available prefixes
	if totalAvailable > math.MaxInt32 {
		totalAvailable = math.MaxInt32
	}
	return totalAvailable, availablePrefixes
}

// acquiredPrefixes return the amount of acquired prefixes of this prefix if this is a parent prefix
func (p *Prefix) acquiredPrefixes() uint64 {
	var count uint64
	for _, available := range p.availableChildPrefixes {
		if !available {
			count++
		}
	}
	return count
}

// Usage report Prefix usage.
func (p *Prefix) Usage() Usage {
	sp, ap := p.availablePrefixes()
	return Usage{
		AvailableIPs:              p.availableips(),
		AcquiredIPs:               p.acquiredips(),
		AcquiredPrefixes:          p.acquiredPrefixes(),
		AvailableSmallestPrefixes: sp,
		AvailablePrefixes:         ap,
	}
}

// NoIPAvailableError indicates that the acquire-operation could not be executed
// because the specified prefix has no free IP anymore.
type NoIPAvailableError struct {
}

func (o NoIPAvailableError) Error() string {
	return "NoIPAvailableError"
}

// NotFoundError is raised if the given Prefix or Cidr was not found
type NotFoundError struct {
}

func (o NotFoundError) Error() string {
	return "NotFound"
}

// OptimisticLockError indicates that the operation could not be executed because the dataset to update has changed in the meantime.
// clients can decide to read the current dataset and retry the operation.
type OptimisticLockError struct {
}

func (o OptimisticLockError) Error() string {
	return "OptimisticLockError"
}

// AlreadyAllocatedError is raised if the given address is already in use
type AlreadyAllocatedError struct {
}

func (o AlreadyAllocatedError) Error() string {
	return "AlreadyAllocatedError"
}

// retries the given function if the reported error is an OptimisticLockError
// with ten attempts and jitter delay ~100ms
// returns only error of last failed attempt
func retryOnOptimisticLock(retryableFunc retry.RetryableFunc) error {

	return retry.Do(
		retryableFunc,
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, ErrOptimisticLockError)
		}),
		retry.Attempts(10),
		retry.DelayType(retry.CombineDelay(retry.BackOffDelay, retry.RandomDelay)),
		retry.LastErrorOnly(true))
}

