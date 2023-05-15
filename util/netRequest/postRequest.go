package netrequest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"miniK8s/pkg/k8log"
)

// Post请求
func PostRequestByTarget(uri string, target interface{}) (int, interface{}, error) {
	jsonData, err := json.Marshal(target)
	if err != nil {
		k8log.ErrorLog("postRequest", "PostRequestByTarget: Marshal object failed "+err.Error())
		return 0, nil, err
	}
	response, err := http.Post(uri, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		k8log.ErrorLog("postRequest", "PostRequestByTarget: Post object failed "+err.Error())
		return 0, nil, err
	}
	defer response.Body.Close()

	var bodyJson interface{}
	if err := json.NewDecoder(response.Body).Decode(&bodyJson); err != nil {
		k8log.ErrorLog("postRequest", "PostRequestByTarget: Decode response failed "+err.Error())
		return 0, nil, err
	}

	return response.StatusCode, bodyJson, nil
}
