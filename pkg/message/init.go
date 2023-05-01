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
// 支持的有apiServer、scheduler、controller，这三个都是控制平面的组件

func init() {
	conf := DefaultMsgConfig()
	url := "amqp://" + conf.User + ":" + conf.Password + "@" + conf.Host + ":" + fmt.Sprint(conf.Port) + "/" + conf.VHost
	connection, err := amqp.Dial(url)
	if err != nil {
		k8log.FatalLog("message", "Failed to connect to RabbitMQ:"+err.Error())
	}
	defer connection.Close()

	// 检查交换机K8sExchange是否存在
	ch, err := connection.Channel()
	if err != nil {
		k8log.FatalLog("message", "Failed to open a channel:"+err.Error())
	}
	defer ch.Close()
	// 声明一个交换机，存在的话会检查类型，不存在new一个，
	// 这里考虑用直接交换机，根据key来路由消息
	// 假如你要发送给kublet消息，那么key就是kublet
	err = ch.ExchangeDeclare("K8sExchange", "direct", true, false, false, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to declare an exchange:"+err.Error())
	}

	// 声明相关的队列
	// 1. scheduler队列
	_, err = ch.QueueDeclare("scheduler", true, false, false, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to declare scheduler queue:"+err.Error())
	}

	// 2. controller队列
	_, err = ch.QueueDeclare("controller", true, false, false, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to declare controller queue:"+err.Error())
	}

	// 3. apiServer队列
	_, err = ch.QueueDeclare("apiServer", true, false, false, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to declare apiServer queue:"+err.Error())
	}

	// 4. serviceUpdate队列
	_, err = ch.QueueDeclare("serviceUpdate", true, false, false, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to declare serviceUpdate queue:"+err.Error())
	}

	// 5. endpointUpdate队列
	_, err = ch.QueueDeclare("endpointUpdate", true, false, false, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to declare endpointUpdate queue:"+err.Error())
	}


	// 绑定队列和交换机
	// 绑定scheduler队列
	err = ch.QueueBind("scheduler", "scheduler", K8sExchange, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to bind scheduler queue:"+err.Error())
	}
	// 绑定controller队列
	err = ch.QueueBind("controller", "controller", K8sExchange, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to bind controller queue:"+err.Error())
	}
	// 绑定apiServer队列
	err = ch.QueueBind("apiServer", "apiServer", K8sExchange, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to bind apiServer queue:"+err.Error())
	}
	// 绑定serviceUpdate队列
	err = ch.QueueBind("serviceUpdate", "serviceUpdate", K8sExchange, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to bind serviceUpdate queue:"+err.Error())
	}
	// 绑定endpointUpdate队列
	err = ch.QueueBind("endpointUpdate", "endpointUpdate", K8sExchange, false, nil)
	if err != nil {
		k8log.FatalLog("message", "Failed to bind endpointUpdate queue:"+err.Error())
	}

	k8log.DebugLog("message", "init binding message finished")
}
