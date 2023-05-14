package pleg

import (
	"fmt"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/runtime"
	"miniK8s/pkg/kubelet/status"
	"miniK8s/util/executor"
)

// Pleg定义的是Pod的生命周期的事件类型
type PodLifeCycleEventType string

const (
	// ContainerStarted - event type when the new state of container is running.
	// 容器新的状态是启动的
	ContainerStarted PodLifeCycleEventType = "ContainerStarted"

	// ContainerDied - event type when the new state of container is exited.
	// 容器新的状态是退出的
	ContainerDied PodLifeCycleEventType = "ContainerDied"

	// ContainerRemoved - event type when the old state of container is exited.
	// 容器旧的状态是退出的
	ContainerRemoved PodLifeCycleEventType = "ContainerRemoved"

	// PodSync is used to trigger syncing of a pod when the observed change of
	// the state of the pod cannot be captured by any single event above.
	// 不是上面的任何一个事件，就是PodSync
	PodSync PodLifeCycleEventType = "PodSync"

	// ContainerChanged - event type when the new state of container is unknown.
	// 容器新的状态是未知的
	ContainerChanged PodLifeCycleEventType = "ContainerChanged"
)

type PodLifecycleEvent struct {
	// Pod的UUID
	ID string
	// 事件的类型
	Type PodLifeCycleEventType
	// 参数和数据，取决于事件的类型
	Data interface{}
}

// // podRecord是一个Pod的旧状态和新状态的记录
type podRecord struct {
	old     *runtime.RunTimePodStatus
	current *runtime.RunTimePodStatus
}

// podRecords是一个Pod的UUID到PodRecord的映射
type podRecords map[string]*podRecord

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

func NewPlegManager(statusManager status.StatusManager) PlegManager {
	return &plegManager{
		PlegChannel:   make(chan *PodLifecycleEvent, 100),
		statusManager: statusManager,
		podStatus:     make(podRecords),
	}
}

// ************************************************************
// 这里都是podStatus的增删改查函数
func (p *plegManager) UpdatePodRecord(podID string, newStatus *runtime.RunTimePodStatus) error {
	// 遍历podStatus，找到podID对应的podRecord
	for _, podRecord := range p.podStatus {
		if podRecord.old.PodID == podID {
			podRecord.old = podRecord.current
			podRecord.current = newStatus
			return nil
		}
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
	return nil
}

// ************************************************************
func (p *plegManager) Run() {
	routineJob := func() {
		result := p.checkAllPod()
		if result != nil {
			logStr := fmt.Sprintf("plegManager checkAllPod error: %v", result)
			k8log.ErrorLog("kubelet-Pleg", logStr)
		}
	}

	// 每隔一段时间，就检查一次所有的Pod
	// 这个函数会阻塞在这里！
	executor.Period(PlegFirstRunDelay, PlegRunPeriod, routineJob, PlegRunRoutine)

}
