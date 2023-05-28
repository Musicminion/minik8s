package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/app/etcdclient"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/k8log"
	"miniK8s/util/uuid"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
)

// 这个库用来存放测试代码中可能用到的工具函数

// 创建Docker虚拟网络
func SetupTestNetwork(containerID, containerIP string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}

	// 创建一个名为'test_network'的Docker虚拟网络
	ipamConfig := []network.IPAMConfig{
		{
			Subnet: "192.168.0.0/16",
		},
	}

	ipam := &network.IPAM{
		Config: ipamConfig,
	}

	networkConfig := types.NetworkCreate{
		CheckDuplicate: true,
		IPAM:           ipam,
	}

	networkRes, err := cli.NetworkCreate(ctx, containerID, networkConfig)
	if err != nil {
		return "", err
	}

	endpointConfig := &network.EndpointSettings{
		NetworkID: networkRes.ID, // 设置容器所属的网络 ID
		// 设置容器接口的 IP 地址和网关
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: containerIP,
		},
		Gateway: "192.168.0.1",
	}

	// 将容器连接到网络
	err = cli.NetworkConnect(ctx, networkRes.ID, containerID, endpointConfig)
	if err != nil {
		return "", err
	}

	// 返回容器的IP地址
	return containerIP, nil
}

// 清理Docker虚拟网络
func CleanupTestNetwork(containerID string) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	// 断开容器与网络的连接
	err = cli.NetworkDisconnect(ctx, containerID, containerID, true)
	if err != nil {
		return err
	}

	// 删除网络
	err = cli.NetworkRemove(ctx, containerID)
	if err != nil {
		return err
	}

	return nil
}

// 添加Pod到etcd中
func AddPodToEtcd() {
	
	for i := 1; i <= 2; i++ {
		path := "./testFile/yamlFile/Pod-" + fmt.Sprint(i) + ".yaml"
		file, _ := os.Open(path)

		// 读取文件内容
		content, err := io.ReadAll(file)
		if err != nil {
			k8log.ErrorLog("APIServer", "AddPod: read file failed "+err.Error())
			return
		}

		// 将文件内容转换为Pod对象
		// 通过调用gin引擎的ServeHTTP方法，可以模拟一个http请求，从而测试AddPod方法。
		pod := &apiObject.Pod{}
		err = yaml.Unmarshal(content, pod)
		if err != nil {
			k8log.ErrorLog("APIServer", "AddPod: unmarshal yaml failed "+err.Error())
			return
		}

		// 检查name是否重复
		newPodName := pod.GetObjectName()
		key := fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", pod.GetObjectNamespace(), newPodName)
		res, err := etcdclient.EtcdStore.Get(key)
		if len(res) != 0 {
			k8log.ErrorLog("APIServer", "AddPod: pod name has exist")
			return
		}
		if err != nil {
			k8log.ErrorLog("APIServer", "AddPod: get pod failed "+err.Error())
			return
		}
		pod.Metadata.UUID = uuid.NewUUID()

		// 把Pod转化为PodStore
		podStore := pod.ToStore()

		// 把PodStore转化为json
		podStoreJson, err := json.Marshal(podStore)
		if err != nil {
			k8log.ErrorLog("APIServer", "AddPod: marshal pod failed "+err.Error())
			return
		}

		key = fmt.Sprintf(serverconfig.EtcdPodPath+"%s/%s", pod.GetObjectNamespace(), newPodName)

		// 将pod存储到etcd中
		err = etcdclient.EtcdStore.Put(key, podStoreJson)
		if err != nil {
			k8log.ErrorLog("APIServer", "AddPod: put pod failed "+err.Error())
		}
	}
}
