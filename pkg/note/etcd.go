package note

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

func (etcd *etcdDB) CreateNote(ctx context.Context, n *Note) (*Note, error) {
	etcd.lock.Lock()
	defer etcd.lock.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	get, err := etcd.client.Get(ctx, n.Instance)
	defer cancel()
	if err != nil {
		return nil, fmt.Errorf("unable to read existing note:%v, error:%w", n.Instance, err)
	}

	if get.Count != 0 {
		return nil, fmt.Errorf("note already exists:%v", n.Instance)
	}

	value, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	_, err = etcd.client.Put(ctx, n.Instance, string(value))
	defer cancel()
	if err != nil {
		return nil, errors.New("数据录入数据库失败")
	}

	return n, nil
}

func (etcd *etcdDB) DeleteAllNote(ctx context.Context) error {
	etcd.lock.RLock()
	defer etcd.lock.RUnlock()
	ctx, cancel := context.WithTimeout(ctx, 50*time.Minute)
	defaultOpts := []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithKeysOnly(), clientv3.WithSerializable()}
	notes, err := etcd.client.Get(ctx, "", defaultOpts...)
	defer cancel()
	if err != nil {
		return fmt.Errorf("unable to get all note :%w", err)
	}

	for _, note := range notes.Kvs {
		_, err := etcd.client.Delete(ctx, string(note.Key))
		if err != nil {
			return fmt.Errorf("unable to delete note:%w", err)
		}
	}
	return err
}

func (etcd *etcdDB) ReadAllNote(ctx context.Context) (Notes, error) {
	etcd.lock.Lock()
	defer etcd.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defaultOpts := []clientv3.OpOption{clientv3.WithPrefix(), clientv3.WithKeysOnly(), clientv3.WithSerializable()}
	pfxs, err := etcd.client.Get(ctx, "", defaultOpts...)
	defer cancel()
	if err != nil {
		return nil, fmt.Errorf("unable to get all prefix cidrs:%w", err)
	}

	result := Notes{}
	for _, idc := range pfxs.Kvs {
		v, err := etcd.client.Get(ctx, string(idc.Key))
		if err != nil {
			return nil, err
		}
		note, err := fromJSON(v.Kvs[0].Value)
		if err != nil {
			return nil, err
		}
		result = append(result, *note)
	}
	return result, nil
}

func (etcd *etcdDB) DeleteNote(ctx context.Context, instance string) error {
	etcd.lock.Lock()
	defer etcd.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	_, err := etcd.client.Delete(ctx, instance)
	defer cancel()
	if err != nil {
		return err
	}
	return nil
}
