package pleg

import "miniK8s/pkg/apiObject"

// 这个包里面主要提供一些工具函数，用来向plegChannel里面添加事件
// 不同的事件通过不同的函数添加。不要直接添加事件，注意传递的参数
func (p *plegManager) AddContainerNeedStartEvent(podID string) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: ContainerNeedStart,
		Data: nil,
	}
}

func (p *plegManager) AddContainerNeedStopEvent(podID string) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: ContainerNeedStop,
		Data: nil,
	}
}

func (p *plegManager) AddContainerNeedDeleteEvent(podID string) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: ContainerNeedDelete,
		Data: nil,
	}
}

func (p *plegManager) AddContainerNeedCreateEvent(podID string, podData *apiObject.PodStore) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: ContainerNeedCreate,
		Data: podData,
	}
}

func (p *plegManager) AddContainerNeedUpdateEvent(podID string, podData *apiObject.PodStore) {
	p.PlegChannel <- &PodLifecycleEvent{
		ID:   podID,
		Type: ContainerNeedUpdate,
		Data: podData,
	}
}

func (p *plegManager) AddContainerNeedRestartEvent(podID string) {
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
