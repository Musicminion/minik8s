package etcdstore

import (
	"context"
	// "fmt"
	etcd "go.etcd.io/etcd/clientv3"
	// "minik8s/pkg/klog"
	"time"
)

type Store struct {
	client *etcd.Client
}
type WatchResType int

const (
	PUT    WatchResType = 0
	DELETE WatchResType = 1
)

type String interface {
	ToString() string
}

type WatchRes struct {
	ResType         WatchResType
	ResourceVersion int64
	CreateVersion   int64
	IsCreate        bool // true when ResType == PUT and the key is new
	IsModify        bool // true when ResType == PUT and the key is old
	Key             string
	ValueBytes      []byte
}

type ListRes struct {
	ResourceVersion int64
	CreateVersion   int64
	Key             string
	ValueBytes      []byte
}

func NewEtcdStore(endpoints []string, timeout time.Duration) (*Store, error) {
	cli, err := etcd.New(etcd.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})
	if err != nil {
		return nil, err
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // 取消上下文，避免资源泄露
	_, err = cli.Status(timeoutContext, endpoints[0])
	if err != nil {
		return nil, err
	}
	return &Store{client: cli}, nil
}

func (s *Store) Get(key string) ([]ListRes, error) {
	response, err := s.client.Get(context.TODO(), key)
	if err != nil {
		return []ListRes{}, err
	}
	if len(response.Kvs) == 0 {
		return []ListRes{}, nil
	} else {
		listRes := ListRes{
			ResourceVersion: response.Kvs[0].ModRevision,
			CreateVersion:   response.Kvs[0].CreateRevision,
			Key:             string(response.Kvs[0].Key),
			ValueBytes:      response.Kvs[0].Value,
		}
		return []ListRes{listRes}, nil
	}
}

func (s *Store) Put(key string, val []byte) error {
	_, err := s.client.Put(context.TODO(), key, string(val))
	return err
}

func (s *Store) Del(key string) error {
	_, err := s.client.Delete(context.TODO(), key)
	return err
}

func (s *Store) Watch(key string) (context.CancelFunc, <-chan WatchRes) {
	ctx, cancel := context.WithCancel(context.TODO())
	watchResChan := make(chan WatchRes)
	watch := func(c chan<- WatchRes) {
		watchChan := s.client.Watch(ctx, key)
		for watchResponse := range watchChan {
			for _, event := range watchResponse.Events {
				var res WatchRes
				switch event.Type {
				case etcd.EventTypePut:
					res.ResType = PUT
					res.ResourceVersion = event.Kv.ModRevision
					res.CreateVersion = event.Kv.CreateRevision
					res.IsCreate = event.IsCreate()
					res.IsModify = event.IsModify()
					res.Key = string(event.Kv.Key)
					res.ValueBytes = event.Kv.Value
					break
				case etcd.EventTypeDelete:
					res.ResType = DELETE
					res.ResourceVersion = event.Kv.ModRevision
					res.CreateVersion = event.Kv.CreateRevision
					res.IsCreate = event.IsCreate()
					res.IsModify = event.IsModify()
					res.Key = string(event.Kv.Key)
					break
				}
				c <- res
			}
		}
		close(c)
	}
	go watch(watchResChan)

	return cancel, watchResChan
}

func (s *Store) PrefixWatch(key string) (context.CancelFunc, <-chan WatchRes) {
	ctx, cancel := context.WithCancel(context.TODO())
	watchResChan := make(chan WatchRes)
	watch := func(c chan<- WatchRes) {
		watchChan := s.client.Watch(ctx, key, etcd.WithPrefix())
		for watchResponse := range watchChan {
			for _, event := range watchResponse.Events {
				var res WatchRes
				switch event.Type {
				case etcd.EventTypePut:
					res.ResType = PUT
					res.ResourceVersion = event.Kv.ModRevision
					res.CreateVersion = event.Kv.CreateRevision
					res.IsCreate = event.IsCreate()
					res.IsModify = event.IsModify()
					res.Key = string(event.Kv.Key)
					res.ValueBytes = event.Kv.Value
					break
				case etcd.EventTypeDelete:
					res.ResType = DELETE
					res.ResourceVersion = event.Kv.ModRevision
					res.CreateVersion = event.Kv.CreateRevision
					res.IsCreate = event.IsCreate()
					res.IsModify = event.IsModify()
					res.Key = string(event.Kv.Key)
					res.ValueBytes = event.Kv.Value
					break
				}
				c <- res
			}
		}
		close(c)
	}
	go watch(watchResChan)
	return cancel, watchResChan
}

func (s *Store) PrefixGet(key string) ([]ListRes, error) {
	response, err := s.client.Get(context.TODO(), key, etcd.WithPrefix())
	if err != nil {
		return []ListRes{}, err
	}
	var ret []ListRes
	for _, kv := range response.Kvs {
		res := ListRes{
			ResourceVersion: kv.ModRevision,
			CreateVersion:   kv.CreateRevision,
			Key:             string(kv.Key),
			ValueBytes:      kv.Value,
		}
		ret = append(ret, res)
	}
	return ret, nil
}
