package runtime

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/kubelet/runtime/container"
	"miniK8s/pkg/kubelet/runtime/image"
)

type RuntimeManager interface {
	CreatePod(pod *apiObject.PodStore) error
	DeletePod(pod *apiObject.PodStore) error
	StartPod(pod *apiObject.PodStore) error
	StopPod(pod *apiObject.PodStore) error
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
	pauseID, err := r.createPauseContainer(pod)

	if err != nil {
		return err
	}

	// 创建pod中的所有容器
	_, err = r.createPodAllContainer(pod, pauseID)

	if err != nil {
		return err
	}

	// TODO:
	return nil
}

// DeletePod 删除pod
func (r *runtimeManager) DeletePod(pod *apiObject.PodStore) error {
	// TODO:
	// 先删除pod中的所有容器
	_, err := r.removePodAllContainer(pod)

	if err != nil {
		return err
	}

	// 最后再删除pause容器
	_, err = r.removePauseContainer(pod)

	if err != nil {
		return err
	}

	return nil
}

// StartPod 启动pod
func (r *runtimeManager) StartPod(pod *apiObject.PodStore) error {
	// 先启动pause容器
	_, err := r.startPauseContainer(pod)

	if err != nil {
		return err
	}

	// 启动pod中的所有容器
	_, err = r.startPodAllContainer(pod)

	if err != nil {
		return err
	}

	return nil
}

// StopPod 停止pod
func (r *runtimeManager) StopPod(pod *apiObject.PodStore) error {
	// 先停止pod中的所有容器
	_, err := r.stopPodAllContainer(pod)

	if err != nil {
		return err
	}

	// 最后停止pause容器
	_, err = r.stopPauseContainer(pod)

	if err != nil {
		return err
	}

	return nil

}
