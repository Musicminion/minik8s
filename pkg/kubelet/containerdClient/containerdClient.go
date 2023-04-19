package containerdClient

import "github.com/containerd/containerd"

// 封装一层，设置minik8s的名字空间，方便调用
func NewContainerdClient() (*containerd.Client, error) {
	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace("minik8s"))
	if err != nil {
		return nil, err
	}
	return client, nil
}
