package status

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/kubelet/runtime"
	"miniK8s/util/rediscache"
)

type StatusManager interface {
	// AddPodToCache 将Pod的存储对象添加到缓存中
	AddPodToCache(pod *apiObject.PodStore) error
	// GetPodFromCache 从缓存中获取Pod的存储对象
	GetPodFromCache(podUUID string) (*apiObject.PodStore, error)
	// DelPodFromCache 从缓存中删除Pod的存储对象
	DelPodFromCache(podUUID string) error
	// UpdatePodToCache 更新缓存中的Pod的存储对象
	UpdatePodToCache(pod *apiObject.PodStore) error
}

type statusManager struct {
	cache          rediscache.RedisCache
	runtimeManager runtime.RuntimeManager
}

const (
	// 缓存在Redis里面的数据库的ID编号，用于区分不同的缓存数据库
	CacheDBID_PodCache = 0
)

func NewStatusManager() StatusManager {
	return &statusManager{
		cache:          rediscache.NewRedisCache(CacheDBID_PodCache),
		runtimeManager: runtime.NewRuntimeManager(),
	}
}

// ************************************************************
// 这里都是缓存的增删改查函数
func (s *statusManager) AddPodToCache(pod *apiObject.PodStore) error {
	return s.cache.Put(pod.GetPodUUID(), pod)
}

func (s *statusManager) GetPodFromCache(podUUID string) (*apiObject.PodStore, error) {
	var parsedPod apiObject.PodStore
	_, err := s.cache.GetObject(podUUID, &parsedPod)
	if err != nil {
		return nil, err
	}
	return &parsedPod, nil
}

func (s *statusManager) DelPodFromCache(podUUID string) error {
	return s.cache.Delete(podUUID)
}

func (s *statusManager) UpdatePodToCache(pod *apiObject.PodStore) error {
	return s.cache.Update(pod.GetPodUUID(), pod)
}

// ************************************************************

// func (s *statusManager) run() {

// }
