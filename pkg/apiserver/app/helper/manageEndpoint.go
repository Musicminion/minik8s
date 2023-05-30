package helper

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/message"
	"miniK8s/util/uuid"

	// "miniK8s/pkg/apiserver/app/handlers"

	// "miniK8s/pkg/apiserver/app/handlers"

	"miniK8s/pkg/entity"
	"path"
)

// 根据key和value获取所有的endpoints
func GetEndpoints(key, value string) ([]apiObject.Endpoint, error) {
	// 构建终端数组URL
	endpointsKVURL := path.Join(serverconfig.EndpointPath, key, value)

	// 从Etcd中获取终端数组
	endpointsJsonStr, err := etcdclient.EtcdStore.PrefixGet(endpointsKVURL)
	if err != nil {
		return nil, err
	}

	//解析endpointsJson
	endpoints := []apiObject.Endpoint{}
	for _, endpointJson := range endpointsJsonStr {
		endpoint := apiObject.Endpoint{}
		if err := json.Unmarshal([]byte(endpointJson.Value), &endpoint); err == nil {
			endpoints = append(endpoints, endpoint)
		} else {
			k8log.ErrorLog("APIServer", "Unmarshal endpoints failed"+err.Error())
			return nil, err
		}
	}

	return endpoints, nil
}

func UpdateEndPoints(pod apiObject.PodStore) error {
	// 构建终端数组URL
	for key, value := range pod.Metadata.Labels {
		endpointsKVURL := path.Join(serverconfig.EndpointPath, key, value)
		totalEndpoints, err := GetEndpoints(key, value)
		if err != nil {
			k8log.ErrorLog("APIServer", "get endpoints failed"+err.Error())
			return err
		}

		exist := false

		// 寻找totalEndpoints中是否已经存在对应该pod的endpoint
		for _, endpoint := range totalEndpoints {
			if endpoint.PodUUID == pod.Metadata.UUID {
				exist = true
				break
			}
		}

		if !exist {
			// 添加新的endpoint
			endpoint := apiObject.Endpoint{
				Basic: apiObject.Basic{
					Metadata: apiObject.Metadata{
						UUID: uuid.NewUUID(),
					},
				},
				IP:      pod.Status.PodIP,
				Ports:   []string{},
				PodUUID: pod.Metadata.UUID,
			}
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					// 更新endpoint的port
					k8log.DebugLog("APIServer", "add endpoint uuid: "+endpoint.Metadata.UUID)
					endpoint.Ports = append(endpoint.Ports, port.ContainerPort)
				}
			}
			// 将新的endpoint添加到etcd中
			// endpoint的URL： /registry/endpoint/key/value/podUUID
			endpointJson, err := json.Marshal(endpoint)
			if err != nil {
				k8log.ErrorLog("APIServer", "marshal endpoint failed"+err.Error())
				return err
			}
			etcdclient.EtcdStore.Put(path.Join(endpointsKVURL, endpoint.PodUUID), endpointJson)
			totalEndpoints = append(totalEndpoints, endpoint)
		}

		// 根据Label从etcd找出所有匹配的service
		serviceLRs, err := etcdclient.EtcdStore.PrefixGet(path.Join(serverconfig.EtcdServiceSelectorPath, key, value))
		if err != nil {
			k8log.ErrorLog("APIServer", "get service failed"+err.Error())
			return err
		}

		// 对每个service，发送一个UpdateService的消息
		for _, serviceLR := range serviceLRs {
			serviceStore := apiObject.ServiceStore{}
			if err := json.Unmarshal([]byte(serviceLR.Value), &serviceStore); err != nil {
				k8log.ErrorLog("APIServer", "unmarshal service failed"+err.Error())
			}
			// 更新service的endpoints
			serviceStore.Status.Endpoints = totalEndpoints

			// 创建用于更新service的serviceUpdate对象，
			serviceUpdate := &entity.ServiceUpdate{
				Action:        message.UPDATE,
				ServiceTarget: serviceStore,
			}

			// 写入etcd
			serviceJson, err := json.Marshal(serviceStore)
			if err != nil {
				k8log.ErrorLog("APIServer", "marshal service failed"+err.Error())
				return err
			}
			svcSelectorURL := path.Join(serverconfig.EtcdServiceSelectorPath, key, value, serviceStore.Metadata.UUID)
			err = etcdclient.EtcdStore.Put(svcSelectorURL, serviceJson)
			if err != nil {
				k8log.ErrorLog("APIServer", "put service failed"+err.Error())
				return err
			}
			svcSelectorURL = path.Join(serverconfig.EtcdServicePath, serviceStore.Metadata.Namespace, serviceStore.Metadata.Name)
			err = etcdclient.EtcdStore.Put(svcSelectorURL, serviceJson)
			if err != nil {
				k8log.ErrorLog("APIServer", "put service failed"+err.Error())
				return err
			}

			// 加入到消息队列中以便kubeproxy更新service
			k8log.DebugLog("APIServer", "PublishUpdateService")
			err = message.PublishUpdateService(serviceUpdate)
			if err != nil {
				k8log.ErrorLog("APIServer", "publish endpoint update message failed"+err.Error())
			}
		}
	}

	return nil

}

func DeleteEndpoints(pod apiObject.PodStore) error {
	for key, value := range pod.Metadata.Labels {
		endpointsKVURL := path.Join(serverconfig.EndpointPath, key, value)
		totalEndpoints, err := GetEndpoints(key, value)
		if err != nil {
			k8log.ErrorLog("APIServer", "get endpoints failed"+err.Error())
			return err
		}

		// 从totalEndpoints中删除对应该pod的endpoint
		for index, endpoint := range totalEndpoints {
			if endpoint.PodUUID == pod.Metadata.UUID {
				// 从etcd中删除endpoint
				etcdclient.EtcdStore.Del(path.Join(endpointsKVURL, endpoint.PodUUID))
				// 从totalEndpoints中删除endpoint
				totalEndpoints = append(totalEndpoints[:index], totalEndpoints[index+1:]...)
				break
			}
		}

		// 根据Label从etcd找出所有匹配的service
		serviceLRs, err := etcdclient.EtcdStore.PrefixGet(path.Join(serverconfig.EtcdServiceSelectorPath, key, value))
		if err != nil {
			k8log.ErrorLog("APIServer", "get service failed"+err.Error())
			return err
		}

		// 对每个service，发送一个UpdateService的消息
		for _, serviceLR := range serviceLRs {
			serviceStore := apiObject.ServiceStore{}
			if err := json.Unmarshal([]byte(serviceLR.Value), &serviceStore); err != nil {
				k8log.ErrorLog("APIServer", "unmarshal service failed"+err.Error())
			}
			// 更新service的endpoints
			serviceStore.Status.Endpoints = totalEndpoints
			// 创建用于更新service的serviceUpdate对象，
			serviceUpdate := &entity.ServiceUpdate{
				Action:        message.UPDATE,
				ServiceTarget: serviceStore,
			}

			// 写入etcd
			serviceJson, err := json.Marshal(serviceStore)
			if err != nil {
				k8log.ErrorLog("APIServer", "marshal service failed"+err.Error())
				return err
			}
			svcSelectorURL := path.Join(serverconfig.EtcdServiceSelectorPath, key, value, serviceStore.Metadata.UUID)
			err = etcdclient.EtcdStore.Put(svcSelectorURL, serviceJson)
			if err != nil {
				k8log.ErrorLog("APIServer", "put service failed"+err.Error())
				return err
			}
			svcSelectorURL = path.Join(serverconfig.EtcdServicePath, serviceStore.Metadata.Namespace, serviceStore.Metadata.Name)
			err = etcdclient.EtcdStore.Put(svcSelectorURL, serviceJson)
			if err != nil {
				k8log.ErrorLog("APIServer", "put service failed"+err.Error())
				return err
			}

			// 加入到消息队列中以便kubeproxy更新service
			err = message.PublishUpdateService(serviceUpdate)
			if err != nil {
				k8log.ErrorLog("APIServer", "publish endpoint update message failed"+err.Error())
			}
		}

	}

	return nil
}
