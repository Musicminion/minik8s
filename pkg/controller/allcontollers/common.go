package allcontollers

import (
	"errors"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	netrequest "miniK8s/util/netRequest"
	"net/http"
)

func GetAllPodFromAPIServer() ([]apiObject.PodStore, error) {
	url := config.GetAPIServerURLPrefix() + config.GlobalPodsURL

	allPods := make([]apiObject.PodStore, 0)

	code, err := netrequest.GetRequestByTarget(url, &allPods, "data")

	if err != nil {
		return nil, err
	}

	if code != http.StatusOK {
		return nil, errors.New("get all pods from apiserver failed")
	}

	return allPods, nil
}

func CheckIfPodMeetRequirement(pod *apiObject.PodStore, selectors map[string]string) bool {
	// 这里的匹配策略是：只要pod的label中有一个key-value对与selector中的key-value对相同，就认为pod满足要求
	podLabel := pod.Metadata.Labels
	for key, value := range selectors {
		// if podLabel[key] == value {
		// 	return true
		// } else {
		// 	continue
		// }
		if podLabel[key] != value {
			return false
		} else {
			continue
		}
	}

	return true
}

func GetAllNodesName() []string {
	URL := config.GetAPIServerURLPrefix() + config.NodesURL
	allNodes := make([]apiObject.NodeStore, 0)
	code, err := netrequest.GetRequestByTarget(URL, &allNodes, "data")
	if err != nil {
		return nil
	}
	if code != http.StatusOK {
		return nil
	}
	allNodesName := make([]string, 0)
	for _, node := range allNodes {
		allNodesName = append(allNodesName, node.GetName())
	}

	return allNodesName
}
