package status

import "miniK8s/pkg/apiObject"

// 这个文件主要存放和APIServer打交道的函数
// 注意包含从APIServer Pull数据和Push数据的函数

func (s *statusManager) PushNodeStatus() (*apiObject.NodeStatus, error) {
	// TODO: 向APIServer推送Node的状态信息
	return nil, nil
}

func (s *statusManager) PullNodePodStatus() (*apiObject.PodStatus, error) {
	// TODO: 从APIServer拉取Pod的状态信息
	return nil, nil
}

func (s *statusManager) PushNodePodStatus() (*apiObject.PodStatus, error) {
	// TODO: 向APIServer推送Pod的状态信息
	return nil, nil
}