package message

import (
	"fmt"
	"miniK8s/pkg/k8log"
	"time"

	"github.com/streadway/amqp"
)

type Publisher struct {
	conn    *amqp.Connection
	amqpURI string
	// 重连次数
	// 最大重连次数
	maxReconnect int
	// 重连间隔时间
	reconnectInterval int
}

func NewPublisher(conf *MsgConfig) (*Publisher, error) {
	url := "amqp://" + conf.User + ":" + conf.Password + "@" + conf.Host + ":" + fmt.Sprint(conf.Port) + "/" + conf.VHost
	k8log.DebugLog("message", "url is "+url)
	p := new(Publisher)
	connection, err := amqp.Dial(url)
	// 配置Publisher
	p.conn = connection
	p.amqpURI = url
	p.maxReconnect = conf.MaxReconnect
	p.reconnectInterval = conf.ReconnectInterval

	if err != nil {
		return nil, err
	}

	// 为了保证链接持续和自动重连的机制
	// 注册一个channel，如果链接断开了，会自动重连
	errorChannnel := make(chan *amqp.Error)

	// 开一个线程在那等着，如果链接断开了，就会自动往通道里面写入消息
	// 然后处理重新链接的逻辑
	go p.AutoReconnect(errorChannnel)

	// 注册通道，这样当链接断开的时候，会自动往通道里面写入消息
	p.conn.NotifyClose(errorChannnel)

	return p, nil
}

// 此外在发布消息的时候，也要检测对应的key绑定的队列是否存在，如果不存在，消息会丢失
// key：你要发给谁，就把key设置成谁的名字
// contentType：消息的类型，比如json、text等,参考下面的变量
// const ContentTypeJson = "application/json"
// const ContentTypeText = "text/plain"
// msg：消息的内容
func (p *Publisher) Publish(key string, contentType string, msg []byte) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 检测这个key是否绑定了队列，队列不存在就创建队列

	_, err = ch.QueueDeclare(key, true, false, false, false, nil)
	if err != nil {
		return err
	}
	// 如果没有绑定队列，就尝试绑定队列
	err = ch.QueueBind(key, key, queueToExchange[key], false, nil)
	if err != nil {
		return err
	}

	// 发布消息
	err = ch.Publish(queueToExchange[key], key, false, false, amqp.Publishing{
		ContentType: contentType,
		Body:        msg,
	})
	if err != nil {
		return err
	}
	return nil
}

// 回收的时候执行关闭函数
func (p *Publisher) Close() {
	k8log.DebugLog("message", "close message connection")
	p.conn.Close()
}

// 接受一个通道
func (p *Publisher) AutoReconnect(errCh <-chan *amqp.Error) {
	// 通道在读取的时候,如果没有消息就会阻塞在这里
	// 定义一个err
	err := <-errCh
	if err != nil {
		k8log.DebugLog("message", "connection closed, reconnecting...")
		// 尝试重连
		for i := 0; i < p.maxReconnect; i++ {
			// 重连
			connection, err := amqp.Dial(p.amqpURI)

			// 如果重连成功，就退出循环
			if err == nil {
				k8log.DebugLog("message", "reconnect success")

				// 重新设置链接
				p.conn = connection
				// 重新设置通道
				errCh = p.conn.NotifyClose(make(chan *amqp.Error))
				// 如果重连成功，就退出循环
				break
			} else {
				// 如果重连失败，就等待一段时间再重连
				k8log.DebugLog("message", "reconnect failed, retry after "+fmt.Sprint(p.reconnectInterval)+"s")
				// 如果重连失败，就等待一段时间再重连
				time.Sleep(time.Duration(p.reconnectInterval) * time.Second)
			}
		}
	}

}
