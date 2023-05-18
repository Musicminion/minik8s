package pleg

import (
	"miniK8s/pkg/kubelet/runtime"
)

// Pleg定义的是Pod的生命周期的事件类型
type PodLifeCycleEventType string

const (
	// @Reminder
	// 本来我想用容器作为最小化的调度，但是后来觉得K8s的调度最小是Pod，所以干脆索性遇到问题重启Pod的方式
	// 因为我考虑到如果因为一个容器崩溃，可能导致其他容器的某些状态出问题，所以不如发现一个容器崩溃，直接重启Pod，也就是把Pod当做最小的调度单元
	// 但是这些字段我还是保留在这了，以后在做考虑
	ContainerNeedStart   PodLifeCycleEventType = "ContainerNeedStart"   // 容器需要启动
	ContainerNeedCreate  PodLifeCycleEventType = "ContainerNeedCreate"  // 容器需要创建
	ContainerNeedReStart PodLifeCycleEventType = "ContainerNeedReStart" // 容器需要重新启动
	ContainerNeedStop    PodLifeCycleEventType = "ContainerNeedStop"    // 容器需要停止
	ContainerNeedDelete  PodLifeCycleEventType = "ContainerNeedDelete"  // 容器需要删除
	ContainerNeedUpdate  PodLifeCycleEventType = "ContainerNeedUpdate"  // 容器需要更新

	PodNeedStart   PodLifeCycleEventType = "PodNeedStart"   // Pod需要启动
	PodNeedStop    PodLifeCycleEventType = "PodNeedStop"    // Pod需要停止
	PodNeedRestart PodLifeCycleEventType = "PodNeedRestart" // Pod需要重启
	PodNeedCreate  PodLifeCycleEventType = "PodNeedCreate"  // Pod需要创建
	PodNeedDelete  PodLifeCycleEventType = "PodNeedDelete"  // Pod需要删除
	PodNeedUpdate  PodLifeCycleEventType = "PodNeedUpdate"  // Pod需要更新

	PodContainerNeedRecreate PodLifeCycleEventType = "PodContainerNeedRecreate" // Pod的容器需要重新创建

	// 不是上面的任何一个事件，就是PodSync
	PodSync PodLifeCycleEventType = "PodSync"
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

// // ContainerStarted - event type when the new state of container is running.
// // 容器新的状态是启动的
// ContainerStarted PodLifeCycleEventType = "ContainerStarted"

// // ContainerDied - event type when the new state of container is exited.
// // 容器新的状态是退出的
// ContainerDied PodLifeCycleEventType = "ContainerDied"

// // ContainerRemoved - event type when the old state of container is exited.
// // 容器旧的状态是退出的
// ContainerRemoved PodLifeCycleEventType = "ContainerRemoved"

// // PodSync is used to trigger syncing of a pod when the observed change of
// // the state of the pod cannot be captured by any single event above.
// // 不是上面的任何一个事件，就是PodSync
// PodSync PodLifeCycleEventType = "PodSync"

// // ContainerChanged - event type when the new state of container is unknown.
// // 容器新的状态是未知的
// ContainerChanged PodLifeCycleEventType = "ContainerChanged"
