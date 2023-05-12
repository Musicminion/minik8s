package status

import (
	"encoding/json"
	"errors"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/stringutil"
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
	targetURL := stringutil.Replace(config.NodeSpecStatusURL, config.URI_PARAM_NAME_PART, nodeStatus.Hostname)

	// 发送PUT请求
	code, res, err := netrequest.PutRequestByTarget(targetURL, nodeStatus)

	if err != nil {
		return err
	}

	if code != 200 {
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
func (s *statusManager) PushNodePodStatus() {
	// TODO: 向APIServer推送Pod的状态信息
	allPodStatus, allPodToName, allPodToNamespace, err := s.runtimeManager.GetRuntimeAllPodStatus()
	if err != nil {
		return
	}

	// 遍历allPod
	for podUUID, podStatus := range allPodStatus {
		curPodName := allPodToName[podUUID]
		curPodNamespace := allPodToNamespace[podUUID]

		// 获取Pod的状态信息的URL
		targetURL := stringutil.Replace(config.PodSpecStatusURL, config.URI_PARAM_NAME_PART, curPodName)
		targetURL = stringutil.Replace(targetURL, config.URL_PARAM_NAMESPACE_PART, curPodNamespace)

		// 发送PUT请求
		code, res, err := netrequest.PutRequestByTarget(targetURL, podStatus)

		if err != nil {
			logStr := "Push Pod Status Error: " + err.Error()
			k8log.ErrorLog("kubelet", logStr)
		}

		if code != 200 {
			bodyBytes, err := json.Marshal(res)
			if err != nil {
				logStr := "Parse Update Pod Status resp Error: " + err.Error()
				k8log.ErrorLog("kubelet", logStr)
			}

			logStr := "Update Pod Status Error: " + string(bodyBytes)
			k8log.ErrorLog("kubelet", logStr)
		}

	}

}

// NodeAllPodsURL = "/api/v1/nodes/:name/pods"
func (s *statusManager) PullNodeAllPods() ([]apiObject.PodStore, error) {
	// TODO: 从APIServer拉取Pod的状态信息
	// 获取Node的状态信息的URL
	nodeName := s.runtimeManager.GetRuntimeNodeName()

	targetURL := stringutil.Replace(config.NodeAllPodsURL, config.URI_PARAM_NAME_PART, nodeName)

	var pods []apiObject.PodStore
	// 发送GET请求
	code, err := netrequest.GetRequestByTarget(targetURL, &pods, "data")

	if err != nil {
		return nil, err
	}

	if code != 200 {
		return nil, errors.New("pull node all pods failed")
	}

	// TODO: 这里需要做一些处理，比如将Pod的状态信息存储到本地

	// 遍历pods，将Pod的状态信息存储到本地
	for _, pod := range pods {
		// 将Pod的状态信息存储到本地
		s.UpdatePodToCache(&pod)
	}

	return pods, nil
}
