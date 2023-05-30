package message

import (
	"miniK8s/pkg/k8log"
	"testing"
)

func TestPublish(t *testing.T) {
	// 创建一个消息发布者
	p, err := NewPublisher(DefaultMsgConfig())
	if err != nil {
		t.Fatal(err)
	}
	// 发布消息
	err = p.Publish("apiServer", "text/plain", []byte("hello"))
	k8log.InfoLog("message", "Publish a message: hello")
	if err != nil {
		t.Fatal(err)
	}
}
