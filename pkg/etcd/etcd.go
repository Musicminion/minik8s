package etcd

import (
	"context"
	"time"

	etcd "go.etcd.io/etcd/client/v3"
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
	Value           string
}

type ListRes struct {
	ResourceVersion int64
	CreateVersion   int64
	Key             string
	Value           string
}

// 创建一个新的Etcd客户端存储
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

// 修复了一下逻辑，通过返回的时候创建一个切片，而不是直接组装
func (s *Store) Get(key string) ([]ListRes, error) {
	response, err := s.client.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}
	if len(response.Kvs) == 0 {
		return nil, nil
	}
	// 定义一个ListRes的切片，用于存储结果
	var res []ListRes

	// 遍历response.Kvs，将每一个key-value转换为ListRes
	for id, kv := range response.Kvs {
		res = append(res, ListRes{
			ResourceVersion: response.Header.Revision,
			CreateVersion:   response.Kvs[id].CreateRevision,
			Key:             string(kv.Key),
			Value:           string(kv.Value),
		})
		// 如果id超过1，就退出循环
		if id >= 1 {
			break
		}
	}

	return res, nil
	// return []ListRes{ListRes{
	// 	ResourceVersion: response.Kvs[0].ModRevision,
	// 	CreateVersion:   response.Kvs[0].CreateRevision,
	// 	Key:             string(response.Kvs[0].Key),
	// 	Value:           string(response.Kvs[0].Value),
	// }}, nil
}

// 值得注意的是，多次Put一个相同的key，会覆盖之前的值！
func (s *Store) Put(key string, val []byte) error {
	_, err := s.client.Put(context.TODO(), key, string(val))
	return err
}

// 删除指定的Key
func (s *Store) Del(key string) error {
	_, err := s.client.Delete(context.TODO(), key)
	return err
}

// 删除所有的key
func (s *Store) DelAll() error {
	_, err := s.client.Delete(context.TODO(), "", etcd.WithPrefix())
	return err
}

func convertEventToWatchRes(event *etcd.Event) WatchRes { // 根据event的类型转换为不同的WatchRes
	res := WatchRes{
		ResourceVersion: event.Kv.ModRevision,
		CreateVersion:   event.Kv.CreateRevision,
		IsCreate:        event.IsCreate(),
		IsModify:        event.IsModify(),
		Key:             string(event.Kv.Key),
	}
	switch event.Type {
	case etcd.EventTypePut:
		res.ResType = PUT
		res.Value = string(event.Kv.Value)
	case etcd.EventTypeDelete:
		res.ResType = DELETE
	}
	return res
}

func (s *Store) Watch(key string) (context.CancelFunc, <-chan WatchRes) {
	ctx, cancel := context.WithCancel(context.TODO())
	watchResChan := make(chan WatchRes)
	go func(c chan<- WatchRes) { // 匿名函数
		for watchResponse := range s.client.Watch(ctx, key) {
			for _, event := range watchResponse.Events {
				res := convertEventToWatchRes(event)
				c <- res
			}
		}
		close(c)
	}(watchResChan)

	return cancel, watchResChan
}

func (s *Store) PrefixWatch(key string) (context.CancelFunc, <-chan WatchRes) {
	ctx, cancel := context.WithCancel(context.TODO())
	watchResChan := make(chan WatchRes)
	go func(c chan<- WatchRes) {
		for watchResponse := range s.client.Watch(ctx, key, etcd.WithPrefix()) {
			for _, event := range watchResponse.Events {
				res := convertEventToWatchRes(event)
				c <- res
			}
		}
		close(c)
	}(watchResChan)
	return cancel, watchResChan
}

// 之前写的是直接赋值ret[i],这样会寄!,应该调用append
func (s *Store) PrefixGet(key string) ([]ListRes, error) {
	response, err := s.client.Get(context.TODO(), key, etcd.WithPrefix())
	if err != nil {
		println(err)
		return []ListRes{}, err
	}
	var ret []ListRes
	println(len(response.Kvs))
	// 遍历response.Kvs，将每一个key-value转换为ListRes
	for id, kv := range response.Kvs {
		ret = append(ret, ListRes{
			ResourceVersion: response.Header.Revision,
			CreateVersion:   response.Kvs[id].CreateRevision,
			Key:             string(kv.Key),
			Value:           string(kv.Value),
		})
	}

	// for i, kv := range response.Kvs {
	// 	ret[i] = ListRes{
	// 		ResourceVersion: kv.ModRevision,
	// 		CreateVersion:   kv.CreateRevision,
	// 		Key:             string(kv.Key),
	// 		Value:           string(kv.Value),
	// 	}
	// }
	return ret, nil
}
