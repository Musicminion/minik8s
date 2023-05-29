package message

import (
	"fmt"
	"miniK8s/pkg/k8log"

	"github.com/streadway/amqp"
)

// 初始化的时候需要检查消息RabbitMQ服务器是否存在
// 如果不存在，那么就创建一个，然后还要根据key分配到不同的队列
// 目前我的设计是，根据key来路由消息，如果key不存在，那么就创建一个
// 如下图所示，交换机根据key来路由消息，如果key不存在，丢弃消息
//          /-- queue1 --\
// exchange --- queue2 --
//          \-- queue3 --/

// 所以你要发给谁，就把key设置成谁的名字

func init() {
	conf := DefaultMsgConfig()
	url := "amqp://" + conf.User + ":" + conf.Password + "@" + conf.Host + ":" + fmt.Sprint(conf.Port) + "/" + conf.VHost
	connection, err := amqp.Dial(url)
	if err != nil {
		k8log.FatalLog("message", "Failed to connect to RabbitMQ:"+err.Error())
	}
	defer connection.Close()

	// 打开一个channel
	ch, err := connection.Channel()
	if err != nil {
		k8log.FatalLog("message", "Failed to open a channel:"+err.Error())
	}
	defer ch.Close()

	// 声明一个交换机，存在的话会检查类型，不存在new一个，
	// 假如你要发送给kublet消息，那么key就是kubelet
	err = ch.ExchangeDeclare(DirectK8sExchange, "direct", true, false, false, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to declare an exchange:"+err.Error())
	}

	err = ch.ExchangeDeclare(FanoutK8sExchange, "fanout", true, false, false, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to declare an exchange:"+err.Error())
	}
}

// K8s消息交换机名字
const DirectK8sExchange = "DirectK8sExchange"
const FanoutK8sExchange = "FanoutK8sExchange"

// fanout模式的队列
const (
	HostUpdateQueue = HostUpdateTopic
)

var FanoutQueue = []string{
	HostUpdateQueue,
}
