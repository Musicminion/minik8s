package status

import (
	"encoding/json"
	"errors"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/apiserver/serverconfig"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
	"net/http"
)

// 这个文件主要存放和APIServer打交道的函数
// 注意包含从APIServer Pull数据和Push数据的函数

// NodeSpecStatusURL = "/api/v1/nodes/:name/status"
func (s *statusManager) PushNodeStatus() error {

	// TODO: 向APIServer推送Node的状态信息
	nodeStatus, err := s.runtimeManager.GetRuntimeNodeStatus()
	if err != nil {
		return err
	}

	// 获取Node的状态信息的URL
	targetURL := stringutil.Replace(config.NodeSpecStatusURL, config.URL_PARAM_NAME_PART, nodeStatus.Hostname)

	// 给targetURL添加前缀
	targetURL = s.apiserverURLPrefix + targetURL

	// 发送PUT请求
	code, res, err := netrequest.PutRequestByTarget(targetURL, nodeStatus)

	if err != nil {
		return err
	}

	if code != http.StatusOK {
		bodyBytes, err := json.Marshal(res)
		if err != nil {
			return err
		}
		return errors.New(string(bodyBytes))
	}
	return nil
}

// PodSpecStatusURL = "/api/v1/namespaces/:namespace/pods/:name/status"
// 更新Pod的状态信息，发送给APIServer
func (s *statusManager) PushNodePodStatus() error {

	// TODO: 向APIServer推送Pod的状态信息
	allPodStatus, err := s.runtimeManager.GetRuntimeAllPodStatus()
	if err != nil {
		return err
	}

	errorMsgAll := ""

	// 遍历allPodStatus
	for _, podStatus := range allPodStatus {
		curPodName := podStatus.PodName
		curPodNamespace := podStatus.PodNamespace

		// 获取Pod的状态信息的URL
		targetURL := s.apiserverURLPrefix + config.PodSpecStatusURL
		// 注意必须要先替换namespace，再替换name，不然替换短的会导致替换长的时候出现问题
		targetURL = stringutil.Replace(targetURL, config.URL_PARAM_NAMESPACE_PART, curPodNamespace)
		targetURL = stringutil.Replace(targetURL, config.URL_PARAM_NAME_PART, curPodName)

		// 发送POST请求
		code, res, err := netrequest.PostRequestByTarget(targetURL, podStatus.PodStatus)

		if err != nil {
			logStr := "Push Pod Status Error: " + err.Error()
			k8log.ErrorLog("kubelet", logStr)
			errorMsgAll += logStr
		}

		if code != http.StatusOK {
			bodyBytes, err := json.Marshal(res)
			if err != nil {
				logStr := "Parse Update Pod Status resp Error: " + err.Error()
				k8log.ErrorLog("kubelet", logStr)
			}

			logStr := "Update Pod Status Error: " + string(bodyBytes)
			k8log.ErrorLog("kubelet", logStr)

			errorMsgAll += logStr
		}
	}

	if errorMsgAll != "" {
		return errors.New(errorMsgAll)
	}

	return nil
}

// NodeAllPodsURL = "/api/v1/nodes/:name/pods"
func (s *statusManager) PullNodeAllPods() ([]apiObject.PodStore, error) {
	// TODO: 从APIServer拉取Pod的状态信息
	// 获取Node的状态信息的URL
	nodeName := s.runtimeManager.GetRuntimeNodeName()

	targetURL := stringutil.Replace(config.NodeAllPodsURL, config.URL_PARAM_NAME_PART, nodeName)
	targetURL = s.apiserverURLPrefix + targetURL

	var pods []apiObject.PodStore
	// 发送GET请求
	code, err := netrequest.GetRequestByTarget(targetURL, &pods, "data")

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, errors.New("pull node all pods failed")
	}

	// TODO: 这里需要做一些处理，比如将Pod的状态信息存储到本地

	// // 遍历pods，将Pod的状态信息存储到本地
	// for _, pod := range pods {
	// 	// 将Pod的状态信息存储到本地
	// 	s.UpdatePodToCache(&pod)
	// }
	// 先把拉取到的Pod的状态信息转化为map

	remotePodsMap := s.PodsArrayToMap(pods)

	// 然后条件性更新本地缓存，更新的规则是：如果本地缓存中没有这个Pod，就添加，如果远端的Pod没有这个Pod，就删除
	updateResult := s.UpdatePulledPodsToCache(remotePodsMap)

	if updateResult != nil {
		return pods, updateResult
	}

	return pods, nil
}

// 把一个Pod的数组转化为map[UUID]->Pod的映射
func (s *statusManager) PodsArrayToMap(pods []apiObject.PodStore) map[string]*apiObject.PodStore {
	podMap := make(map[string]*apiObject.PodStore)

	for _, pod := range pods {
		newPod := pod
		podMap[pod.GetPodUUID()] = &newPod
	}

	return podMap
}

func (s *statusManager) UpdatePulledPodsToCache(remotePodsMap map[string]*apiObject.PodStore) error {
	// 把本地缓存所有的Pod的状态信息全部拉出来
	localPodsMap, err := s.GetAllPodFromCache()

	if err != nil {
		return err
	}

	errorInfo := ""

	// 遍历localPodsMap，如果remotePodsMap中没有，就删除
	for uuid, localPod := range localPodsMap {
		if _, ok := remotePodsMap[uuid]; !ok {
			k8log.DebugLog("kubelet", "DelPodFromCache: "+uuid)
			result := s.DelPodFromCache(uuid)
			if result != nil {
				errorInfo += result.Error() + "\n"
			}
		} else {
			// 如果remotePodsMap中有，就比较两者的事件戳，如果remotePodsMap中的事件戳比较新，就更新
			remotePod := remotePodsMap[uuid]

			// remotePod.Status.UpdateTime > localPod.Status.UpdateTime
			if !remotePod.Status.UpdateTime.Before(localPod.Status.UpdateTime) {
				result := s.UpdatePodToCache(remotePod)
				if result != nil {
					errorInfo += result.Error() + "\n"
				}
			}
		}
	}

	// 遍历remotePodsMap，如果localPodsMap中没有，就添加
	for uuid, pod := range remotePodsMap {
		if _, ok := localPodsMap[uuid]; !ok {
			result := s.UpdatePodToCache(pod)
			if result != nil {
				errorInfo += result.Error() + "\n"
			}

		}
	}

	if errorInfo != "" {
		return errors.New(errorInfo)
	}

	return nil
}

// 这个函数用于向APIserver查询Node是否已经注册
// const config.NodeSpecURL untyped string = "/api/v1/nodes/:name"
func (s *statusManager) CheckIfRegisterd() bool {
	nodeName := s.runtimeManager.GetRuntimeNodeName()

	// 获取Node的状态信息的URL
	targetURL := stringutil.Replace(config.NodeSpecURL, config.URL_PARAM_NAME_PART, nodeName)

	// 拼接URL
	targetURL = s.apiserverURLPrefix + targetURL

	// 创建一个NodeStore对象
	var nodeStore apiObject.NodeStore

	// 发送GET请求
	code, err := netrequest.GetRequestByTarget(targetURL, &nodeStore, "data")

	if err != nil {
		k8log.DebugLog("kubelet", "CheckIfRegisterd Error, get data failed: "+err.Error())
		return false
	}

	if code == http.StatusOK {
		k8log.InfoLog("kubelet", "Node has been registered before")
		return true
	}

	k8log.InfoLog("kubelet", "Node has not been registered before, run register node")
	return false
}

// RegisterNode 和 UnRegisterNode 两个函数用来注册和注销Node到APIServer
// 注册的时候会检查是否已经注册，如果已经注册，则不需要再注册
// 反注册的时候，会告诉APIServer节点下线。这个函数需要在Kubelet生命周期结束的时候调用
// 这个函数用于注册Node到APIServer，只需要在Node初次启动的时候发起一次，即可
// API  "/api/v1/nodes"
func (s *statusManager) RegisterNode() error {

	// 检查是否已经注册
	if s.CheckIfRegisterd() {
		return nil
	}

	// 没有注册过，就执行全新的注册流程
	nodeIP, err := s.runtimeManager.GetRuntimeNodeIP()
	if err != nil {
		return err
	}

	// 组装一个Node的数据
	node := apiObject.Node{
		NodeBasic: apiObject.NodeBasic{
			APIVersion: serverconfig.APIVersion,
			Kind:       apiObject.NodeKind,
			NodeMetadata: apiObject.NodeMetadata{
				Name: s.runtimeManager.GetRuntimeNodeName(),
			},
		},
		IP: nodeIP,
	}

	targetURL := config.NodesURL
	targetURL = s.apiserverURLPrefix + targetURL

	// 发送POST请求
	code, res, err := netrequest.PostRequestByTarget(targetURL, node)

	if err != nil {
		// 打印日志
		logStr := "Register Node Error: " + err.Error()
		k8log.ErrorLog("kubelet", logStr)
	}

	if code != http.StatusCreated {
		bodyBytes, err := json.Marshal(res)
		if err != nil {
			return err
		}
		return errors.New(string(bodyBytes))
	}

	return nil
}

// 这个函数用于注销Node到APIServer，只需要在Node停止的时候调用一次，即可
// API  "/api/v1/nodes/:name"
// 函数实现是直接更新Node状态信息为不活跃
func (s *statusManager) UnRegisterNode() error {
	// 获取Node的状态信息的URL
	nodeStatus, err := s.runtimeManager.GetRuntimeNodeStatus()

	if err != nil {
		return err
	}

	// 设置Node的状态为不活跃
	nodeStatus.Condition = apiObject.NodeCondition(apiObject.Unknown)

	// 获取Node的状态信息的URL
	targetURL := stringutil.Replace(config.NodeSpecStatusURL, config.URL_PARAM_NAME_PART, nodeStatus.Hostname)

	// 补充URL前缀
	targetURL = s.apiserverURLPrefix + targetURL

	// 发送PUT请求
	code, res, err := netrequest.PutRequestByTarget(targetURL, nodeStatus)

	if err != nil {
		return err
	}

	if code != http.StatusOK {
		bodyBytes, err := json.Marshal(res)
		if err != nil {
			return err
		}
		return errors.New(string(bodyBytes))
	}

	return nil
}
