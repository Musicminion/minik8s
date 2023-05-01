package helper

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/etcdclient"

	// "miniK8s/pkg/apiserver/app/handlers"

	// "miniK8s/pkg/apiserver/app/handlers"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"path"
	"sync"
)

// 定义缓存对象
var cache = struct {
	endpoints map[string][]apiObject.Endpoint
	sync.RWMutex
}{endpoints: make(map[string][]apiObject.Endpoint)}

func GetEndpoints(key, value string) ([]apiObject.Endpoint, error) {
	// 构建终端数组URL
	endpointsKVURL := path.Join(config.EndpointURL, key, value)
	//TODO: 从缓存中查找endpoint
	cache.RLock()
	if endpoints, ok := cache.endpoints[endpointsKVURL]; ok {
		cache.RUnlock()
		return endpoints, nil
	}
	cache.RUnlock()

	// 从Etcd中获取终端数组
	endpointsJsonStr, err := etcdclient.EtcdStore.Get(endpointsKVURL)
	if err != nil {
		return nil, err
	}

	// 构造endpoint map并解析endpointsJson
	endpoints := make(entity.Endpoints)
	if len(endpointsJsonStr) != 0 {
		if err := json.Unmarshal([]byte(endpointsJsonStr[0].Value), &endpoints); err != nil {
			return nil, err
		}
	}

	// 并发获取每个终端
	endpointChan := make(chan apiObject.Endpoint, len(endpoints))
	var wg sync.WaitGroup
	for _, arr := range endpoints {
		for _, UID := range arr {
			endpointURL := path.Join(config.EndpointURL, UID)
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				if endpointStr, err := etcdclient.EtcdStore.Get(url); err == nil {
					endpoint := apiObject.Endpoint{}
					if err := json.Unmarshal([]byte(endpointStr[0].Value), &endpoint); err == nil {
						endpointChan <- endpoint
					}
				}
			}(endpointURL)
		}
	}
	wg.Wait()
	close(endpointChan)

	// 从通道中读取所有终端
	endpointArray := make([]apiObject.Endpoint, 0, len(endpointChan))
	for endpoint := range endpointChan {
		endpointArray = append(endpointArray, endpoint)
	}

	// TODO: 更新缓存数组
	cache.Lock()
	cache.endpoints[endpointsKVURL] = endpointArray
	cache.Unlock()

	return endpointArray, nil
}
