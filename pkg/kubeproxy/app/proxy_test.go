package proxy

import (
	"encoding/json"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/etcdclient"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/util/stringutil"
	"miniK8s/util/uuid"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)


func TestSyncLoopIteration_CreateService(t *testing.T) {
	proxy := NewKubeProxy(listwatcher.DefaultListwatcherConfig())
	go proxy.Run()
	// 等待proxy成功启动
	time.Sleep(5 * time.Second)

	// read from yaml
	filepath := "../testFile/Service.yaml"
	file, err := os.Open(filepath)
	if err != nil {
		t.Fatal(err)
	}
	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	// 将文件内容转换为Pod对象
	// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddPod方法。
	service := &apiObject.Service{}
	err = yaml.Unmarshal(content, service)

	// 给Service设置UUID, 所以哪怕用户故意设置UUID也会被覆盖
	service.Metadata.UUID = uuid.NewUUID()

	// // 将Service转化为ServiceStore
	// serviceStore := service.ToServiceStore()

	serviceUpdate := &entity.ServiceUpdate{
		Action: entity.CREATE,
		ServiceTarget: entity.ServiceWithEndpoints{
			Service:   *service,
			Endpoints: make([]apiObject.Endpoint, 0),
		},
	}

	for key, value := range service.Spec.Selector {
		func(key, value string) {
			// 替换可变参 namespace
			etcdURL := path.Join(config.ServiceURL, key, value, service.Metadata.UUID)
			etcdURL = stringutil.Replace(etcdURL, config.URL_PARAM_NAMESPACE_PART, service.GetNamespace())
			res, err := etcdclient.EtcdStore.PrefixGet(serverconfig.EtcdPodPath)

			if err != nil {
				return
			} else {
				// iterate res
				for _, r := range res {
					// pod to endpoint
					pod := &apiObject.PodStore{}
					if err := json.Unmarshal([]byte(r.Value), pod); err != nil {
						return
					}
					endpoint := apiObject.Endpoint{}
					endpoint.Metadata.Name = pod.GetPodName()
					endpoint.Metadata.Namespace = pod.GetPodNamespace()
					endpoint.Metadata.UUID = pod.Metadata.UUID
					endpoint.Metadata.Labels = pod.Metadata.Labels
					endpoint.Metadata.Annotations = pod.Metadata.Annotations
					endpoint.IP = pod.Status.PodIP
					// add endpoint to serviceUpdate.ServiceTarget.Endpoints
					serviceUpdate.ServiceTarget.Endpoints = append(serviceUpdate.ServiceTarget.Endpoints, endpoint)
				}
				k8log.DebugLog("APIServer", "endpoints number of service "+service.GetName()+" is "+strconv.Itoa(len(serviceUpdate.ServiceTarget.Endpoints)))
				// serviceUpdate.ServiceTarget.Endpoints = append(serviceUpdate.ServiceTarget.Endpoints, endpoints...)
			}
		}(key, value)
	}

	// 向消息队列发送消息
	msgutil.PublishUpdateService(serviceUpdate)

	// assert.Equal(t, "syncLoopIteration: create Service action", k8log.LastLog())
	// assert.True(t, proxy.iptableManager.(*MockIptableManager).CreateServiceCalled)
}
