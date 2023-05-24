package kubelet

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/kubelet/pleg"
)

func (k *Kubelet) ListenChan() {
	// k.plegChan监听到事件之后，会把事件发送给k.workManager
	// k.workManager会根据事件的类型，调用不同的函数

	for event := range k.plegChan {
		switch event.Type {
		case pleg.PodNeedCreate:
			k.PlegPodNeedCreateHandler(event)
		case pleg.PodNeedUpdate:
			k.PlegPodNeedUpdateHandler(event)
		case pleg.PodNeedStart:
			k.PlegPodNeedStartHandler(event)
		case pleg.PodNeedStop:
			k.PlegPodNeedStopHandler(event)
		case pleg.PodNeedDelete:
			k.PlegPodNeedDeleteHandler(event)
		case pleg.PodNeedRestart:
			k.PlegPodNeedRestartHandler(event)
		case pleg.PodContainerNeedRecreate:
			k.PlegPodContainerNeedRecreateHandler(event)
		case pleg.PodSync:
			k.PlegPodSyncHandler(event)

		}
	}
}

func (k *Kubelet) PlegPodNeedCreateHandler(event *pleg.PodLifecycleEvent) {
	// 把data解析为pod对象
	podData := event.Data.(*apiObject.PodStore)

	// 把pod对象添加到workManager的podStore中
	k.workManager.AddPod(podData)
}

func (k *Kubelet) PlegPodNeedUpdateHandler(event *pleg.PodLifecycleEvent) {
	// 把data解析为pod对象
	// podData := event.Data.(*apiObject.PodStore)
	// TODO
}

func (k *Kubelet) PlegPodNeedStartHandler(event *pleg.PodLifecycleEvent) {
	// TODO
	// 把data解析为pod对象
	podData := event.Data.(*apiObject.PodStore)

	// 把pod对象添加到workManager的podStore中
	k.workManager.StartPod(podData)
}

func (k *Kubelet) PlegPodNeedStopHandler(event *pleg.PodLifecycleEvent) {
	// 把data解析为pod对象
	podData := event.Data.(*apiObject.PodStore)

	// 把pod对象添加到workManager的podStore中
	k.workManager.StopPod(podData)
}

func (k *Kubelet) PlegPodNeedDeleteHandler(event *pleg.PodLifecycleEvent) {
	// TODO

	// 把pod对象添加到workManager的podStore中
	k.workManager.DelPodByPodID(event.ID)
}

func (k *Kubelet) PlegPodNeedRestartHandler(event *pleg.PodLifecycleEvent) {
	// TODO
	// 把data解析为pod对象
	podData := event.Data.(*apiObject.PodStore)

	// 把pod对象添加到workManager的podStore中
	k.workManager.RestartPod(podData)
}

func (k *Kubelet) PlegPodContainerNeedRecreateHandler(event *pleg.PodLifecycleEvent) {
	// TODO
	// 把data解析为pod对象
	podData := event.Data.(*apiObject.PodStore)
	k.workManager.RecreatePodContainer(podData)
}

func (k *Kubelet) PlegPodSyncHandler(event *pleg.PodLifecycleEvent) {
	// TODO

}
