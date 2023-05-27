package scheduler

import (
	"fmt"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/k8log"
	netrequest "miniK8s/util/netRequest"
	"net/http"
)

func (sch *Scheduler) GetAllNodes() (nodes []apiObject.NodeStore, err error) {
	uriPrefix := config.GetAPIServerURLPrefix()
	uri := uriPrefix + config.NodesURL
	var allNodes []apiObject.NodeStore
	code, err := netrequest.GetRequestByTarget(uri, &allNodes, "data")

	if err != nil {
		k8log.ErrorLog("Scheduler", "get all nodes failed "+err.Error())
		return nil, err
	}

	if code != http.StatusOK {
		k8log.ErrorLog("Scheduler", "get all nodes failed, code: "+fmt.Sprint(code))
		return nil, fmt.Errorf("get all nodes failed, code: %d", code)
	}

	return allNodes, nil
}
