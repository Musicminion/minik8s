package message

import (
	"miniK8s/pkg/k8log"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestSubscribe(t *testing.T) {
	// 创建一个消息订阅者
	s, err := NewSubscriber(DefaultMsgConfig())
	if err != nil {
		t.Fatal(err)
	}
	// 创建一个chan
	ch := make(chan struct{})

	go s.Subscribe("apiServer", func(msg amqp.Delivery) {
		k8log.InfoLog("message", "Received a message: "+string(msg.Body))
		// t.Log(string(msg.ContentType))
		// t.Log(string(msg.Body))
	}, ch)

	time.Sleep(5 * time.Second)
	// 关闭chan
	close(ch)

	// // 订阅消息
	// err =
	// if err != nil {
	// 	// t.Fatal(err)
	// 	k8log.FatalLog("message", "Failed to subscribe message:"+err.Error())
	// }

}
