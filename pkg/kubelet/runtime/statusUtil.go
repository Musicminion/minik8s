package runtime

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/util/host"
	"time"
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

	nodeStatus := apiObject.NodeStatus{
		Hostname:   hostname,
		Ip:         nodeIp,
		Condition:  nodeCondition,
		CpuPercent: nodeCpuPercent,
		MemPercent: nodeMemPercent,
		NumPods:    0,
		UpdateTime: time.Now(),
	}

	return &nodeStatus, nil
}

// 获取运行时Pod的状态信息
func (r *runtimeManager) GetRuntimePodStatus(pod *apiObject.PodStore) ([]*apiObject.PodStatus, error) {
	r.containerManager.ListContainers()
}
