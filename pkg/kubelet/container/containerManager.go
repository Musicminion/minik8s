package container

// import (
// 	"context"

// 	"github.com/containerd/containerd"
// )

// // 定义容器管理者的接口
// type ContainerManager interface {
// 	// 创建容器
// 	CreateContainer(config *createContainerConfig) error
// }

// type ContainerManagerImpl struct {
// }

// func (cml *ContainerManagerImpl) CreateContainer(config *createContainerConfig) error {
// 	// 创建一个containerd client
// 	client, err := containerd.New("/run/containerd/containerd.sock")
// 	if err != nil {
// 		return err
// 	}
// 	defer client.Close()
// 	// 创建一个context
// 	context := context.Background()
// 	// 使用名字空间

// 	//

// 	client.NewContainer(
// 		context,
// 		"my-container",
// 		containerd.WithImage(),
// 	)

// }
