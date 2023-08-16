package note

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	redigo "github.com/redis/go-redis/v9"
	// Use v8 for simpler context handling
)

type redisDB struct {
	rdb  *redigo.Client
	lock sync.RWMutex
}

func NewRedis(ctx context.Context, redisOptions *redis.Options) (*redisDB, error) {
	client := redis.NewClient(redisOptions)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	rdb := redigo.NewClient(&redigo.Options{
		Addr:     redisOptions.Addr,
		Password: redisOptions.Password,
		DB:       redisOptions.DB,
	})
	return &redisDB{rdb: rdb, lock: sync.RWMutex{}}, nil
}

func (r *redisDB) Name() string {
	return "redis"
}

func (r *redisDB) CreateNote(ctx context.Context, n *Note) (*Note, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	existing, err := r.rdb.Exists(ctx, n.Instance).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to read existing note:%v, error:%w", n, err)
	}
	if existing != 0 {
		return nil, fmt.Errorf("note:%v already exists", n)
	}
	value, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	err = r.rdb.Set(ctx, n.Instance, value, 0).Err()
	return n, err
}

func (r *redisDB) DeleteAllNote(ctx context.Context) error {
	r.lock.RLock()
	defer r.lock.RUnlock()
	_, err := r.rdb.FlushAll(ctx).Result()
	return err
}

func (r *redisDB) ReadAllNote(ctx context.Context) (Notes, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	instances, err := r.rdb.Keys(ctx, "*").Result()
	if err != nil {
		return nil, fmt.Errorf("unable to get all note:%w", err)
	}

	result := Notes{}
	for _, note := range instances {
		vjs, err := r.rdb.Get(ctx, note).Bytes()
		if err != nil {
			return nil, err
		}
		note, err := fromJSON(vjs)
		if err != nil {
			return nil, err
		}
		result = append(result, *note)
	}
	return result, nil
}

func (r *redisDB) DeleteNote(ctx context.Context, instance string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	_, err := r.rdb.Del(ctx, instance).Result()
	if err != nil {
		return err
	}
	return nil
}
