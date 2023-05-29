package proxy

import (
	"encoding/json"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	"miniK8s/util/uuid"
	"os"
	"strconv"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

func TestSyncLoopIteration_CreateService(t *testing.T) {
	proxy := NewKubeProxy(listwatcher.DefaultListwatcherConfig())
	go proxy.Run()
	// 等待proxy成功启动
	time.Sleep(3 * time.Second)

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
	if err != nil {
		t.Fatal(err)
	}

	// 给Service设置UUID, 所以哪怕用户故意设置UUID也会被覆盖
	service.Metadata.UUID = uuid.NewUUID()

	// // 将Service转化为ServiceStore
	// serviceStore := service.ToServiceStore()

	serviceUpdate := &entity.ServiceUpdate{
		Action:        message.CREATE,
		ServiceTarget: apiObject.ServiceStore{},
	}

	for key, value := range service.Spec.Selector {
		func(key, value string) {
			// 替换可变参 namespace
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
					serviceUpdate.ServiceTarget.Status.Endpoints = append(serviceUpdate.ServiceTarget.Status.Endpoints, endpoint)
				}
				k8log.DebugLog("APIServer", "endpoints number of service "+service.GetObjectName()+" is "+strconv.Itoa(len(serviceUpdate.ServiceTarget.Status.Endpoints)))
				// serviceUpdate.ServiceTarget.Endpoints = append(serviceUpdate.ServiceTarget.Endpoints, endpoints...)
			}
		}(key, value)
	}

	// 向消息队列发送消息
	message.PublishUpdateService(serviceUpdate)

	// assert.Equal(t, "syncLoopIteration: create Service action", k8log.LastLog())
	// assert.True(t, proxy.iptableManager.(*MockIptableManager).CreateServiceCalled)
}

// func TestHandleDnsUpdate(t *testing.T) {
// 	// 执行测试函数
// 	proxy := NewKubeProxy(listwatcher.DefaultListwatcherConfig())
// 	go proxy.Run()
// 	// 等待proxy成功启动
// 	time.Sleep(3 * time.Second)

// 	hostList := []string{"192.168.0.1 example.com"}
// 	message.PubelishUpdateHost(hostList)

// 	time.Sleep(1 * time.Second)
// 	// 检查 hosts 文件是否正确生成
// 	expectedOutput := "127.0.0.1 localhost\n192.168.0.1 example.com\n"
// 	fileContent, err := os.ReadFile(config.HostsConfigFilePath)
// 	if err != nil {
// 		t.Fatalf("Failed to read file: %v", err)
// 	}
// 	if len(fileContent) == 0 {
// 		t.Fatalf("File is empty")
// 	}
// 	if len(fileContent) != len(expectedOutput) {
// 		t.Log(len(fileContent))
// 		t.Log(len(expectedOutput))
// 		t.Fatalf("File content length is not correct")
// 	}
// }
