package netrequest

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Post请求
func PostRequestByTarget(uri string, target interface{}) (int, interface{}, error) {
	jsonData, err := json.Marshal(target)
	if err != nil {
		return 0, nil, err
	}

	response, err := http.Post(uri, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()

	var bodyJson interface{}
	if err := json.NewDecoder(response.Body).Decode(&bodyJson); err != nil {
		return 0, nil, err
	}

	return response.StatusCode, bodyJson, nil
}
