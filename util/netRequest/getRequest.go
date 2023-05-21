package netrequest

import (
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
)

// Get请求
// GetRequest 从指定的uri获取数据，并将数据反序列化到target指向的对象中
func GetRequestByTarget(uri string, target interface{}, key string) (int, error) {
	code, res, err := GetRequest(uri)

	if err != nil {
		// k8log.ErrorLog("netrequest", "GetRequestByTarget failed, for get failed, err: "+err.Error())
		return code, err
	}

	if code != http.StatusOK {
		// k8log.ErrorLog("netrequest", "GetRequestByTarget failed, code: "+fmt.Sprint(code))
		return code, err
	}

	// 尝试在res中获取key对应的值
	data, ok := res[key]
	if !ok {
		// k8log.ErrorLog("netrequest", "GetRequestByTarget failed, for get key failed, key: "+key)
		return code, err
	}

	// 如果data为nil，直接返回
	if data == nil {
		return code, errors.New("resp[key] is nil")
	}

	// 将data转化为字符串
	dataStr := fmt.Sprint(data)

	// k8log.DebugLog("netrequest", "GetRequestByTarget dataStr: "+dataStr)

	// 将dataStr反序列化到target指向的对象中
	err = json.Unmarshal([]byte(dataStr), target)

	if err != nil {
		// k8log.ErrorLog("netrequest", "GetRequestByTarget failed, for decode failed, err: "+err.Error()+dataStr)
		return 0, err
	}

	return code, nil

	// k8log.DebugLog("netrequest", "GetRequestByTarget uri: "+uri)
	// k8log.DebugLog("netrequest", "target type: "+fmt.Sprint(reflect.TypeOf(target)))

	// response, err := http.Get(uri)
	// if err != nil {
	// 	k8log.ErrorLog("netrequest", "GetRequestByTarget failed, for get failed, err: "+err.Error())
	// 	return 0, err
	// }
	// defer response.Body.Close()

	// // decoder := json.NewDecoder(response.Body)
	// // err = decoder.Decode(target)
	// body, err := ioutil.ReadAll(response.Body)

	// k8log.DebugLog("netrequest", "GetRequestByTarget body: "+string(body))

	// if err != nil {
	// 	k8log.ErrorLog("netrequest", "GetRequestByTarget failed, for read body failed, err: "+err.Error())
	// 	return 0, err
	// }

	// err = json.Unmarshal(body, &target)
	// if err != nil {
	// 	k8log.ErrorLog("netrequest", "GetRequestByTarget failed, for decode failed, err: "+err.Error())
	// 	return 0, err
	// }

	// return response.StatusCode, nil
}

func GetRequest(uri string) (int, map[string]interface{}, error) {
	response, err := http.Get(uri)
	if err != nil {
		// k8log.ErrorLog("netrequest", "GetRequestByTarget failed, for get failed, err: "+err.Error())
		return 0, nil, err
	}
	defer response.Body.Close()

	var result map[string]interface{}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&result)
	if err != nil {
		// k8log.ErrorLog("netrequest", "GetRequest failed, for decode failed, err: "+err.Error())
		return 0, nil, err
	}

	return response.StatusCode, result, nil
}
