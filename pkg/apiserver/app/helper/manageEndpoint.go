package helper

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/etcdclient"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/k8log"
	"miniK8s/util/uuid"

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
	endpointsKVURL := path.Join(serverconfig.EndpointPath, key, value)
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
			endpointURL := path.Join(serverconfig.EndpointPath, UID)
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

func AddEndPoints(pod apiObject.PodStore) error {
	// 构建终端数组URL
	for key, value := range pod.Metadata.Labels {
		endpointsKVURL := path.Join(serverconfig.EndpointPath, key, value)
		// 从ETCD取出所有符合该Label的endpoints
		endpointLRs, err := etcdclient.EtcdStore.PrefixGet(endpointsKVURL)
		if err != nil {
			k8log.ErrorLog("APIServer", "get endpoints failed"+err.Error())
			return err
		}

		// unmarshall endpoint
		totalEndpoints := make([]apiObject.Endpoint, 0)
		for _, endpointLR := range endpointLRs {
			endpoint := apiObject.Endpoint{}
			// endpoint的存储URL ：EndpointURL + key + value + endpointUUID
			if json.Unmarshal([]byte(endpointLR.Value), &endpoint); err != nil {
				k8log.ErrorLog("APIServer", "unmarshal endpoint failed"+err.Error())
				return err
			}
			totalEndpoints = append(totalEndpoints, endpoint)
		}

		// 添加新的endpoint
		// newEndpoints := make([]apiObject.Endpoint, 0)
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				endpoint := apiObject.Endpoint{
					Basic: apiObject.Basic{
						Metadata: apiObject.Metadata{
							UUID: uuid.NewUUID(),
						},
					},
					IP:   pod.Status.PodIP, // TODO: alloc ip for pod
					Port: port.ContainerPort,
				}
				// 更新endpoint map
				k8log.DebugLog("APIServer", "add endpoint uuid: "+endpoint.Metadata.UUID)
				endpointJson, err := json.Marshal(endpoint)
				if err != nil {
					k8log.ErrorLog("APIServer", "marshal endpoint failed"+err.Error())
					return err
				}

				// 将新的endpoint添加到etcd中
				etcdclient.EtcdStore.Put(path.Join(endpointsKVURL, endpoint.Metadata.UUID), endpointJson)
				// newEndpoints = append(newEndpoints, endpoint)
				totalEndpoints = append(totalEndpoints, endpoint)
			}
		}

		// 根据Label从etcd找出所有匹配的service
		serviceLRs, err := etcdclient.EtcdStore.PrefixGet(path.Join(config.ServiceURL, key, value))
		if err != nil {
			k8log.ErrorLog("APIServer", "get service failed"+err.Error())
			return err
		}

		// 对每个service，发送一个UpdateEndpoint的消息
		for _, serviceLR := range serviceLRs {
			serviceStore := apiObject.ServiceStore{}
			if err := json.Unmarshal([]byte(serviceLR.Value), &serviceStore); err != nil {
				k8log.ErrorLog("APIServer", "unmarshal service failed"+err.Error())
			}
			// 创建用于更新service的endpointUpdate对象，
			endpointUpdate := &entity.EndpointUpdate{
				Action: entity.CREATE,
				ServiceTarget: entity.ServiceWithEndpoints{
					Endpoints: totalEndpoints,
					Service:   serviceStore,
				},
			}
			// 加入到消息队列中以便kubeproxy更新service
			err = msgutil.PublishUpdateEndpoints(endpointUpdate)
			if err != nil {
				k8log.ErrorLog("APIServer", "publish endpoint update message failed"+err.Error())
			}

			// // 将endpoint持久化到etcd中
			// endpointJson, err := json.Marshal(endpointMap)
			// if err != nil {
			// 	k8log.ErrorLog("APIServer", "marshal endpoint failed"+err.Error())
			// 	return err
			// }
			// if err := etcdclient.EtcdStore.Put(endpointsKVURL, endpointJson); err != nil {
			// 	k8log.ErrorLog("APIServer", "put endpoint failed"+err.Error())
			// 	return err
			// }
		}
		// TODO: 更新缓存数组
		cache.Lock()
		cache.endpoints[endpointsKVURL] = totalEndpoints
		cache.Unlock()
	}

	return nil

}
