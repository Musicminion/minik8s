package status

import (
	"encoding/json"
	"errors"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
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

// NodeAllPodsURL = "/api/v1/nodes/:name/pods"
func (s *statusManager) PullNodeAllPods() ([]*apiObject.PodStore, error) {
	// TODO: 从APIServer拉取Pod的状态信息
	return nil, nil
}

func (s *statusManager) PushNodePodStatus([]*apiObject.PodStore) error {
	// TODO: 向APIServer推送Pod的状态信息
	return nil
}
