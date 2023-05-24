package pleg

import (
	"fmt"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/runtime"
	"miniK8s/pkg/kubelet/status"
	"miniK8s/util/executor"
)

type PlegManager interface {
	// Run 运行plegManager
	Run()
}

type plegManager struct {
	// 这个变量由kubelet创建，然后传递给plegManager
	PlegChannel chan *PodLifecycleEvent
	// statusManager用来获取Pod的状态信息
	statusManager status.StatusManager
	// podStatus是一个Pod的UUID到PodRecord的映射
	podStatus podRecords
}

// 创建PlegManager的时候，必须要传递一个statusManager，以及PlegChannel
func NewPlegManager(statusManager status.StatusManager, plegchan chan *PodLifecycleEvent) PlegManager {
	return &plegManager{
		PlegChannel:   plegchan,
		statusManager: statusManager,
		podStatus:     make(podRecords),
	}
}

// ************************************************************
// 这里都是podStatus的增删改查函数
func (p *plegManager) UpdatePodRecord(podID string, newStatus *runtime.RunTimePodStatus) error {
	// // 遍历podStatus，找到podID对应的podRecord
	// 如果在podStatus里面存在podID对应的podRecord，就更新podRecord
	tryFindPodRecord, ok := p.podStatus[podID]
	if ok && tryFindPodRecord != nil {
		tryFindPodRecord.old = tryFindPodRecord.current
		tryFindPodRecord.current = newStatus

		// 如果podRecord的old和current都是nil，就删除这个podRecord，回收垃圾
		if tryFindPodRecord.old == nil && tryFindPodRecord.current == nil {
			delete(p.podStatus, podID)
		}
		return nil
	}

	// 如果没有找到，就创建一个新的podRecord
	p.podStatus[podID] = &podRecord{
		old:     nil,
		current: newStatus,
	}
	return nil
}

func (p *plegManager) GetPodRecord(podID string) (*podRecord, error) {
	podRecord, ok := p.podStatus[podID]
	if !ok {
		return nil, nil
	}
	return podRecord, nil
}

func (p *plegManager) DeletePodRecord(podID string) error {
	delete(p.podStatus, podID)
	return nil
}

func (p *plegManager) checkAllPod() error {
	// k8log.WarnLog("plegManager", "checkAllPod")

	// 获取运行时的Pod的状态
	runtimePodStatuses, err := p.statusManager.GetAllPodFromRuntime()

	if err != nil {
		return err
	}

	for _, podStatus := range runtimePodStatuses {
		k8log.DebugLog("plegManager", fmt.Sprintf("runtimePodStatuses is: %v", podStatus))
	}

	// 从缓存里面拿到所有的Pod的状态
	cachePods, err := p.statusManager.GetAllPodFromCache()

	if err != nil {
		return err
	}

	// for _, podStatus := range cachePods {
	// 	k8log.WarnLog("plegManager", fmt.Sprintf("cachePods is: : %v", podStatus))
	// }

	// 比较所有的缓存的Pod和运行时的Pod的状态，然后生成事件
	p.plegGenerator(runtimePodStatuses, cachePods)

	return nil
}

// ************************************************************
func (p *plegManager) Run() {
	k8log.DebugLog("plegManager", "plegManager Run")
	routineJob := func() {
		result := p.checkAllPod()
		if result != nil {
			logStr := fmt.Sprintf("plegManager checkAllPod error: %v", result)
			k8log.ErrorLog("kubelet-Pleg", logStr)
		}
	}

	// 每隔一段时间，就检查一次所有的Pod
	// 这个函数会阻塞在这里！
	go executor.Period(PlegFirstRunDelay, PlegRunPeriod, routineJob, PlegRunRoutine)

}
