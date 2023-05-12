package runtime

import (
	"miniK8s/pkg/apiObject"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/util/host"
	"time"

	"github.com/docker/docker/api/types"
)

// 获取运行时Node的状态信息
func (r *runtimeManager) GetRuntimeNodeStatus() (*apiObject.NodeStatus, error) {
	hostname := host.GetHostName()
	nodeIp, err := host.GetHostIp()

	if err != nil {
		return nil, err
	}

	nodeCondition := apiObject.NodeCondition(apiObject.Ready)
	nodeCpuPercent, err := host.GetHostSystemCPUUsage()
	if err != nil {
		return nil, err
	}

	nodeMemPercent, err := host.GetHostSystemMemoryUsage()
	if err != nil {
		return nil, err
	}

	runPods, _, _, err := r.GetRuntimeAllPodStatus()

	if err != nil {
		return nil, err
	}

	nodePodNum := len(runPods)

	nodeStatus := apiObject.NodeStatus{
		Hostname:   hostname,
		Ip:         nodeIp,
		Condition:  nodeCondition,
		CpuPercent: nodeCpuPercent,
		MemPercent: nodeMemPercent,
		NumPods:    nodePodNum,
		UpdateTime: time.Now(),
	}

	return &nodeStatus, nil
}

// 获取运行时Pod的状态信息
// 返回的参数是(map[podUUID]->PodStatus, map[podUUID]->PodName, map[PodUUID]->PodNamespace)
func (r *runtimeManager) GetRuntimeAllPodStatus() (map[string]*apiObject.PodStatus, map[string]string, map[string]string, error) {
	containers, err := r.containerManager.ListContainers()
	if err != nil {
		return nil, nil, nil, err
	}

	// 创建一个map，从podID到一组容器的映射
	podIDToContainers := make(map[string][]types.Container)
	// 创建一个map，从podID到PodName的映射
	podIDToPodName := make(map[string]string)
	// 创建一个map，从podID到PodNamespace的映射
	podIDToPodNamespace := make(map[string]string)

	// 遍历所有容器，将容器按照podID进行分类
	for _, container := range containers {
		podID := container.Labels[minik8stypes.ContainerLabel_PodUID]

		// 如果不是Pod相关的容器，就跳过
		if podID == "" {
			continue
		}

		podName := container.Labels[minik8stypes.ContainerLabel_PodName]

		if podName == "" {
			continue
		}

		podNamespace := container.Labels[minik8stypes.ContainerLabel_PodNamespace]

		if podNamespace == "" {
			continue
		}

		// 最后一块写入
		podIDToContainers[podID] = append(podIDToContainers[podID], container)
		podIDToPodName[podID] = container.Labels[minik8stypes.ContainerLabel_PodName]
		podIDToPodNamespace[podID] = container.Labels[minik8stypes.ContainerLabel_PodNamespace]

	}

	// 创建一个map，从podID到PodStatus的映射
	podIDToPodStatus := make(map[string]*apiObject.PodStatus)

	// // 遍历所有的Pod，将Pod的状态信息填充到podIDToPodStatus中
	for podID, containers := range podIDToContainers {
		// 创建一个PodStatus
		podStatus := apiObject.PodStatus{}

		// 遍历所有的容器，获取容器的状态信息
		for _, container := range containers {
			containerID := container.ID
			res, err := r.containerManager.GetContainerInspectInfo(containerID)

			// 将容器的状态信息转换为Pod的状态信息
			containerStatus := r.ParseInspectInfoToContainerState(res)

			if err != nil {
				return nil, nil, nil, err
			}

			podStatus.ContainerStatuses = append(podStatus.ContainerStatuses, *containerStatus)
		}

		// 将Pod的状态信息填充到podIDToPodStatus中
		podIDToPodStatus[podID] = &podStatus
	}

	return podIDToPodStatus, nil, nil, nil
}

func (r *runtimeManager) ParseInspectInfoToContainerState(inspectInfo *types.ContainerJSON) *types.ContainerState {

	containerState := types.ContainerState{
		Status:     inspectInfo.State.Status,
		StartedAt:  inspectInfo.State.StartedAt,
		FinishedAt: inspectInfo.State.FinishedAt,
		Health:     inspectInfo.State.Health,
		Error:      inspectInfo.State.Error,
		ExitCode:   inspectInfo.State.ExitCode,
		Pid:        inspectInfo.State.Pid,
		Running:    inspectInfo.State.Running,
		Paused:     inspectInfo.State.Paused,
		Restarting: inspectInfo.State.Restarting,
		OOMKilled:  inspectInfo.State.OOMKilled,
		Dead:       inspectInfo.State.Dead,
	}

	return &containerState
}

func (r *runtimeManager) GetRuntimeNodeName() string {
	hostname := host.GetHostName()
	return hostname
}
