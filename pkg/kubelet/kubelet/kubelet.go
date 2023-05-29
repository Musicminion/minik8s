package kubelet

import (
	"encoding/json"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/kubeletconfig"
	"miniK8s/pkg/kubelet/pleg"
	"miniK8s/pkg/kubelet/status"
	"miniK8s/pkg/kubelet/worker"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

type Kubelet struct {
	config        *kubeletconfig.KubeletConfig
	lw            *listwatcher.Listwatcher
	workManager   worker.PodWorkerManager
	statusManager status.StatusManager

	// plegManager用来管理pod的生命周期
	plegManager pleg.PlegManager
	// kubelet通过这个通道来接收plegManager发送的事件，然后发送给WorkManager
	plegChan chan *pleg.PodLifecycleEvent

	// 用来同步的
	wg sync.WaitGroup

	// 用来接收podUpdate消息的通道
	podUpdates chan *entity.PodUpdate
}

func NewKubelet(conf *kubeletconfig.KubeletConfig) (*Kubelet, error) {
	newlw, err := listwatcher.NewListWatcher(conf.LWConf)
	if err != nil {
		return nil, err
	}

	Kubelet_StatusManager := status.NewStatusManager(conf.APIServerURLPrefix)
	Kubelet_PlegChan := make(chan *pleg.PodLifecycleEvent)

	k := &Kubelet{
		config:        conf,
		lw:            newlw,
		workManager:   worker.NewPodWorkerManager(),
		statusManager: Kubelet_StatusManager,
		plegChan:      Kubelet_PlegChan,
		plegManager:   pleg.NewPlegManager(Kubelet_StatusManager, Kubelet_PlegChan),
		podUpdates:    make(chan *entity.PodUpdate, 20),
	}

	return k, nil
}

func (k *Kubelet) RegisterNode() {
	k8log.InfoLog("Kubelet", "Try to register node")
	registerResult := k.statusManager.RegisterNode()
	if registerResult != nil {
		k8log.ErrorLog("Kubelet", "Register node failed, for "+registerResult.Error())
		// 开一个协程，每隔一段时间重试一次
		k.wg.Add(1)
		go func() {
			for {
				registerResult := k.statusManager.RegisterNode()
				if registerResult == nil {
					k8log.InfoLog("Kubelet", "Register node success")
					break
				}
				time.Sleep(30 * time.Second)
				k8log.ErrorLog("Kubelet", "Register node failed, for "+registerResult.Error())
			}
			k.wg.Done()
		}()
	}
	k8log.InfoLog("Kubelet", "Register node success")
}

func (k *Kubelet) UnRegisterNode() {
	k8log.InfoLog("Kubelet", "Try to unregister node")
	unregisterResult := k.statusManager.UnRegisterNode()
	if unregisterResult != nil {
		k8log.ErrorLog("Kubelet", "Unregister node failed, for "+unregisterResult.Error())
	}
	k8log.InfoLog("Kubelet", "Unregister node success")
}

func (k *Kubelet) Run() {
	k8log.InfoLog("Kubelet", "Launch Kubelet")
	k.RegisterNode()

	// 创建一个通道来接收信号
	sigs := make(chan os.Signal, 10)
	// 注册一个信号接收函数，将接收到的信号发送到通道
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 启动所有的Manager
	// TODO: 开启pleg的时候，有时候会把新建的pod给删除，奇怪
	k.statusManager.Run()
	k.plegManager.Run()

	go k.ListenChan()

	// 监听 podUpdate 的消息队列

	listenTopic := message.PodUpdateWithNode(k.statusManager.GetNodeName())
	k8log.InfoLog("Kubelet", "Start to listen on "+listenTopic+" queue")
	go k.lw.WatchQueue_Block(listenTopic, k.HandlePodUpdate, make(chan struct{}))
	go k.SyncLoopIterationWrap()

	<-sigs
	k.UnRegisterNode()
}

func (k *Kubelet) HandlePodUpdate(msg amqp.Delivery) {
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	k8log.InfoLog("Kubelet", "HandlePodUpdate: receive message"+string(msg.Body))
	if err != nil {
		k8log.ErrorLog("Kubelet", "消息格式错误,无法转换为Message")
	}
	podUpdate := &entity.PodUpdate{}
	err = json.Unmarshal([]byte(parsedMsg.Content), podUpdate)
	if err != nil {
		k8log.ErrorLog("Kubelet", "HandlePodUpdate: failed to unmarshal")
		return
	}
	k.podUpdates <- podUpdate

}

func (k *Kubelet) syncLoopIteration(podUpdates <-chan *entity.PodUpdate) bool {
	k8log.InfoLog("Kubelet", "syncLoopIteration: Sync loop Iteration")
	podUpdate, ok := <-podUpdates
	if !ok {
		k8log.InfoLog("Kubelet", "syncLoopIteration: podUpdates channel closed")
		return false
	}
	k8log.InfoLog("Kubelet", "podUpdate.Action: "+podUpdate.Action)

	switch podUpdate.Action {
	case message.CREATE:
		err := k.workManager.AddPod(&podUpdate.PodTarget)
		if err != nil {
			k8log.ErrorLog("Kubelet", "syncLoopIteration: AddPod failed, for "+err.Error())
		}
		err = k.statusManager.AddPodToCache(&podUpdate.PodTarget)

		// 输出pod的信息
		k8log.WarnLog("Kubelet", "syncLoopIteration: PodInfo: "+podUpdate.PodTarget.Metadata.UUID+" "+podUpdate.PodTarget.Metadata.Name+" "+podUpdate.PodTarget.Metadata.Namespace)

		if err != nil {
			k8log.ErrorLog("Kubelet", "syncLoopIteration: AddPodToCache failed, for "+err.Error())
		}
	case message.UPDATE:
		err := k.workManager.DelPodByPodID(podUpdate.PodTarget.Metadata.UUID)
		if err != nil {
			k8log.ErrorLog("Kubelet", "syncLoopIteration: DelPodByPodID failed, for "+err.Error())
			break
		}
		err = k.workManager.AddPod(&podUpdate.PodTarget)
		if err != nil {
			k8log.ErrorLog("Kubelet", "syncLoopIteration: AddPod failed, for "+err.Error())
			break
		}
		err = k.statusManager.UpdatePodToCache(&podUpdate.PodTarget)
		if err != nil {
			k8log.ErrorLog("Kubelet", "syncLoopIteration: AddPodToCache failed, for "+err.Error())
			break
		}
	case message.DELETE:
		err := k.workManager.DelPodByPodID(podUpdate.PodTarget.Metadata.UUID)
		if err != nil {
			k8log.ErrorLog("Kubelet", "syncLoopIteration: DelPodByPodID failed, for "+err.Error())
			break
		}
		err = k.statusManager.DelPodFromCache(podUpdate.PodTarget.Metadata.UUID)
		if err != nil {
			k8log.ErrorLog("Kubelet", "syncLoopIteration: DelPodFromCache failed, for "+err.Error())
			break
		}
	case message.EXEC:
		_, err := k.workManager.ExecPodContainer(&podUpdate.PodTarget, podUpdate.Cmd)
		if err != nil {
			k8log.ErrorLog("Kubelet", "syncLoopIteration: ExecPodContainer failed, for "+err.Error())
			break
		}
	}
	return true
}

func (k *Kubelet) SyncLoopIterationWrap() {
	for k.syncLoopIteration(k.podUpdates) {
	}
}
