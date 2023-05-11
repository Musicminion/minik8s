package worker

import (
	"errors"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/kubelet/runtime"
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
	if err != nil {
		return err
	}

	return nil
}


