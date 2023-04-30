package netrequest

import (
	"net/http"
)

// 对于指定的端点发送一个DELETE请求
func DelRequest(uri string) (int, error) {
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return 0, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
