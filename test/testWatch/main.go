package main

// import (
// 	"fmt"
// 	"miniK8s/pkg/etcd"
// 	"sync"
// 	"time"

// 	clientv3 "go.etcd.io/etcd/client/v3"
// )

// func main() {
// 	config := clientv3.Config{
// 		Endpoints: []string{"localhost:2379"}, // etcd 服务器地址
// 	}
// 	client, err := clientv3.New(config)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer client.Close()

// 	store, err := etcd.NewEtcdStore([]string{"localhost:2379"}, 5*time.Second)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// 监听以 /registry/pods 为前缀的所有事件
// 	cancel, watchResChan := store.PrefixWatch("/")

// 	var wg sync.WaitGroup
// 	wg.Add(1)

// 	// 处理监听到的事件
// 	go func() {
// 		fmt.Println("Start watching11...")
// 		for res := range watchResChan {
// 			fmt.Printf("Received event: %v\n", res)
// 		}
// 		wg.Done()
// 		return
// 		cancel()
// 	}()
// 	wg.Wait()
// 	// 停止监听
// 	defer cancel()
// }
