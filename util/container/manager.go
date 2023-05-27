package container

import (
	"miniK8s/pkg/kubelet/runtime/container"
	"miniK8s/pkg/kubelet/runtime/image"
)

// 这个包主要作为非kubelet的容器管理时候使用的组件
// 用来管理容器的创建、删除、查询等操作

type HelperContainerManager struct {
	ContainerManager container.ContainerManager
	ImageManager     image.ImageManager
}

func NewHelperContainerManager() *HelperContainerManager {
	return &HelperContainerManager{
		ContainerManager: container.ContainerManager{},
		ImageManager:     image.ImageManager{},
	}
}

// 创建容器的方法,返回容器的ID和错误
