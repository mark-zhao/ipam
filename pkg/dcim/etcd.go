package dcim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdDB is a note.Storage implementation backed by ETCD
type etcdDB struct {
	client *clientv3.Client
	lock   sync.RWMutex
}

func NewEtcd(ctx context.Context, etcdConfig clientv3.Config) (Storage, error) {
	client, err := clientv3.New(etcdConfig)
	if err != nil {
		return nil, err
	}

	return &etcdDB{client: client, lock: sync.RWMutex{}}, nil
}

func (etcd *etcdDB) Name() string {
	return "etcd"
}

func (etcd *etcdDB) CreateIDC(ctx context.Context, idc IDC) (IDC, error) {
	etcd.lock.Lock()
	defer etcd.lock.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	get, err := etcd.client.Get(ctx, idc.IDCName)
	defer cancel()
	if err != nil {
		return IDC{}, fmt.Errorf("unable to read existing IDC:%v, error:%w", idc.IDCName, err)
	}

	if get.Count != 0 {
		return IDC{}, fmt.Errorf("idc already exists:%v", idc.IDCName)
	}

	value, err := json.Marshal(idc)
	if err != nil {
		return IDC{}, err
	}

	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	_, err = etcd.client.Put(ctx, idc.IDCName, string(value))
	defer cancel()
	if err != nil {
		return IDC{}, errors.New("数据录入数据库失败")
	}
	return idc, nil
}

func (etcd *etcdDB) DeleteAllIDC(ctx context.Context) error {
	etcd.lock.RLock()
	defer etcd.lock.RUnlock()
	ctx, cancel := context.WithTimeout(ctx, 50*time.Minute)
	defaultOpts := []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithKeysOnly(), clientv3.WithSerializable()}
	idcs, err := etcd.client.Get(ctx, "", defaultOpts...)
	defer cancel()
	if err != nil {
		return fmt.Errorf("unable to get all idc :%w", err)
	}

	for _, idc := range idcs.Kvs {
		_, err := etcd.client.Delete(ctx, string(idc.Key))
		if err != nil {
			return fmt.Errorf("unable to delete idc:%w", err)
		}
	}
	return err
}

func (etcd *etcdDB) ReadAllIDC(ctx context.Context) (IDCS, error) {
	etcd.lock.Lock()
	defer etcd.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defaultOpts := []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithKeysOnly(), clientv3.WithSerializable()}
	pfxs, err := etcd.client.Get(ctx, "", defaultOpts...)
	defer cancel()
	if err != nil {
		return nil, fmt.Errorf("unable to get all prefix cidrs:%w", err)
	}

	result := IDCS{}
	for _, idc := range pfxs.Kvs {
		v, err := etcd.client.Get(ctx, string(idc.Key))
		if err != nil {
			return nil, err
		}
		idc, err := fromJSON(v.Kvs[0].Value)
		if err != nil {
			return nil, err
		}
		result = append(result, idc)
	}
	return result, nil
}

func (etcd *etcdDB) DeleteIDC(ctx context.Context, idcname string) error {
	etcd.lock.Lock()
	defer etcd.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	_, err := etcd.client.Delete(ctx, idcname)
	defer cancel()
	if err != nil {
		return err
	}
	return nil
}

func (etcd *etcdDB) ReadIDC(ctx context.Context, idcname string) (IDC, error) {
	etcd.lock.Lock()
	defer etcd.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	get, err := etcd.client.Get(ctx, idcname)
	defer cancel()
	if err != nil {
		return IDC{}, fmt.Errorf("unable to read data from ETCD error:%w", err)
	}

	if get.Count == 0 {
		return IDC{}, fmt.Errorf("unable to read existing idc:%v, error:%w", idcname, err)
	}

	return fromJSON(get.Kvs[0].Value)
}

func (etcd *etcdDB) UpdateIDC(ctx context.Context, idc IDC) (IDC, error) {
	etcd.lock.Lock()
	defer etcd.lock.Unlock()

	idcjs, err := json.Marshal(idc)
	if err != nil {
		return IDC{}, fmt.Errorf("unable to marshal prefixes:%w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	p, err := etcd.client.Get(ctx, idc.IDCName)
	defer cancel()
	if err != nil {
		return IDC{}, fmt.Errorf("unable to read cidrs from ETCD:%w", err)
	}

	if p.Count == 0 {
		return IDC{}, fmt.Errorf("unable to get all idc:%w", err)
	}

	// Operation is committed only if the watched keys remain unchanged.
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	_, err = etcd.client.Put(ctx, idc.IDCName, string(idcjs))
	defer cancel()
	if err != nil {
		return IDC{}, fmt.Errorf("unable to update idc:%s, error:%w", idc.IDCName, err)
	}

	return idc, nil
}

func fromJSON(js []byte) (IDC, error) {
	var idc IDC
	err := json.Unmarshal(js, &idc)
	if err != nil {
		return IDC{}, fmt.Errorf("unable to unmarshal note:%w", err)
	}
	return idc, nil
}
