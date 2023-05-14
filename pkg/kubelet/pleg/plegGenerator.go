package pleg

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/kubelet/runtime"
)

// 用来比较运行的状态和缓存的状态，然后生成pleg事件

type RunTimePodStatusMap map[string]*runtime.RunTimePodStatus
type CachePodsMap map[string]*apiObject.PodStore

func (p *plegManager) plegGenerator(runtimePodStatus RunTimePodStatusMap, cachePods CachePodsMap) error {

	return nil
}

// 计算出不同的Pod，返回需要删除的Pod和需要添加的Pod
func (p *plegManager) calculateDiffPods(runtimePodStatus RunTimePodStatusMap, cachePods CachePodsMap) {
	return nil, nil, nil
}
