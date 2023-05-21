package proxy

import (
	"encoding/json"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
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
	lw             *listwatcher.Listwatcher
	stopChannel    <-chan struct{}
	serviceUpdates chan *entity.ServiceUpdate
	dnsUpdates     chan *entity.DnsUpdate
	iptableManager IptableManager
	dnsManager     DnsManager
}

func NewKubeProxy(lsConfig *listwatcher.ListwatcherConfig) *KubeProxy {
	lw, err := listwatcher.NewListWatcher(lsConfig)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "NewKubeProxy: new watcher failed")
	}
	// TODO: health check server
	iptableManager := NewIptableManager()
	dnsManager := NewDnsManager()
	proxy := &KubeProxy{
		lw:             lw,
		iptableManager: iptableManager,
		dnsManager:     dnsManager,
		stopChannel:    make(<-chan struct{}),
		serviceUpdates: make(chan *entity.ServiceUpdate, 10),
		dnsUpdates:     make(chan *entity.DnsUpdate, 10),
	}
	return proxy
}

// 监听到serviceUpdate消息后，解析并写入管道
func (proxy *KubeProxy) HandleServiceUpdate(msg amqp.Delivery) {
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "消息格式错误,无法转换为Message")
	}

	serviceUpdate := &entity.ServiceUpdate{}
	err = json.Unmarshal([]byte(parsedMsg.Content), serviceUpdate)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "HandleServiceUpdate: failed to unmarshal")
		return
	}
	proxy.serviceUpdates <- serviceUpdate

}

// 监听到DnsUpdate消息后，解析并写入管道
func (proxy *KubeProxy) HandleDnsUpdate(msg amqp.Delivery) {
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "消息格式错误,无法转换为Message")
	}
	dnsUpdate := &entity.DnsUpdate{}
	err = json.Unmarshal([]byte(parsedMsg.Content), dnsUpdate)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "HandleDnsUpdate: failed to unmarshal")
		return
	}
	proxy.dnsUpdates <- dnsUpdate

}

// 当管道发生变化时的处理函数
func (proxy *KubeProxy) syncLoopIteration(serviceUpdates <-chan *entity.ServiceUpdate, dnsUpdates <-chan *entity.DnsUpdate) bool {
	k8log.InfoLog("Kubeproxy", "syncLoopIteration: Sync loop Iteration")

	select {
	case serviceUpdate := <-serviceUpdates:
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
	case dnsUpdate := <-dnsUpdates:
		switch dnsUpdate.Action {
		case message.CREATE:
		case message.UPDATE:
		case message.DELETE:
		}
	}
	return true
}

func (proxy *KubeProxy) Run() {
	// serviceUpdate
	go proxy.lw.WatchQueue_Block(msgutil.ServiceUpdate, proxy.HandleServiceUpdate, make(chan struct{}))

	// endpointUpdate
	go proxy.lw.WatchQueue_Block(msgutil.EndpointUpdate, proxy.HandleDnsUpdate, make(chan struct{}))
	// 持续监听serviceUpdates和dnsUpdates的channel
	for proxy.syncLoopIteration(proxy.serviceUpdates, proxy.dnsUpdates) {
	}

}

func (proxy *KubeProxy) Stop() {
	// TODO: stop
}
