package proxy

import (
	"encoding/json"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"

	"github.com/streadway/amqp"
)

var (
// serviceCIDR = flag.String("service-cidr", "10.244.0.0/16", "Service CIDR")
)

type KubeProxy struct {
	lw              *listwatcher.Listwatcher
	stopChannel     <-chan struct{}
	serviceUpdates  chan *entity.ServiceUpdate
	endpointUpdates chan *entity.EndpointUpdate
}

func NewKubeProxy(lsConfig *listwatcher.ListwatcherConfig) *KubeProxy {
    lw, err := listwatcher.NewListWatcher(lsConfig)
    if err != nil {
        k8log.ErrorLog("KubeProxy", "NewKubeProxy: new watcher failed")
    }
	proxy := &KubeProxy{
        lw: lw,
        stopChannel: make(<-chan struct{}),
        serviceUpdates: make(chan *entity.ServiceUpdate, 10),
        endpointUpdates: make(chan *entity.EndpointUpdate, 10),
    }
	return proxy
}

// 监听到serviceUpdate消息后，解析并写入管道
func (proxy *KubeProxy) HandleServiceUpdate(msg amqp.Delivery) {
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("proxy", "消息格式错误,无法转换为Message")
	}
	if parsedMsg.Type == message.PUT {
		serviceUpdate := &entity.ServiceUpdate{}
		err := json.Unmarshal([]byte(parsedMsg.Content), serviceUpdate)
		if err != nil {
			k8log.ErrorLog("proxy", "HandleServiceUpdate: failed to unmarshal")
			return
		}
		proxy.serviceUpdates <- serviceUpdate
	}
}

// 监听到endpointUpdate消息后，解析并写入管道
func (proxy *KubeProxy) HandleEndpointUpdate(msg amqp.Delivery) {
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("proxy", "消息格式错误,无法转换为Message")
	}
	if parsedMsg.Type == message.PUT {
		endpointUpdate := &entity.EndpointUpdate{}
		err := json.Unmarshal([]byte(parsedMsg.Content), endpointUpdate)
		if err != nil {
			k8log.ErrorLog("proxy", "HandleServiceUpdate: failed to unmarshal")
			return
		}
		proxy.endpointUpdates <- endpointUpdate
	}
}

// 当管道发生变化时的处理函数
func (proxy *KubeProxy) syncLoopIteration(serviceUpdates <-chan *entity.ServiceUpdate, endpointUpdates <-chan *entity.EndpointUpdate) bool {
	k8log.InfoLog("proxy", "syncLoopIteration: Sync loop Iteration")

	select {
	case serviceUpdate := <-serviceUpdates:

		switch serviceUpdate.Action {
		case entity.CREATE:
		case entity.UPDATE:
		case entity.DELETE:
		}
	case endpointUpdate := <-endpointUpdates:
		switch endpointUpdate.Action {
		case entity.CREATE:
		case entity.UPDATE:
		case entity.DELETE:
		}
	}
	return true
}


func (proxy *KubeProxy) Run() {
	go proxy.lw.WatchQueue_Block("ServiceUpdate", proxy.HandleServiceUpdate, make(chan struct{}))
	go proxy.lw.WatchQueue_Block("EndpointUpdate", proxy.HandleEndpointUpdate, make(chan struct{}))
    // 持续监听serviceUpdates和endpointUpdates的channel
    for proxy.syncLoopIteration(proxy.serviceUpdates, proxy.endpointUpdates) {
	}
}
