package kubectlutil

import (
	"encoding/json"
	"fmt"
	"miniK8s/pkg/k8log"
	netrequest "miniK8s/util/netRequest"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// 把一个API文件对象从文件的数据流里面读取出来kind字段
// 返回的可能是Pod, Service, Deployment, Namespace, ConfigMap, Secret
func GetAPIObjectTypeFromYamlFile(fileContent []byte) (string, error) {
	var result map[string]interface{}

	err := yaml.Unmarshal(fileContent, &result)
	if err != nil {
		return "", err
	}

	if result["kind"] == nil {
		return "", errors.New("kind field not found")
	}

	return result["kind"].(string), nil
}

// 需要给定一个OBJ对象，然后根据这个玩野
func ParseAPIObjectFromYamlfileContent(fileContent []byte, obj interface{}) error {
	err := yaml.Unmarshal(fileContent, obj)
	if err != nil {
		k8log.ErrorLog("Kubectl", "ParseAPIObjectFromYamlfileContent: Unmarshal object failed "+err.Error())
		return err
	}
	return err
}

// 用来解决发送API对象到服务器的问题
func PostAPIObjectToServer(URL string, obj interface{}) error {
	k8log.DebugLog("PostAPIObjectToServer", "URL: "+URL)
	// 发送到服务器
	code, res, err := netrequest.PostRequestByTarget(URL, obj)
	if err != nil {
		k8log.ErrorLog("Kubectl", "ParseAPIObjectFromYamlfileContent: Unmarshal object failed "+err.Error())
		return err
	}

	fmt.Println("code: ", code)
	bodyBytes, err := json.Marshal(res)
	if err != nil {
		return err
	}
	fmt.Println(string(bodyBytes))

	return nil
}
