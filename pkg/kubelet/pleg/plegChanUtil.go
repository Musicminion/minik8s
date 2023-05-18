package pleg

import "miniK8s/pkg/apiObject"

// 这个包里面主要提供一些工具函数，用来向plegChannel里面添加事件
// 不同的事件通过不同的函数添加。不要直接添加事件，注意传递的参数
func (p *plegManager) AddPodNeedStartEvent(podID string) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: PodNeedStart,
		Data: nil,
	}
}

func (p *plegManager) AddPodNeedStopEvent(podID string) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: PodNeedStop,
		Data: nil,
	}
}

func (p *plegManager) AddPodNeedDeleteEvent(podID string) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: PodNeedDelete,
		Data: nil,
	}
}

func (p *plegManager) AddPodNeedCreateEvent(podID string, podData *apiObject.PodStore) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: PodNeedCreate,
		Data: podData,
	}
}

func (p *plegManager) AddPodNeedUpdateEvent(podID string, podData *apiObject.PodStore) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: ContainerNeedUpdate,
		Data: podData,
	}
}

func (p *plegManager) AddPodNeedRestartEvent(podID string) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: ContainerNeedReStart,
		Data: nil,
	}
}

func (p *plegManager) AddContainerNeedSyncEvent(podID string) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: PodSync,
		Data: nil,
	}
}

func (p *plegManager) AddPodContainerNeedRecreateEvent(podID string, podData *apiObject.PodStore) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: PodContainerNeedRecreate,
		Data: podData,
	}
}
