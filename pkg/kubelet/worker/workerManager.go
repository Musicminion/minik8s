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
	DeletePod(pod *apiObject.PodStore) error
	StartPod(pod *apiObject.PodStore) error
	StopPod(pod *apiObject.PodStore) error
	RestartPod(pod *apiObject.PodStore) error
	DelPodByPodID(podUUID string) error
	RecreatePodContainer(pod *apiObject.PodStore) error
	ExecPodContainer(pod *apiObject.PodStore, cmd []string) (string, error)
}

type podWorkerManager struct {
	// podUUID到PodWorker的映射
	PodWorkersMap map[string]*PodWorker

	// Worker的针对不同事件的处理函数
	AddPodHandler               func(pod *apiObject.PodStore) error
	DelPodHandler               func(pod *apiObject.PodStore) error
	StartPodHandler             func(pod *apiObject.PodStore) error
	StopPodHandler              func(pod *apiObject.PodStore) error
	RestartPodHandler           func(pod *apiObject.PodStore) error
	DelPodByIDHandler           func(podUUID string) error
	RecreatePodContainerHandler func(pod *apiObject.PodStore) error
	ExecPodHandler			  func(pod *apiObject.PodStore, cmd []string) (string, error)
}

func NewPodWorkerManager() PodWorkerManager {
	restorePodWorkersMap := make(map[string]*PodWorker)

	// 从runtimeManager中获取所有的pod
	podStatus, err := runtimeManager.GetRuntimeAllPodStatus()

	if err != nil {
		k8log.ErrorLog("Pod Worker Manager", "get all pod status error, error is "+err.Error())
		panic(err)
	}

	// 遍历所有的pod，创建PodWorker
	for podID, _ := range podStatus {
		restorePodWorkersMap[podID] = NewPodWorker()
		go restorePodWorkersMap[podID].Run()
	}

	return &podWorkerManager{
		PodWorkersMap:               restorePodWorkersMap,
		AddPodHandler:               runtimeManager.CreatePod,
		DelPodHandler:               runtimeManager.DeletePod,
		StartPodHandler:             runtimeManager.StartPod,
		StopPodHandler:              runtimeManager.StopPod,
		RestartPodHandler:           runtimeManager.RestartPod,
		DelPodByIDHandler:           runtimeManager.DelPodByPodID,
		RecreatePodContainerHandler: runtimeManager.RecreatePodContainer,
		ExecPodHandler:              runtimeManager.ExecPodContainer,
	}
}

// AddPod 添加pod
func (p *podWorkerManager) AddPod(podStore *apiObject.PodStore) error {
	k8log.InfoLog("Pod Worker", "add pod, pod name is "+podStore.GetPodName())
	podUUID := podStore.GetPodUUID()
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
			Pod: podStore,
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

	k8log.DebugLog("Pod Worker", "delete pod, task type is "+string(task.TaskType))

	// 把任务添加到PodWorker的任务队列中
	err := p.PodWorkersMap[podUUID].AddTask(task)
	time.Sleep(1 * time.Second)
	if err != nil {
		return err
	}

	// 停止PodWorker
	p.PodWorkersMap[podUUID].Stop()
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

// 根据pod的UUID删除pod，用来处理pleg的删除事件
func (p *podWorkerManager) DelPodByPodID(podUUID string) error {
	// 遍历PodWorkersMap，如果不存在podUUID对应的PodWorker，则直接返回
	if _, ok := p.PodWorkersMap[podUUID]; !ok {
		return errors.New("pod not exists")
	}

	// 创建任务
	task := WorkTask{
		TaskType: Task_DelPodByPodID,
		TaskArgs: Task_DelPodByPodIDArgs{
			PodUUID: podUUID,
		},
	}

	// 把任务添加到PodWorker的任务队列中
	err := p.PodWorkersMap[podUUID].AddTask(task)
	time.Sleep(1 * time.Second)
	if err != nil {
		return err
	}

	// 停止podWorker
	p.PodWorkersMap[podUUID].Stop()
	// 删除podWorkerMap
	delete(p.PodWorkersMap, podUUID)

	return nil
}

// 重建pod的容器
func (p *podWorkerManager) RecreatePodContainer(pod *apiObject.PodStore) error {
	podUUID := pod.GetPodUUID()
	// 遍历PodWorkersMap，如果不存在podUUID对应的PodWorker，则直接返回
	if _, ok := p.PodWorkersMap[podUUID]; !ok {
		return errors.New("pod not exists")
	}

	// 创建任务
	task := WorkTask{
		TaskType: Task_RecreatePodContainer,
		TaskArgs: Task_RecreatePodContainerArgs{
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

func (p *podWorkerManager) ExecPodContainer(pod *apiObject.PodStore, cmd []string) (string, error) {
	podUUID := pod.GetPodUUID()
	// 遍历PodWorkersMap，如果不存在podUUID对应的PodWorker，则直接返回
	if _, ok := p.PodWorkersMap[podUUID]; !ok {
		return "", errors.New("pod not exists")
	}

	// 创建任务
	task := WorkTask{
		TaskType: Task_ExecPod,
		TaskArgs: Task_ExecPodArgs{
			Pod: pod,
			Cmd: cmd,
		},
	}

	// 把任务添加到PodWorker的任务队列中
	err := p.PodWorkersMap[podUUID].AddTask(task)
	time.Sleep(1 * time.Second)
	if err != nil {
		return "", err
	}

	return "", nil
}

// ************************************************************
// 这里写PodWorker的函数
