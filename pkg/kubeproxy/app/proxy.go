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
	iptableManager  *IptableManager
}

func NewKubeProxy(lsConfig *listwatcher.ListwatcherConfig) *KubeProxy {
	lw, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "NewKubeProxy: new watcher failed")
	}
	// TODO: health check server
	iptableManager := New()
	proxy := &KubeProxy{
		lw:              lw,
		iptableManager:  iptableManager,
		stopChannel:     make(<-chan struct{}),
		serviceUpdates:  make(chan *entity.ServiceUpdate, 10),
		endpointUpdates: make(chan *entity.EndpointUpdate, 10),
	}
	return proxy
}

// 监听到serviceUpdate消息后，解析并写入管道
func (proxy *KubeProxy) HandleServiceUpdate(msg amqp.Delivery) {
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "消息格式错误,无法转换为Message")
	}
	if parsedMsg.Type == message.CREATE {
		serviceUpdate := &entity.ServiceUpdate{}
		err := json.Unmarshal([]byte(parsedMsg.Content), serviceUpdate)
		if err != nil {
			k8log.ErrorLog("Kubeproxy", "HandleServiceUpdate: failed to unmarshal")
			return
		}
		proxy.serviceUpdates <- serviceUpdate
	}
}

// 监听到endpointUpdate消息后，解析并写入管道
func (proxy *KubeProxy) HandleEndpointUpdate(msg amqp.Delivery) {
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "消息格式错误,无法转换为Message")
	}
	if parsedMsg.Type == message.CREATE {
		endpointUpdate := &entity.EndpointUpdate{}
		err := json.Unmarshal([]byte(parsedMsg.Content), endpointUpdate)
		if err != nil {
			k8log.ErrorLog("Kubeproxy", "HandleServiceUpdate: failed to unmarshal")
			return
		}
		proxy.endpointUpdates <- endpointUpdate
	}
}

// 当管道发生变化时的处理函数
func (proxy *KubeProxy) syncLoopIteration(serviceUpdates <-chan *entity.ServiceUpdate, endpointUpdates <-chan *entity.EndpointUpdate) bool {
	k8log.InfoLog("Kubeproxy", "syncLoopIteration: Sync loop Iteration")

	// select {

	serviceUpdate, ok := <-serviceUpdates
	if !ok {
		k8log.InfoLog("Kubeproxy", "syncLoopIteration: serviceUpdates channel closed")
	}
	switch serviceUpdate.Action {
	case message.CREATE:
		k8log.InfoLog("Kubeproxy", "syncLoopIteration: create Service action")
		proxy.iptableManager.CreateService(serviceUpdate)

	case message.UPDATE:
		k8log.InfoLog("Kubeproxy", "syncLoopIteration: update Service action")
		proxy.iptableManager.UpdateService(serviceUpdate)

	case message.DELETE:
		k8log.InfoLog("Kubeproxy", "syncLoopIteration: delete Service action")
		proxy.iptableManager.DeleteService(serviceUpdate)
	}
	// case endpointUpdate := <-endpointUpdates:
	// 	switch endpointUpdate.Action {
	// 	case message.CREATE:
	// 		k8log.InfoLog("Kubeproxy", "syncLoopIteration: create Endpoint action")
	// 		proxy.iptableManager.CreateEndpoint(endpointUpdate)
	// 	case message.UPDATE:
	// 		k8log.InfoLog("Kubeproxy", "syncLoopIteration: update Endpoint action")
	// 		proxy.iptableManager.UpdateEndpoint(endpointUpdate)
	// 	case message.DELETE:
	// 		k8log.InfoLog("Kubeproxy", "syncLoopIteration: delete Endpoint action")
	// 		proxy.iptableManager.DeleteEndpoint(endpointUpdate)
	// 	}
	// }
	return true
}

func (proxy *KubeProxy) Run() {
	go proxy.lw.WatchQueue_Block("serviceUpdate", proxy.HandleServiceUpdate, make(chan struct{}))
	go proxy.lw.WatchQueue_Block("endpointUpdate", proxy.HandleEndpointUpdate, make(chan struct{}))
	// 持续监听serviceUpdates和endpointUpdates的channel
	for proxy.syncLoopIteration(proxy.serviceUpdates, proxy.endpointUpdates) {
	}

}

func (proxy *KubeProxy) Stop() {
	// TODO: stop
}
