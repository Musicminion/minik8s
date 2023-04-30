package handlers

import (
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/message"

	"github.com/streadway/amqp"
)

// 负责处理调度结果的消息, content是表示调度到节点的名称
func ScheduleResultHandler(content string) {
	k8log.WarnLog("scheduleResultHandler", "收到调度结果消息!!!!!!!!!!!!!"+content)
	// 在etcd里面搜索这个节点的信息，检查是否存在，存在就把

}

// 主要是一个消息分发的功能，分发个消息的处理函数
func MessageHandler(msg amqp.Delivery) {
	// 把msg.body转换为Json格式的Message
	result, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("messageHandler", "消息格式错误,无法转换为Message")
	}

	switch result.Type {
	// 检测到调度完成的消息
	case message.ScheduleResult:
		ScheduleResultHandler(result.Content)

	default:
		k8log.WarnLog("messageHandler", "消息类型不在处理范围中,无法处理")
	}

}
