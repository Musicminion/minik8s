package netrequest

import (
	"encoding/json"

	"net/http"
)

// Get请求
// GetRequest 从指定的uri获取数据，并将数据反序列化到target指向的对象中
func GetRequestByTarget(uri string, target interface{}) error {
	response, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(target)
	if err != nil {
		return err
	}

	return nil
}

func GetRequest(uri string) (map[string]interface{}, error) {
	response, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
