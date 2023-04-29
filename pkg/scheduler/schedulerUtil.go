package scheduler

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	netrequest "miniK8s/util/netRequest"
)

func (sch *Scheduler) GetAllNodes() (nodes []apiObject.NodeStore, err error) {
	// TODO
	uriPrefix := "http://" + sch.apiServerHost + ":" + fmt.Sprint(sch.apiServerPort)
	uri := uriPrefix + config.NodesURL
	nodes = make([]apiObject.NodeStore, 0)
	code, err := netrequest.GetRequestByTarget(uri, &nodes)

	if err != nil {
		return nil, err
	}

	if code != 200 {
		return nil, fmt.Errorf("get all nodes failed, code: %d", code)
	}

	return nodes, nil
}
