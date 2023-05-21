package status

import (
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/kubelet/runtime"
	"miniK8s/util/executor"
	"miniK8s/util/rediscache"
)

// StatusManager 状态管理器功能介绍
// 1. 用于管理缓存，和Redis打交道
// 2. 和API Server打交道，发布自己的Node的状态信息，
// 3. 获取Pod的状态信息，发布Pod的状态信息

type StatusManager interface {
	// AddPodToCache 将Pod的存储对象添加到缓存中
	AddPodToCache(pod *apiObject.PodStore) error
	// GetPodFromCache 从缓存中获取Pod的存储对象
	GetPodFromCache(podUUID string) (*apiObject.PodStore, error)
	// GetAllPodFromCache 从缓存中获取所有Pod的存储对象
	GetAllPodFromCache() (map[string]*apiObject.PodStore, error)

	// DelPodFromCache 从缓存中删除Pod的存储对象
	DelPodFromCache(podUUID string) error
	// UpdatePodToCache 更新缓存中的Pod的存储对象
	UpdatePodToCache(pod *apiObject.PodStore) error

	// resetCache 重置缓存
	ResetCache() error

	// 获取运行时候的Pod的状态信息
	GetAllPodFromRuntime() (map[string]*runtime.RunTimePodStatus, error)

	// 注册和反注册的功能
	// 注册节点
	RegisterNode() error
	// 注销节点
	UnRegisterNode() error
	// 获取节点名称
	GetNodeName() string

	// Run 运行状态管理器，函数不会阻塞
	Run()
}

type statusManager struct {
	cache          rediscache.RedisCache
	runtimeManager runtime.RuntimeManager
	// apiserverURLPrefix API Server的URL前缀
	apiserverURLPrefix string
}

func NewStatusManager(apiserverURLPrefix string) StatusManager {
	return &statusManager{
		cache:              rediscache.NewRedisCache(CacheDBID_PodCache),
		runtimeManager:     runtime.NewRuntimeManager(),
		apiserverURLPrefix: apiserverURLPrefix,
	}
}

// ************************************************************
// 这里都是缓存的增删改查函数
func (s *statusManager) AddPodToCache(pod *apiObject.PodStore) error {
	k8log.DebugLog("Kubelet-StatusManager", "Add Pod To Cache")
	return s.cache.Put(pod.GetPodUUID(), pod)
}

// 查找到不存在的Pod，只会返回nil，不会返回error
func (s *statusManager) GetPodFromCache(podUUID string) (*apiObject.PodStore, error) {
	var parsedPod apiObject.PodStore
	res, err := s.cache.GetObject(podUUID, &parsedPod)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	return &parsedPod, nil
}

func (s *statusManager) DelPodFromCache(podUUID string) error {
	k8log.DebugLog("Kubelet-StatusManager", "Del Pod From Cache")
	return s.cache.Delete(podUUID)
}

func (s *statusManager) UpdatePodToCache(pod *apiObject.PodStore) error {
	return s.cache.Update(pod.GetPodUUID(), pod)
}

func (s *statusManager) GetAllPodFromCache() (map[string]*apiObject.PodStore, error) {
	var podObj apiObject.PodStore
	res, err := s.cache.GetAllObject(&podObj)

	if err != nil {
		return nil, err
	}

	var pods = make(map[string]*apiObject.PodStore)
	for k, v := range res {
		pods[k] = v.(*apiObject.PodStore)
	}

	return pods, nil
}

func (s *statusManager) ResetCache() error {
	return s.cache.InitCache()
}

// ************************************************************

func (s *statusManager) GetAllPodFromRuntime() (map[string]*runtime.RunTimePodStatus, error) {
	result, err := s.runtimeManager.GetRuntimeAllPodStatus()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// run 用于启动状态管理器，这个函数会启动一个goroutine，用于定时向API Server发送心跳
// 函数不会阻塞
func (s *statusManager) Run() {
	registerWrap := func() {
		k8log.DebugLog("Kubelet-StatusManager", "Send Node HeartBeat")
		res := s.PushNodeStatus()
		if res != nil {
			k8log.ErrorLog("Register Node Error: ", res.Error())
		}
	}

	pushPodStatusWrap := func() {
		k8log.DebugLog("Kubelet-StatusManager", "Push Pod Status")
		res := s.PushNodePodStatus()
		if res != nil {
			k8log.ErrorLog("Push Pod Status Error: ", res.Error())
		}
	}

	pullPodWrap := func() {
		k8log.DebugLog("Kubelet-StatusManager", "Pull Pod Status")
		_, res := s.PullNodeAllPods()
		if res != nil {
			k8log.ErrorLog("Pull Pod Status Error: ", res.Error())
		}
	}

	// Node心跳的协程
	go executor.Period(NodeHeartBeatDelay, NodeHeartBeatInterval, registerWrap, NodeHeartBeatLoop)

	// // Pod状态更新到API Server的协程
	// go executor.Period(PodStatusUpdateDelay, PodStatusUpdateInterval, pushPodStatusWrap, PodStatusUpdateLoop)

	// Pod最新数据拉取到缓存的协程
	go executor.Period(PodPullDelay, PodPullInterval, pullPodWrap, PodPullIfLoop)

	// Pod推送
	go executor.Period(PodPushDelay, PodPushInterval, pushPodStatusWrap, PodPushIfLoop)
}

func (s *statusManager) GetNodeName() string {
	return s.runtimeManager.GetRuntimeNodeName()
}
