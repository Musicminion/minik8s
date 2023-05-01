package runtime

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/kubelet/runtime/container"
	"miniK8s/pkg/kubelet/runtime/image"
)

type RuntimeManager interface {
	CreatePod(pod *apiObject.PodStore) error
	DeletePod(pod *apiObject.PodStore) error
}

type runtimeManager struct {
	// 用于管理容器、镜像的管理器
	containerManager container.ContainerManager
	imageManager     image.ImageManager
}

// NewRuntimeManager 创建一个RuntimeManager
func NewRuntimeManager() RuntimeManager {
	manager := &runtimeManager{
		containerManager: container.ContainerManager{},
		imageManager:     image.ImageManager{},
	}
	return manager
}

// ************************************************************
// 这里写RunTimeManager的函数

// CreatePod 创建pod
func (r *runtimeManager) CreatePod(pod *apiObject.PodStore) error {
	// 创建pause容器
	_, err := r.createPauseContainer(pod)

	if err != nil {
		return err
	}

	// 创建pod中的所有容器
	_, err = r.createPodAllContainer(pod)

	if err != nil {
		return err
	}

	// TODO:
	return nil
}

// DeletePod 删除pod
func (r *runtimeManager) DeletePod(pod *apiObject.PodStore) error {
	// TODO:

	// 最后再删除pause容器
	r.removePauseContainer(pod)
	return nil
}
