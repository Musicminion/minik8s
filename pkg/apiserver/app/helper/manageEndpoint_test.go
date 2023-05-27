package helper

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"path"
	"testing"
)

var testPod = apiObject.PodStore{
	Basic: apiObject.Basic{
		APIVersion: "v1",
		Kind:       "Pod",
		Metadata: apiObject.Metadata{
			Name:      "testPod",
			Namespace: "testNamespace",
			UUID:      "1f3a54a3-c1b9-4e47-b063-2a6d84fde222",
			Labels: map[string]string{
				"app": "test",
			},
		},
	},
	Spec: apiObject.PodSpec{
		Containers: []apiObject.Container{
			{
				Name:  "testContainer-1",
				Image: "docker.io/library/redis",
				Ports: []apiObject.ContainerPort{
					{
						ContainerPort: "80",
					},
				},
			},
			{
				Name:  "testContainer-2",
				Image: "docker.io/library/nginx",
				Ports: []apiObject.ContainerPort{
					{
						ContainerPort: "8080",
					},
				},
			},
		},
	},
}

var testService = apiObject.ServiceStore{
	Basic: apiObject.Basic{
		APIVersion: "v1",
		Kind:       "Service",
		Metadata: apiObject.Metadata{
			Name:      "testService",
			Namespace: "testNamespace",
			UUID:      "1f3a54a3-c1b9-4e47-b063-2a6d84fde222",
		},
	},
	Spec: apiObject.ServiceSpec{
		Selector: map[string]string{
			"app": "test",
		},
		Ports: []apiObject.ServicePort{
			{
				Port:       80,
				TargetPort: 80,
				Name:       "testPort",
			},
		},
	},
}

func TestAddEndPoints(t *testing.T) {
	// 清空etcd
	etcdclient.EtcdStore.PrefixDel("/")
	
	// 向etcd中写入带有selector的service
	serviceJson, err := json.Marshal(testService)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	etcdURL := path.Join(serverconfig.EtcdServicePath, testService.Metadata.Namespace, testService.Metadata.Name)
	etcdclient.EtcdStore.Put(etcdURL, serviceJson)

	// etcdclient.EtcdStore.PrefixGet(path.Join(config.ServiceURL, "app", testService.Spec.Selector["app"]))
	err = UpdateEndPoints(testPod)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 验证cache中是否存在新的endpoint
	endpoints, err := GetEndpoints("app", testPod.Metadata.Labels["app"])
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 验证endpoints的size
	if len(endpoints) != len(testPod.Metadata.Labels) {
		t.Errorf("expected %+v, but got %+v", len(testPod.Metadata.Labels), len(endpoints))
	}

}

func TestGetEndpoints(t *testing.T) {
	// 创建测试用例

	err := UpdateEndPoints(testPod)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 执行测试

	endpoints, err := GetEndpoints("app", testPod.Metadata.Labels["app"])
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 验证结果
	if len(endpoints) != len(testPod.Metadata.Labels) {
		t.Errorf("expected %+v, but got %+v", len(testPod.Metadata.Labels), len(endpoints))
	}

	// 清空etcd
	etcdclient.EtcdStore.PrefixDel("/")

}
