package runtime

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	minik8stypes "miniK8s/pkg/minik8sTypes"
	"miniK8s/util/host"
	"miniK8s/util/weave"
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

	runTimePodStatus, err := r.GetRuntimeAllPodStatus()

	if err != nil {
		return nil, err
	}

	nodePodNum := len(runTimePodStatus)

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
// 返回的参数是map[string]*RunTimePodStatus
func (r *runtimeManager) GetRuntimeAllPodStatus() (map[string]*RunTimePodStatus, error) {
	containers, err := r.containerManager.ListContainers()
	if err != nil {
		return nil, err
	}

	// 创建一个map，从podID到一组容器的映射
	podIDToContainers := make(map[string][]types.Container)

	// 创建一个map，从podID到PodStatus的映射,这个将会作为返回值
	podIDToPodStatus := make(map[string]*RunTimePodStatus)

	// // 创建一个map，从podID到PodName的映射
	// podIDToPodName := make(map[string]string)
	// // 创建一个map，从podID到PodNamespace的映射
	// podIDToPodNamespace := make(map[string]string)

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
		podIDToPodStatus[podID] = &RunTimePodStatus{
			PodID:        podID,
			PodName:      podName,
			PodNamespace: podNamespace,
			PodStatus:    apiObject.PodStatus{},
		}

		// // 将Pod的名字和命名空间写入

		// podIDToPodName[podID] = container.Labels[minik8stypes.ContainerLabel_PodName]
		// podIDToPodNamespace[podID] = container.Labels[minik8stypes.ContainerLabel_PodNamespace]
	}

	// // 遍历所有的Pod，将Pod的状态信息填充到podIDToPodStatus中
	for podID, containers := range podIDToContainers {
		// 创建一个PodStatus
		podStatus := apiObject.PodStatus{}
		// 创建一个containerIP的数组
		podIPs := []string{}

		// 遍历所有的容器，获取容器的状态信息
		for _, container := range containers {
			containerID := container.ID
			res, err := r.containerManager.GetContainerInspectInfo(containerID)

			// 将容器的状态信息转换为Pod的状态信息
			containerStatus := r.ParseInspectInfoToContainerState(res)

			if err != nil {
				return nil, err
			}

			podStatus.ContainerStatuses = append(podStatus.ContainerStatuses, *containerStatus)

			cpuPercent, memoryPercent, err := r.containerManager.CalculateContainerResource(containerID)
			if err != nil {
				k8log.ErrorLog("kubelet", err.Error())
				continue
			}
			// 将容器的资源使用情况累加到Pod的资源使用情况中
			podStatus.CpuPercent += cpuPercent
			podStatus.MemPercent += memoryPercent

			// 通过Weave网络获取容器的IP
			containerIP, err := weave.WeaveFindIpByContainerID(containerID)

			if err != nil {
				logStr := "GetRuntimeAllPodStatus: " + err.Error()
				k8log.ErrorLog("kubelet", logStr)
			}
			podIPs = append(podIPs, containerIP)
		}

		// 将Pod的状态信息填充到podIDToPodStatus中
		// podIDToPodStatus[podID]
		// 检查podIDToPodStatus里面是否有这个podID
		// 如果有，就直接写入，反之忽略
		_, ok := podIDToPodStatus[podID]

		if ok {
			// 需要添加Pod的其他信息
			// 处理Pod的IP问题，需要检查所有容器的IP是否相同
			// 检查Pod的IP是否为空
			if len(podIPs) != 0 {
				// 遍历检查所有容器的IP是否相同
				for _, podIP := range podIPs {
					if podIP != podIPs[0] {
						k8log.WarnLog("kubelet", "GetRuntimeAllPodStatus: PodIP is not the same")
						break
					}
				}
				podStatus.PodIP = podIPs[0]
			}

			// 然后处理Pod的状态信息
			podCalculatedPhase, err := r.CalculatePodPhaseByContainerStatus(&podStatus.ContainerStatuses)

			if err != nil {
				k8log.ErrorLog("kubelet", "GetRuntimeAllPodStatus: "+err.Error())
				podStatus.Phase = apiObject.PodUnknown
				podStatus.UpdateTime = time.Now()
				podIDToPodStatus[podID].PodStatus = podStatus
			} else {
				// 将计算出来的Pod的状态信息写入
				podStatus.Phase = podCalculatedPhase
				podStatus.UpdateTime = time.Now()
				podIDToPodStatus[podID].PodStatus = podStatus
			}
		}
	}

	// 组装返回值
	return podIDToPodStatus, nil
}

// 注意：没有容器的时候，Pod的状态是Pending
func (r *runtimeManager) CalculatePodPhaseByContainerStatus(allContainerStatus *[]types.ContainerState) (string, error) {

	// 如果没有容器，就直接返回
	if len(*allContainerStatus) == 0 {
		return apiObject.PodPending, nil
	}

	// ==========================================================
	isPodRunning := true
	// 如果有容器，就遍历所有容器，检查容器的状态
	for _, containerStatus := range *allContainerStatus {
		isPodRunning = isPodRunning && containerStatus.Running
	}
	if isPodRunning {
		return apiObject.PodRunning, nil
	}

	// ==========================================================
	// 如果有容器正在终止，就返回Terminating
	isPodTerminating := false
	// 如果有容器，就遍历所有容器，检查容器的状态
	for _, containerStatus := range *allContainerStatus {
		isPodTerminating = isPodTerminating || containerStatus.Dead
	}
	if isPodTerminating {
		return apiObject.PodTerminating, nil
	}

	// ==========================================================
	// 如果所有容器都终止了，就检查所有容器的退出码
	allContainerTerminated := true
	for _, containerStatus := range *allContainerStatus {
		allContainerTerminated = allContainerTerminated && containerStatus.Dead
	}

	if allContainerTerminated {
		// 检查退出码
		for _, containerStatus := range *allContainerStatus {
			if containerStatus.ExitCode != 0 {
				return apiObject.PodFailed, nil
			}
		}
		return apiObject.PodSucceeded, nil
	}

	// ==========================================================
	// 反之就是未知状态
	return apiObject.PodUnknown, nil
}

func (r *runtimeManager) ParseInspectInfoToContainerState(inspectInfo *types.ContainerJSON) *types.ContainerState {
	if inspectInfo == nil {
		return &types.ContainerState{}
	}

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

func (r *runtimeManager) GetRuntimeNodeIP() (string, error) {
	nodeIp, err := host.GetHostIp()

	if err != nil {
		return "", err
	}
	return nodeIp, nil
}
