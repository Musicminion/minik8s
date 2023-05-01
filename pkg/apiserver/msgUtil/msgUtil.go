package msgutil

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/config"
	"miniK8s/pkg/message"
	"strings"
)

type MsgUtil struct {
	Publisher *message.Publisher
}

var ServerMsgUtil *MsgUtil

func init() {
	// 初始化消息队列
	newPublisher, err := message.NewPublisher(message.DefaultMsgConfig())
	if err != nil {
		panic(err)
	}
	ServerMsgUtil = &MsgUtil{
		Publisher: newPublisher,
	}
}

// 发布消息的组件
func PublishMsg(queueName string, msg []byte) error {
	return ServerMsgUtil.Publisher.Publish(queueName, message.ContentTypeJson, msg)
}

// 发布消息的组件函数
func PublishRequestNodeScheduleMsg(pod *apiObject.PodStore) error {
	resourceURI := strings.Replace(config.PodSpecURL, config.URI_PARAM_NAME_PART, pod.GetPodName(), -1)
	resourceURI = strings.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, pod.GetPodNamespace(), -1)
	message := message.Message{
		Type:         message.RequestSchedule,
		Content:      pod.GetPodName(),
		ResourceURI:  resourceURI,
		ResourceName: pod.GetPodName(),
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return err
	}

	return PublishMsg("scheduler", jsonMsg)
}

// func PublishUpdateService(service *apiObject.ServiceStore) error {
// 	resourceURI := strings.Replace(config.PodSpecURL, config.URI_PARAM_NAME_PART, service.GetName(), -1)
// 	resourceURI = strings.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, service.GetNamespace(), -1)
// 	message := message.Message{
// 		Type:         message.PUT,
// 		Content:      service.GetName(),
// 		ResourceURI:  resourceURI,
// 		ResourceName: service.GetName(),
// 	}

// 	jsonMsg, err := json.Marshal(message)

// 	if err != nil {
// 		return err
// 	}

// 	return PublishMsg("apiServer", jsonMsg)
// }

func PublishUpdateService(serviceUpdate *entity.ServiceUpdate) error {
	resourceURI := strings.Replace(config.PodSpecURL, config.URI_PARAM_NAME_PART, serviceUpdate.ServiceTarget.Service.GetName(), -1)
	resourceURI = strings.Replace(resourceURI, config.URL_PARAM_NAMESPACE_PART, serviceUpdate.ServiceTarget.Service.GetNamespace(), -1)
	message := message.Message{
		Type:         message.PUT,
		Content:      serviceUpdate.ServiceTarget.Service.GetName(),
		ResourceURI:  resourceURI,
		ResourceName: serviceUpdate.ServiceTarget.Service.GetName(),
	}

	jsonMsg, err := json.Marshal(message)

	if err != nil {
		return err
	}

	return PublishMsg("serviceUpdate", jsonMsg)
}
 