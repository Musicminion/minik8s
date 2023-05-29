package trigger

import (
	"math/rand"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/k8log"
	netrequest "miniK8s/util/netRequest"
)

func Trigger(x int, y int) (res string, err error) {
	pods := []apiObject.PodStore{}
	n := len(pods)
	if n == 0 {
		return "", err
	}

	randomId := rand.Intn(n)
	pod := pods[randomId]
	namespace :=pod.Metadata.Labels
	URL := "http://localhost:28080/"+

	params := map[string]interface{}{
		"x": x,
		"y": y,
	}

	// targetURL := config.NodesURL
	// targetURL = s.apiserverURLPrefix + targetURL

	// 发送POST请求
	_, _, err = netrequest.PostRequestByTarget(URL, params)

	if err != nil {
		k8log.ErrorLog("Trigger", "Run: failed to post request"+err.Error())
		return
	}

	k8log.InfoLog("Function", "Trigger: success to trigger")

	return "", err
}
