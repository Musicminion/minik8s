package message

import (
	"fmt"
	"miniK8s/pkg/k8log"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

type Subscriber struct {
	conn    *amqp.Connection
	amqpURI string
	// 重连次数
	// 最大重连次数
	maxReconnect int
	// 重连间隔时间
	reconnectInterval int
}

func NewSubscriber(conf *MsgConfig) (*Subscriber, error) {
	url := "amqp://" + conf.User + ":" + conf.Password + "@" + conf.Host + ":" + fmt.Sprint(conf.Port) + "/" + conf.VHost
	k8log.DebugLog("message", "url is "+url)
	s := new(Subscriber)
	connection, err := amqp.Dial(url)
	if err != nil {
		k8log.ErrorLog("message", "dial error: "+err.Error())
		return nil, err
	}

	// 配置Publisher
	s.conn = connection
	s.amqpURI = url
	s.maxReconnect = conf.MaxReconnect
	s.reconnectInterval = conf.ReconnectInterval

	// 为了保证链接持续和自动重连的机制
	// 注册一个channel，如果链接断开了，会自动重连
	errorChannnel := make(chan *amqp.Error)
	go s.AutoReconnect(errorChannnel)
	s.conn.NotifyClose(errorChannnel)

	return s, nil
}

// 接受一个通道
func (s *Subscriber) AutoReconnect(errCh <-chan *amqp.Error) {
	// 通道在读取的时候,如果没有消息就会阻塞在这里
	// 定义一个err
	err := <-errCh
	if err != nil {
		k8log.DebugLog("message", "connection closed, reconnecting...")
		// 尝试重连
		for i := 0; i < s.maxReconnect; i++ {
			// 重连
			connection, err := amqp.Dial(s.amqpURI)

			// 如果重连成功，就退出循环
			if err == nil {
				k8log.DebugLog("message", "reconnect success")

				// 重新设置链接
				s.conn = connection
				// 重新设置通道
				errCh = s.conn.NotifyClose(make(chan *amqp.Error))
				// 如果重连成功，就退出循环
				break
			} else {
				// 如果重连失败，就等待一段时间再重连
				k8log.DebugLog("message", "reconnect failed, retry after "+fmt.Sprint(s.reconnectInterval)+"s")
				// 如果重连失败，就等待一段时间再重连
				time.Sleep(time.Duration(s.reconnectInterval) * time.Second)
			}
		}
	}
}

func (s *Subscriber) handleMsgs(msgs <-chan amqp.Delivery, handleFunc func(amqp.Delivery)) {
	k8log.DebugLog("message", "start to handle messages")
	for msg := range msgs {
		k8log.DebugLog("message", "received a message: "+string(msg.Body))
		handleFunc(msg)
	}
	k8log.DebugLog("message", "handleMsgs exit")
}

// 调用者主动关闭这个通道，否则程序阻塞在里面运行
// 传递处理函数，然后在子线程里面会一直监听新来的消息，一旦有消息就会调用处理函数
func (s *Subscriber) Subscribe(queueName string, handleFunc func(amqp.Delivery), stopChannelCh <-chan struct{}) error {
	ch, err := s.conn.Channel()

	if err != nil {
		return err
	}
	defer ch.Close()

	// 检查queueName是否存在，不存在就创建
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive，排他性队列，只有创建这个队列的链接才能使用这个队列
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return err
	}

	var isFanoutExchangeQueue bool = false
	// 如果是fanoutExchange的队列，就绑定到fanoutExchange上, 注意需要判断queueName是否包含fanoutExchange的前缀
	for _, feq := range FanoutExchangeQueues {
		if strings.Contains(queueName, feq) {
			err = ch.QueueBind(
				queueName,                  // queue name
				queueName,                  // routing key
				queueToExchange[feq], // exchange
				false,
				nil,
			)
			if err != nil {
				return err
			}
			isFanoutExchangeQueue = true
			continue
		}
	}
	
	// 如果不是fanoutExchange的队列，就绑定到directExchange上
	if !isFanoutExchangeQueue {
		err = ch.QueueBind(
			queueName,                  // queue name
			queueName,                  // routing key
			queueToExchange[queueName], // exchange
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	if err != nil {
		return err
	}

	stop := func(stopChannelCh <-chan struct{}) {
		<-stopChannelCh
		k8log.DebugLog("message", "stopChannelHanlder exit")
	}

	// 启动一个goroutine来处理消息
	// go s.handleMsgs(msgs, handleFunc)
	go s.handleMsgs(msgs, handleFunc)

	// 把函数阻塞在这里，直到收到stopChannelCh的消息,才会退出
	stop(stopChannelCh)
	return nil
}

// 关闭链接并且关闭通道，这时候
func (s *Subscriber) UnSubscribe(stopChannelCh chan<- struct{}) {
	close(stopChannelCh)
	s.conn.Close()
}
