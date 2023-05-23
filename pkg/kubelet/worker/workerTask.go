package worker

import "miniK8s/pkg/apiObject"

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

type Task_DelPodByPodIDArgs struct {
	// PodUUID
	PodUUID string
}

type Task_RecreatePodContainerArgs struct {
	// PodStore
	Pod *apiObject.PodStore
}


type Task_ExecPodArgs struct {
	// PodStore
	Pod *apiObject.PodStore
	Cmd []string
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
	Task_None          TypeOfTask = "none"
	Task_AddPod        TypeOfTask = "addPod"
	Task_DelPod        TypeOfTask = "delPod"
	Task_Start         TypeOfTask = "startPod"
	Task_Stop          TypeOfTask = "stopPod"
	Task_Restart       TypeOfTask = "restartPod"
	Task_DelPodByPodID TypeOfTask = "delPodByPodID"
	Task_RecreatePodContainer TypeOfTask = "recreatePodContainer"
	Task_ExecPod TypeOfTask = "execPod"
)
