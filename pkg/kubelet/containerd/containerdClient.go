package containerdClient

import (
	"fmt"
	"github.com/containerd/containerd"
	// "github.com/containerd/containerd/oci"
	// "github.com/containerd/containerd/namespaces"
)

// ContainerdClient是一个统一的客户端，用于与containerd进行交互
// 为了避免反复new containerd.Client，ContainerdClient提供了一个内部的containerd.Client
type containerdClient struct {
	client *containerd.Client // containerd.Client
}


/*
 * 1. getClient为内部函数提供获取一个containerd.Client
 * 2. 如果containerd.Client已经存在，则直接返回
 * 3. 如果containerd.Client不存在，则创建一个新的containerd.Client
 * 4. 如果创建containerd.Client失败，则返回错误
*/ 
func (c *containerdClient) getClient() (*containerd.Client, error) {
	if c.client != nil {
		return c.client, nil
	}
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, err
	}
	c.client = client
	return c.client, nil
}

func (c *containerdClient) GetClient() (*containerd.Client, error) {
	return c.getClient()
}

/*
* closeClient关闭containerd.Client
*/ 

func (c *containerdClient) closeClient() error {
	if c.client == nil {
		return nil
	}
	result := c.client.Close()
	return result
}

func (c *containerdClient) CloseClient() error {
	return c.closeClient()
}
