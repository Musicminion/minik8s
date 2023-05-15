package worker

import "miniK8s/pkg/apiObject"

// TaskType 任务类型
// const (
// 	Task_None    = "none"
// 	Task_AddPod  = "addPod"
// 	Task_DelPod  = "delPod"
// 	Task_Start   = "startPod"
// 	Task_Stop    = "stopPod"
// 	Task_Restart = "restartPod"
// )

// 同意定义的任务的参数
// type WorkTaskArgs struct {
// 	// PodStore
// 	Pod *apiObject.PodStore
// }

type Task_DelPodArgs struct {
	// PodStore
	Pod *apiObject.PodStore
}

type Task_StartPodArgs struct {
	// PodStore
	Pod *apiObject.PodStore
}

type Task_StopPodArgs struct {
	// PodStore
	Pod *apiObject.PodStore
}

type Task_RestartPodArgs struct {
	// PodStore
	Pod *apiObject.PodStore
}

type Task_AddPodArgs struct {
	// PodStore
	Pod *apiObject.PodStore
}

// 对于一个PodWorker来说，它包含了任务
type WorkTask struct {
	// 任务类型
	TaskType TypeOfTask

	TaskArgs interface{}
}


// TypeOfTask 任务类型
type TypeOfTask string

const (
	Task_None    TypeOfTask = "none"
	Task_AddPod  TypeOfTask = "addPod"
	Task_DelPod  TypeOfTask = "delPod"
	Task_Start   TypeOfTask = "startPod"
	Task_Stop    TypeOfTask = "stopPod"
	Task_Restart TypeOfTask = "restartPod"
)
