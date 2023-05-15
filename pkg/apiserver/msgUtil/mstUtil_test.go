package msgutil

import (
	"encoding/json"
	"fmt"
	"testing"

	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/message"
	"miniK8s/util/file"
	"miniK8s/util/stringutil"

	"gopkg.in/yaml.v3"
)

func TestPublishUpdateService(t *testing.T) {
	fileContent, err := file.ReadFile("./testFile/Service.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var serviceStore apiObject.ServiceStore
	err = yaml.Unmarshal(fileContent, &serviceStore)
	if err != nil {
		k8log.ErrorLog("Kubectl", "ParseAPIObjectFromYamlfileContent: Unmarshal object failed "+err.Error())
		t.Error("Unmarshal service object failed")
	}

	fmt.Println("service Info:", serviceStore)

	serviceUpdate := &entity.ServiceUpdate{
		Action: entity.CREATE,
		ServiceTarget: serviceStore,
	}
	PublishUpdateService(serviceUpdate)
}

func TestPublishRequestNodeScheduleMsg(t *testing.T) {
	fileContent, err := file.ReadFile("./testFile/Service.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var pod apiObject.Pod
	err = yaml.Unmarshal(fileContent, &pod)
	if err != nil {
		t.Errorf("unmarshal pod failed")
	}
	resourceURI := stringutil.Replace(config.PodSpecURL, config.URI_PARAM_NAME_PART, pod.GetPodName())
	resourceURI = stringutil.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, pod.GetPodNamespace())
	message := message.Message{
		Type:         message.RequestSchedule,
		Content:      pod.GetPodName(),
		ResourceURI:  resourceURI,
		ResourceName: pod.GetPodName(),
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		t.Errorf("Error marshal message")
	}

	PublishMsg("scheduler", jsonMsg)
}
