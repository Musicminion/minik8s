package scheduler

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	netrequest "miniK8s/util/netRequest"
)

func (sch *Scheduler) GetAllNodes() (nodes []apiObject.NodeStore, err error) {
	// TODO
	uriPrefix := "http://" + sch.apiServerHost + ":" + fmt.Sprint(sch.apiServerPort)
	uri := uriPrefix + config.NodesURL
	// nodes = make([]apiObject.NodeStore, 0)
	// nodes = []apiObject.NodeStore{}
	var allNodes []apiObject.NodeStore
	code, err := netrequest.GetRequestByTarget(uri, &allNodes, "data")

	if err != nil {
		k8log.DebugLog("scheduler", "get all nodes failed "+err.Error())
		return nil, err
	}

	if code != 200 {
		k8log.DebugLog("scheduler", "get all nodes failed, code: "+fmt.Sprint(code))
		return nil, fmt.Errorf("get all nodes failed, code: %d", code)
	}

	return allNodes, nil
}
