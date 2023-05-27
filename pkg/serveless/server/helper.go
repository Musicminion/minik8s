package server

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
