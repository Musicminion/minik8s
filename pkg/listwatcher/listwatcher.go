package listwatcher

import (
	"miniK8s/pkg/message"

	"github.com/streadway/amqp"
)

type Listwatcher struct {
	subscriber *message.Subscriber
}

func NewListWatcher(conf *ListwatcherConfig) (*Listwatcher, error) {
	// message.NewSubscriber(message.DefaultMsgConfig())

	newSubscriber, err := message.NewSubscriber(conf.subscriberConfig)
	if err != nil {
		return nil, err
	}

	ls := &Listwatcher{
		subscriber: newSubscriber,
	}

	return ls, nil
}

// WatchQueue_Block 阻塞的方式监听队列，一旦有消息就会调用handleFunc
// 那么,只有当调用者主动关闭done的时候，才会退出
func (ls *Listwatcher) WatchQueue_Block(queueName string, handleFunc func(amqp.Delivery), done chan struct{}) error {
	ls.subscriber.Subscribe(queueName, handleFunc, done)
	return nil
}

// WatchQueue_NoBlock 非阻塞的方式监听队列，一旦有消息就会调用handleFunc
// 函数立刻返回，返回一个函数，调用这个函数可以取消监听
func (ls *Listwatcher) WatchQueue_NoBlock(queueName string, handleFunc func(amqp.Delivery)) (func(), error) {
	stopChannelChan := make(chan struct{})

	go ls.subscriber.Subscribe(queueName, handleFunc, stopChannelChan)

	cancelFunc := func() {
		ls.subscriber.UnSubscribe(stopChannelChan)
	}

	return cancelFunc, nil
}

