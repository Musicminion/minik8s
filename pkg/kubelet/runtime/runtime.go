package runtime

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/runtime/container"
	"miniK8s/pkg/kubelet/runtime/image"
)

type RuntimeManager interface {
	CreatePod(pod *apiObject.PodStore) error
	DeletePod(pod *apiObject.PodStore) error
	StartPod(pod *apiObject.PodStore) error
	StopPod(pod *apiObject.PodStore) error
	RestartPod(pod *apiObject.PodStore) error

	// GetRuntimeNodeStatus 获取运行时Node的状态信息
	GetRuntimeNodeStatus() (*apiObject.NodeStatus, error)

	// GetRuntimeAllPodStatus 获取运行时Pod的状态信息
	GetRuntimeAllPodStatus() (map[string]*RunTimePodStatus, error)

	// GetRuntimePodStatus 获取运行时Node的名字
	GetRuntimeNodeName() string

	// GetRuntimeNodeIp 获取运行时Node的IP
	GetRuntimeNodeIP() (string, error)
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
		k8log.DebugLog("[Runtime Manager]", err.Error())
		return err
	}

	// 创建pod中的所有容器
	_, err = r.createPodAllContainer(pod, pauseID)

	if err != nil {
		k8log.ErrorLog("[Runtime Manager]", err.Error())
		return err
	}

	LogStr := "[Runtime Manager] create pod success" + pod.GetPodName()
	k8log.InfoLog("kubelet", LogStr)
	// TODO:  send pod info to apiserver
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

	LogStr := "[Runtime Manager] delete pod success" + pod.GetPodName()
	k8log.InfoLog("kubelet", LogStr)
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

	LogStr := "[Runtime Manager] start pod success" + pod.GetPodName()
	k8log.InfoLog("kubelet", LogStr)
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

	LogStr := "[Runtime Manager] stop pod success" + pod.GetPodName()
	k8log.InfoLog("kubelet", LogStr)

	return nil

}

// RestartPod 重启pod
func (r *runtimeManager) RestartPod(pod *apiObject.PodStore) error {
	// 先重启pause容器
	_, err := r.restartPauseContainer(pod)

	if err != nil {
		return err
	}

	// 重启pod中的所有容器
	_, err = r.restartPodAllContainer(pod)

	if err != nil {
		return err
	}

	LogStr := "[Runtime Manager] restart pod success" + pod.GetPodName()
	k8log.InfoLog("kubelet", LogStr)
	return nil
}
