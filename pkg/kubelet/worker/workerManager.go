package worker

import (
	"errors"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/runtime"
	"time"
)

var runtimeManager = runtime.NewRuntimeManager()

type PodWorkerManager interface {
	AddPod(pod *apiObject.PodStore) error
}

type podWorkerManager struct {
	// podUUID到PodWorker的映射
	PodWorkersMap map[string]*PodWorker

	// Worker的针对不同事件的处理函数
	AddPodHandler     func(pod *apiObject.PodStore) error
	DelPodHandler     func(pod *apiObject.PodStore) error
	StartPodHandler   func(pod *apiObject.PodStore) error
	StopPodHandler    func(pod *apiObject.PodStore) error
	RestartPodHandler func(pod *apiObject.PodStore) error
}

func NewPodWorkerManager() *podWorkerManager {
	return &podWorkerManager{
		PodWorkersMap:     make(map[string]*PodWorker),
		AddPodHandler:     runtimeManager.CreatePod,
		DelPodHandler:     runtimeManager.DeletePod,
		StartPodHandler:   runtimeManager.StartPod,
		StopPodHandler:    runtimeManager.StopPod,
		RestartPodHandler: runtimeManager.RestartPod,
	}
}

// AddPod 添加pod
func (p *podWorkerManager) AddPod(pod *apiObject.PodStore) error {
	podUUID := pod.GetPodUUID()
	// 遍历PodWorkersMap，如果存在podUUID对应的PodWorker，则直接返回
	if _, ok := p.PodWorkersMap[podUUID]; ok {
		return errors.New("pod already exists")
	}

	// 创建PodWorker
	podWorker := NewPodWorker()
	p.PodWorkersMap[podUUID] = podWorker

	// 启动PodWorker
	go podWorker.Run()

	

	// 创建任务
	task := WorkTask{
		TaskType: Task_AddPod,
		TaskArgs: Task_AddPodArgs{
			Pod: pod,
		},
	}

	// 把任务添加到PodWorker的任务队列中
	err := podWorker.AddTask(task)

	time.Sleep(1 * time.Second)

	if err != nil {
		return err
	}


	return nil
}

// DeletePod 删除pod
func (p *podWorkerManager) DeletePod(pod *apiObject.PodStore) error {
	podUUID := pod.GetPodUUID()
	// 遍历PodWorkersMap，如果不存在podUUID对应的PodWorker，则直接返回``
	if _, ok := p.PodWorkersMap[podUUID]; !ok {
		return errors.New("pod not exists")
	}

	// 创建任务
	task := WorkTask{
		TaskType: Task_DelPod,
		TaskArgs: Task_DelPodArgs{
			Pod: pod,
		},
	}

	k8log.DebugLog("[Pod Worker]", "delete pod, task type is "+ string(task.TaskType))

	// 把任务添加到PodWorker的任务队列中
	err := p.PodWorkersMap[podUUID].AddTask(task)
	time.Sleep(1 * time.Second)
	if err != nil {
		return err
	}
	// 删除对应的podWorkersMap
	delete(p.PodWorkersMap, podUUID)

	return nil
}

// StartPod 启动pod
func (p *podWorkerManager) StartPod(pod *apiObject.PodStore) error {
	podUUID := pod.GetPodUUID()
	// 遍历PodWorkersMap，如果不存在podUUID对应的PodWorker，则直接返回
	if _, ok := p.PodWorkersMap[podUUID]; !ok {
		return errors.New("pod not exists")
	}

	// 创建任务
	task := WorkTask{
		TaskType: Task_Start,
		TaskArgs: Task_StartPodArgs{
			Pod: pod,
		},
	}

	// 把任务添加到PodWorker的任务队列中
	err := p.PodWorkersMap[podUUID].AddTask(task)
	time.Sleep(1 * time.Second)
	if err != nil {
		return err
	}

	return nil
}

// StopPod 停止pod
func (p *podWorkerManager) StopPod(pod *apiObject.PodStore) error {
	podUUID := pod.GetPodUUID()
	// 遍历PodWorkersMap，如果不存在podUUID对应的PodWorker，则直接返回
	if _, ok := p.PodWorkersMap[podUUID]; !ok {
		return errors.New("pod not exists")
	}

	// 创建任务
	task := WorkTask{
		TaskType: Task_Stop,
		TaskArgs: Task_StopPodArgs{
			Pod: pod,
		},
	}

	// 把任务添加到PodWorker的任务队列中
	err := p.PodWorkersMap[podUUID].AddTask(task)
	time.Sleep(1 * time.Second)
	if err != nil {
		return err
	}

	return nil
}

// RestartPod 重启pod
func (p *podWorkerManager) RestartPod(pod *apiObject.PodStore) error {
	podUUID := pod.GetPodUUID()
	// 遍历PodWorkersMap，如果不存在podUUID对应的PodWorker，则直接返回
	if _, ok := p.PodWorkersMap[podUUID]; !ok {
		return errors.New("pod not exists")
	}

	// 创建任务
	task := WorkTask{
		TaskType: Task_Restart,
		TaskArgs: Task_RestartPodArgs{
			Pod: pod,
		},
	}

	// 把任务添加到PodWorker的任务队列中
	err := p.PodWorkersMap[podUUID].AddTask(task)
	time.Sleep(1 * time.Second)
	if err != nil {
		return err
	}

	return nil
}



// ************************************************************
// 这里写PodWorker的函数
