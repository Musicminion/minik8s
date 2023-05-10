package worker

import (
	"errors"
	"miniK8s/pkg/apiObject"
)

// 隔壁的接口：RunTimeManager
// CreatePod(pod *apiObject.PodStore) error
// DeletePod(pod *apiObject.PodStore) error
// StartPod(pod *apiObject.PodStore) error
// StopPod(pod *apiObject.PodStore) error
// RestartPod(pod *apiObject.PodStore) error

type PodWorker struct {
	// 任务队列
	TaskQueue chan WorkTask

	// Worker的针对不同事件的处理函数
	AddPodHandler     func(pod *apiObject.PodStore) error
	DelPodHandler     func(pod *apiObject.PodStore) error
	StartPodHandler   func(pod *apiObject.PodStore) error
	StopPodHandler    func(pod *apiObject.PodStore) error
	RestartPodHandler func(pod *apiObject.PodStore) error
}

// NewPodWorker
func NewPodWorker() *PodWorker {

	return &PodWorker{
		TaskQueue:         make(chan WorkTask, WorkerChannelBufferSize),
		AddPodHandler:     runtimeManager.CreatePod,
		DelPodHandler:     runtimeManager.DeletePod,
		StartPodHandler:   runtimeManager.StartPod,
		StopPodHandler:    runtimeManager.StopPod,
		RestartPodHandler: runtimeManager.RestartPod,
	}
}

// Run 这是一个阻塞的函数，会一直运行
// 调用的时候需要放到goroutine中
func (p *PodWorker) Run() {
	// 当通道被关闭时，for循环会自动结束
	for task := range p.TaskQueue {
		p.RunTask(task)
	}
}

func (p *PodWorker) RunTask(task WorkTask) {
	switch task.TaskType {
	case Task_AddPod:
		// p.AddPodHandler(task.TaskArgs.Pod)
	case Task_DelPod:
		// p.DelPodHandler(task.TaskArgs.Pod)
	case Task_Start:
		// p.StartPodHandler(task.TaskArgs.Pod)
	case Task_Stop:
		// p.StopPodHandler(task.TaskArgs.Pod)
	case Task_Restart:
		// p.RestartPodHandler(task.TaskArgs.Pod)
	default:
		// log.Error("unknown task type")
	}
}

// Worker添加任务
func (p *PodWorker) AddTask(task WorkTask) error {
	// TODO: 这里需要考虑任务队列满的情况

	// 检查队列是否已经满了
	if len(p.TaskQueue) == WorkerChannelBufferSize {
		return errors.New("task queue is full")
	}

	p.TaskQueue <- task
	return nil
}
