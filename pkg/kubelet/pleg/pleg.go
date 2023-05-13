package pleg

// Pleg定义的是Pod的生命周期的事件类型
type PodLifeCycleEventType string

const (
	// ContainerStarted - event type when the new state of container is running.
	ContainerStarted PodLifeCycleEventType = "ContainerStarted"
	// ContainerDied - event type when the new state of container is exited.
	ContainerDied PodLifeCycleEventType = "ContainerDied"
	// ContainerRemoved - event type when the old state of container is exited.
	ContainerRemoved PodLifeCycleEventType = "ContainerRemoved"
	// PodSync is used to trigger syncing of a pod when the observed change of
	// the state of the pod cannot be captured by any single event above.
	PodSync PodLifeCycleEventType = "PodSync"
	// ContainerChanged - event type when the new state of container is unknown.
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
// type podRecord struct {
// 	old     *apiObject.PodStore
// 	current *apiObject.PodStore
// }

// // podRecords是一个Pod的UUID到PodRecord的映射
// type podRecords map[string]*podRecord
