package pleg

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/kubelet/runtime"
)

// 用来比较运行的状态和缓存的状态，然后生成pleg事件

type RunTimePodStatusMap map[string]*runtime.RunTimePodStatus
type CachePodsMap map[string]*apiObject.PodStore

// 更新Pleg里面的缓存
func (p *plegManager) updatePlegRecord(runtimePodStatus RunTimePodStatusMap, cachePods CachePodsMap) error {
	// 对于所有的runtimePodStatus，更新plegRecord
	errStr := ""
	for podID, runtimePodStatus := range runtimePodStatus {
		err := p.UpdatePodRecord(podID, runtimePodStatus)

		if err != nil {
			errStr += fmt.Sprintf("update podRecord error: %s", err.Error())
		}
	}

	if errStr != "" {
		return fmt.Errorf(errStr)
	}

	return nil
}

// RuntimePodStatusMap 是PodID到运行时的Pod状态的映射
// CachePodsMap 是PodID到缓存的Pod的映射
// 计算出不同的Pod，返回需要删除的Pod和需要添加的Pod
func (p *plegManager) plegGenerator(runtimePodStatus RunTimePodStatusMap, cachePods CachePodsMap) error {
	// 计算出不同的Pod，返回需要删除的Pod和需要添加的Pod
	p.calculateDiffPods(runtimePodStatus, cachePods)

	// 之后处理runtimePodStatus和cachePods的里面都有的Pod
	for podID := range runtimePodStatus {
		_, ok := cachePods[podID]
		if ok {
			// 比较两个Pod的状态，然后生成pleg事件
			p.comparePodStatus(runtimePodStatus[podID], cachePods[podID])
		}
	}

	return nil
}

// 计算出不同的Pod，返回需要删除的Pod和需要添加的Pod
func (p *plegManager) calculateDiffPods(runtimePodStatus RunTimePodStatusMap, cachePods CachePodsMap) {
	// 需要删除的Pod
	deletePods := make([]string, 0)
	// 需要添加的Pod
	addPods := make([]string, 0)

	// 遍历runtimePodStatus，找到需要删除的Pod
	for podID := range runtimePodStatus {
		_, ok := cachePods[podID]
		if !ok {
			deletePods = append(deletePods, podID)
		}
	}

	// 遍历cachePods，找到需要添加的Pod
	for podID := range cachePods {
		_, ok := runtimePodStatus[podID]
		if !ok {
			addPods = append(addPods, podID)
		}
	}

	// 生成pleg事件
	// 把需要删除的Pod添加到pleg事件中
	for _, podID := range deletePods {
		p.AddContainerNeedDeleteEvent(podID)
	}

	// 把需要添加的Pod添加到pleg事件中
	for _, podID := range addPods {
		p.AddContainerNeedCreateEvent(podID, cachePods[podID])
	}
}

// 比较在Cache里面的一个Pod的情况和该Pod在运行时候的情况，然后生成pleg事件
func (p *plegManager) comparePodStatus(runtimePodStatus *runtime.RunTimePodStatus, cachePod *apiObject.PodStore) {
	//
}
