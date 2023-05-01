package msgutil

import(
	"testing"
	"fmt"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/apiObject"
	"miniK8s/util/file"
	"gopkg.in/yaml.v3"
	"miniK8s/pkg/k8log"

)


func TestPublishUpdateService(t *testing.T) {
	fileContent, err := file.ReadFile("./testFile/Service.yaml")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var service apiObject.Service
	err = yaml.Unmarshal(fileContent, &service)
	if err != nil {
		k8log.ErrorLog("Kubectl", "ParseAPIObjectFromYamlfileContent: Unmarshal object failed "+err.Error())
		t.Error("Unmarshal service object failed")
	}

	fmt.Println("service Info:", service)

	serviceUpdate := &entity.ServiceUpdate{
		Action: entity.CREATE,
		ServiceTarget: entity.ServiceWithEndpoints{
			Service: service,
			Endpoints: make([]apiObject.Endpoint, 0),
		},
	}
	PublishUpdateService(serviceUpdate)
}