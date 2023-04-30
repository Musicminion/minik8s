package msgutil

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
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
	resourceURI := strings.Replace(config.PodSpecURL, ":name", pod.GetPodName(), -1)

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
