package testutil

import (
	"context"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

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
