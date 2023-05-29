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

	// DirectExchange
	for _, fe := range DirectExchange {
		err = ch.ExchangeDeclare(fe, "direct", true, false, false, false, nil)
		if err != nil {
			k8log.FatalLog("message", "Failed to declare an exchange:"+err.Error())
		}
	}
	
	// FanoutExchange
	for _, fe := range FanoutExchange {
		err = ch.ExchangeDeclare(fe, "fanout", true, false, false, false, nil)
		if err != nil {
			k8log.FatalLog("message", "Failed to declare an exchange:"+err.Error())
		}
	}
}

// 绑定node的相关queue到所有的fanout队列
func BindFinoutQueue(nodename string) {
	conf := DefaultMsgConfig()
	url := "amqp://" + conf.User + ":" + conf.Password + "@" + conf.Host + ":" + fmt.Sprint(conf.Port)
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

	for _, feq := range FanoutExchangeQueues {
		// 声明一个队列，存在的话会检查属性，不存在new一个
		_, err = ch.QueueDeclare(feq+"-"+nodename, true, false, false, false, nil)
		if err != nil {
			k8log.FatalLog("message", "Failed to declare a queue:"+err.Error())
		}

		// 绑定队列到fanout交换机
		err = ch.QueueBind(feq+"-"+nodename, feq, queueToExchange[feq], false, nil)
		if err != nil {
			k8log.FatalLog("message", "Failed to bind a queue:"+err.Error())
		}
	}
}
