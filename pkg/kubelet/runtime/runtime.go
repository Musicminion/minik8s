package runtime

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/runtime/container"
	"miniK8s/pkg/kubelet/runtime/image"
	minik8sTypes "miniK8s/pkg/minik8sTypes"

	"github.com/docker/docker/api/types"
)

type RuntimeManager interface {
	CreatePod(pod *apiObject.PodStore) error
	DeletePod(pod *apiObject.PodStore) error
	StartPod(pod *apiObject.PodStore) error
	StopPod(pod *apiObject.PodStore) error
	RestartPod(pod *apiObject.PodStore) error
	DelPodByPodID(podUUID string) error
	RecreatePodContainer(pod *apiObject.PodStore) error
	ExecPodContainer(pod *apiObject.PodStore, cmd []string) (string, error)

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
		k8log.DebugLog("Runtime Manager", err.Error())
		return err
	}

	// 创建pod中的所有容器
	_, err = r.createPodAllContainer(pod, pauseID)

	if err != nil {
		k8log.ErrorLog("Runtime Manager", err.Error())
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

// DelPodByID 通过pod的UUID删除pod
func (r *runtimeManager) DelPodByPodID(podUUID string) error {
	// 通过podUUID获取pod
	filter := make(map[string][]string)
	filter[minik8sTypes.ContainerLabel_PodUID] = []string{podUUID}

	// 根据容器的名字过滤器，过滤出来所有的容器
	res, err := r.containerManager.ListContainersWithOpt(filter)

	if err != nil {
		k8log.ErrorLog("Runtime Manager", err.Error())
		return err
	}

	// 遍历所有的容器，然后删除
	for _, container := range res {
		_, err := r.containerManager.RemoveContainer(container.ID)
		if err != nil {
			k8log.ErrorLog("Runtime Manager", err.Error())
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// 判断当前的container是否在给定的containers中
func contains(runContainers []types.Container, containerName string) bool {
	for _, runContainer := range runContainers {
		// 注意，docker的容器名字是以/开头的 ！
		if runContainer.Names[0] == "/"+containerName {
			return true
		}
	}
	return false
}

// 重启pod中的缺失的容器（Pause容器除外）
func (r *runtimeManager) RecreatePodContainer(pod *apiObject.PodStore) error {
	// 获取运行中的Pod的所有容器
	filter := make(map[string][]string)
	filter[minik8sTypes.ContainerLabel_PodUID] = []string{pod.GetPodUUID()}

	// 根据pod的UUID过滤容器
	runContainers, err := r.containerManager.ListContainersWithOpt(filter)
	// 从容器中筛选出pauseContainer

	var pauseContainerID string
	for _, container := range runContainers {
		if container.Labels[minik8sTypes.ContainerLabel_IfPause] == minik8sTypes.ContainerLabel_IfPause_True {
			pauseContainerID = container.ID
		}
	}

	if err != nil {
		return err
	}

	for _, container := range pod.Spec.Containers {
		if !contains(runContainers, container.Name) {
			// 重启容器
			k8log.InfoLog("kubelet", fmt.Sprintf("Recreate container %s in pod %s", container.Name, pod.GetPodName()))
			r.createPodContainer(pod, &container, pauseContainerID)
		}
	}

	return nil
}

// 根据pod的UUID筛选container， 在容器中执行命令
func (r *runtimeManager) ExecPodContainer(pod *apiObject.PodStore, cmd []string) (string, error) {
	k8log.DebugLog("Runtime Manager", "exec pod container, pod name is "+pod.GetPodName())
	// 根据pod的名字找出所有匹配的pod
	filter := make(map[string][]string)
	filter[minik8sTypes.ContainerLabel_PodUID] = []string{pod.GetPodUUID()}
	// 根据pod的UUID过滤容器
	runContainers, err := r.containerManager.ListContainersWithOpt(filter)
	if err != nil {
		k8log.ErrorLog("Runtime Manager", err.Error())
		return "", err
	}

	for _, container := range runContainers {
		_, err = r.containerManager.ExecContainer(container.ID, cmd)
		if err != nil {
			k8log.ErrorLog("Runtime Manager", err.Error())
			return "", err
		}
	}
	return "", nil
}