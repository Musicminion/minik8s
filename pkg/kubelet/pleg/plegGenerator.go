package pleg

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/runtime"
	minik8stypes "miniK8s/pkg/minik8sTypes"
)

// 用来比较运行的状态和缓存的状态，然后生成pleg事件

type RunTimePodStatusMap map[string]*runtime.RunTimePodStatus
type CachePodsMap map[string]*apiObject.PodStore

// 更新Pleg里面的缓存
// 这里涉及到三者的关系，一个是运行时的状态，一个是缓存的Pod，一个是pleg里面的podRecord
// podRecord是一个和状态强相关的结构体，所以更新依据：runtimePodStatus-> podRecord
// 然后我们的比较逻辑是podRecord<->cachePods，比较这两个的情况，然后生成pleg事件
func (p *plegManager) updatePlegRecord(runtimePodStatus RunTimePodStatusMap) error {
	// 对于所有的runtimePodStatus，更新plegRecord
	errStr := ""
	for podID, runtimePodStatus := range runtimePodStatus {
		err := p.UpdatePodRecord(podID, runtimePodStatus)

		if err != nil {
			errStr += fmt.Sprintf("update podRecord error: %s", err.Error())
		}
	}

	// 对于所有的p.podStatus，如果不在runtimePodStatus里面，就更新为nil
	for podID := range p.podStatus {
		_, ok := runtimePodStatus[podID]
		if !ok {
			err := p.UpdatePodRecord(podID, nil)
			if err != nil {
				errStr += fmt.Sprintf("delete podRecord error: %s", err.Error())
			}
		}
	}

	if errStr != "" {
		return fmt.Errorf(errStr)
	}

	return nil
}

// RuntimePodStatusMap 是PodID到运行时的Pod状态的映射
// CachePodsMap 是PodID到缓存的Pod的映射
// 1. pleg产生之前，首先我们根据运行时的状态，更新plegRecord
// 2. 然后我们比较plegRecord和cachePods，生成pleg事件，具体说来如下
// 3. 首先查找需要删除的Pod，遍历plegRecord，如果不在cachePods里面，就是需要删除的Pod
// 4. 然后查找需要添加的Pod，遍历cachePods，如果不在plegRecord里面，就是需要添加的Pod
// 5. 然后遍历plegRecord，查看发生了什么变化，然后对照cachePods，生成对应的事件
func (p *plegManager) plegGenerator(runtimePodStatus RunTimePodStatusMap, cachePods CachePodsMap) error {
	// 根据运行时状态更新plegRecord
	errStr := ""

	err := p.updatePlegRecord(runtimePodStatus)

	if err != nil {
		errStr += fmt.Sprintf("updatePlegRecord error: %s", err.Error())
	}

	// 然后比较plegRecord和cachePods，生成pleg事件
	// 先查找需要删除的Pod，遍历plegRecord，如果不在cachePods里面，就是需要删除的Pod
	for podID := range p.podStatus {
		_, ok := cachePods[podID]
		if !ok {
			k8log.InfoLog("plegManager", fmt.Sprintf("podID %s need delete", podID))

			p.AddPodNeedDeleteEvent(podID)
		}
	}

	// 然后查找需要添加的Pod，遍历cachePods，如果不在plegRecord里面，就是需要添加的Pod
	for podID := range cachePods {
		_, ok := p.podStatus[podID]
		if !ok {
			k8log.InfoLog("plegManager", fmt.Sprintf("podID %s need create", podID))
			p.AddPodNeedCreateEvent(podID, cachePods[podID])
		}
	}

	// 然后遍历plegRecord，查看发生了什么变化，然后对照cachePods，生成对应的事件
	// _是podID，podRecord是podID对应的podRecord
	for _, podRecord := range p.podStatus {
		if podRecord != nil && podRecord.old == nil && podRecord.current != nil {
			// 遍历podRecord.containers，查看发生了什么变化，然后对照cachePods，生成对应的事件
			for _, containerstatus := range podRecord.current.PodStatus.ContainerStatuses {
				switch containerstatus.Status {
				case string(minik8stypes.Created):
					// break
				case string(minik8stypes.Running):
					// break
				case string(minik8stypes.Paused):
					// break
					p.AddPodNeedStartEvent(podRecord.current.PodID)
				case string(minik8stypes.Restart):
					// break
				case string(minik8stypes.Removing):
					// break
				case string(minik8stypes.Exited):
					p.AddPodNeedRestartEvent(podRecord.current.PodID)
					// break
				case string(minik8stypes.Dead):
					p.AddPodNeedRestartEvent(podRecord.current.PodID)
					// break
				default:
					// break
				}
				// }
			}

			if podRecord != nil && podRecord.old != nil && podRecord.current != nil {
				// p.CompareOldAndCurrentPodStatus(podRecord.old, podRecord.current)
				// 如果发现获取到的pod的状态和上一次的状态数量都不一样，那么就是发生了变化
				if len(podRecord.old.PodStatus.ContainerStatuses) != len(podRecord.current.PodStatus.ContainerStatuses) {
					p.AddPodContainerNeedRecreateEvent(podRecord.current.PodID, cachePods[podRecord.current.PodID])
				}

				// 遍历podRecord.containers，查看发生了什么变化，然后对照cachePods，生成对应的事件
				for id, containerstatus := range podRecord.current.PodStatus.ContainerStatuses {
					if containerstatus.Status != podRecord.old.PodStatus.ContainerStatuses[id].Status {
						p.AddPodNeedRestartEvent(podRecord.current.PodID)
					}
				}
			}
		}
	}

	return nil
}

// func (p *plegManager) CompareOldAndCurrentPodStatus(oldStatus *runtime.RunTimePodStatus, newStatus *runtime.RunTimePodStatus) {
// 	// 如果状态没有发生变化
// 	if len(oldStatus.PodStatus.ContainerStatuses) != len(newStatus.PodStatus.ContainerStatuses) {

// 	}
// }

// // 计算出不同的Pod，返回需要删除的Pod和需要添加的Pod
// func (p *plegManager) calculateDiffPods(runtimePodStatus RunTimePodStatusMap, cachePods CachePodsMap) {
// 	// 需要删除的Pod
// 	deletePods := make([]string, 0)
// 	// 需要添加的Pod
// 	addPods := make([]string, 0)

// 	// 遍历runtimePodStatus，找到需要删除的Pod
// 	for podID := range runtimePodStatus {
// 		_, ok := cachePods[podID]
// 		if !ok {
// 			deletePods = append(deletePods, podID)
// 		}
// 	}

// 	// 遍历cachePods，找到需要添加的Pod
// 	for podID := range cachePods {
// 		_, ok := runtimePodStatus[podID]
// 		if !ok {
// 			addPods = append(addPods, podID)
// 		}
// 	}

// 	// 生成pleg事件
// 	// 把需要删除的Pod添加到pleg事件中
// 	for _, podID := range deletePods {
// 		p.AddContainerNeedDeleteEvent(podID)
// 	}

// 	// 把需要添加的Pod添加到pleg事件中
// 	for _, podID := range addPods {
// 		p.AddContainerNeedCreateEvent(podID, cachePods[podID])
// 	}
// }

// // 比较在Cache里面的一个Pod的情况和该Pod在运行时候的情况，然后生成pleg事件
// func (p *plegManager) comparePodStatus(runtimePodStatus *runtime.RunTimePodStatus, cachePod *apiObject.PodStore) {
// 	//
// }
