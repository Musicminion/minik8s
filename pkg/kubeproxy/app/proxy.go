package proxy

import (
	"encoding/json"
	msgutil "miniK8s/pkg/apiserver/msgUtil"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	"miniK8s/util/nginx"
	"os"

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
	hostList       []string
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
		hostList:       make([]string, 0),
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

// 监听到HostUpdate消息后, 并修改本机的host文件
func (proxy *KubeProxy) HandleHostUpdate(msg amqp.Delivery) {
	k8log.DebugLog("Kubeproxy", "HandleHostUpdate: receive host update message")
	parsedMsg, err := message.ParseJsonMessageFromBytes(msg.Body)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "消息格式错误,无法转换为Message")
		return
	}

	hostUpdate := &entity.HostUpdate{}

	err = json.Unmarshal([]byte(parsedMsg.Content), &hostUpdate)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "HandleDnsUpdate: failed to unmarshal")
		return
	}

	// 查看hostUpdate内容
	k8log.DebugLog("Kubeproxy", "HandleHostUpdate: hostUpdate: "+ parsedMsg.Content)

	// 一下内容更新本机的host文件
	// Open hosts file with append mode, clear first
	f, err := os.OpenFile(config.HostsConfigFilePath, os.O_APPEND|os.O_WRONLY|os.O_TRUNC, os.ModeAppend)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to open hosts file")
		return
	}
	defer f.Close()

	// Write 127.0.0.1 localhost to hosts file
	_, err = f.WriteString("127.0.0.1 localhost\n")
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to write to hosts file")
		return
	}

	// Write each host to hosts file
	proxy.hostList = hostUpdate.HostList
	for _, host := range proxy.hostList {
		_, err = f.WriteString(host + "\n")
		if err != nil {
			k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to write to hosts file")
			return
		}
	}

	// 以下内容更新nginx的配置文件
	nginx.WriteConf(*hostUpdate.DnsTarget.ToDns(), hostUpdate.DnsConfig)
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
	go proxy.lw.WatchQueue_Block(msgutil.ServiceUpdateTopic, proxy.HandleServiceUpdate, make(chan struct{}))

	// endpointUpdate
	go proxy.lw.WatchQueue_Block(msgutil.HostUpdateTopic, proxy.HandleHostUpdate, make(chan struct{}))
	// 持续监听serviceUpdates和dnsUpdates的channel
	for proxy.syncLoopIteration(proxy.serviceUpdates, proxy.dnsUpdates) {
	}

}

func (proxy *KubeProxy) Stop() {
	// TODO: stop
}
