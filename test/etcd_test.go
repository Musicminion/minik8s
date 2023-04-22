package test

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	// "miniK8s/pkg/etcd"
	// "context"
	etcd "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

type Store struct {
	client *etcd.Client
}

func TestEtcd(t *testing.T) {
	// cli, err := etcd.New(etcd.Config{
	// 	Endpoints:   []string{"localhost:2379"},
	// 	DialTimeout: 5 * time.Second,
	// })
	// if err != nil {
	// 	// 处理错误
	// }
	// defer cli.Close()
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// _, err = cli.Put(ctx, "my_key", "my_value")
	// cancel()
	// if err != nil {
	// 	// 处理错误
	// }
	// ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	// _, err = cli.Put(ctx, "my_key1", "my_value1")
	// cancel()
	// if err != nil {
	// 	// 处理错误
	// }

}

func TestNewServer(t *testing.T) {
	cli, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 5,
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
	s :=  &Store{client: cli}
	response, err := s.client.Get(context.TODO(), "666")
	if err != nil {
		return nil, err
	}
	if len(response.Kvs) == 0 {
		return nil, nil
	}
	assert.Equal(t, "999", response.Kvs[0])
}
