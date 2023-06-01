package proxy

import (
	"encoding/json"
	"miniK8s/pkg/apiObject"
	"miniK8s/pkg/config"
	"miniK8s/pkg/entity"
	"miniK8s/pkg/k8log"
	"miniK8s/pkg/listwatcher"
	"miniK8s/pkg/message"
	"miniK8s/util/file"
	"miniK8s/util/host"
	netrequest "miniK8s/util/netRequest"
	"miniK8s/util/nginx"
	"miniK8s/util/stringutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"
)

type KubeProxy struct {
	lw             *listwatcher.Listwatcher
	stopChannel    <-chan struct{}
	serviceUpdates chan *entity.ServiceUpdate
	dnsUpdates     chan *entity.DnsUpdate
	iptableManager IptableManager
	dnsManager     DnsManager
	hostList       []string
	nginxPod       *apiObject.Pod
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
		nginxPod:       &apiObject.Pod{},
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

// 通知nginx Pod 更新Config
func (proxy *KubeProxy) updateNginxConfig() {
	podUpdate := &entity.PodUpdate{
		Action: message.EXEC,
		PodTarget: apiObject.PodStore{
			Spec: apiObject.PodSpec{
				NodeName: host.GetHostName(),
			},
			Basic: apiObject.Basic{
				Metadata: apiObject.Metadata{
					UUID: proxy.nginxPod.Metadata.UUID,
				},
			},
		},
		Cmd: []string{"sh", "-c", "nginx -s reload"},
	}
	k8log.DebugLog("Kubeproxy", "updateNginxConfig: send message to reload nginx config")
	// 向本机的kubelet发送消息，通知nginx pod更新配置
	message.PublishUpdatePod(podUpdate)
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

	// 用 append|trunc 模式打开主机文件
	f, err := os.OpenFile(config.HostsConfigFilePath, os.O_APPEND|os.O_WRONLY|os.O_TRUNC, os.ModeAppend)
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to open hosts file")
		return
	}
	defer f.Close()

	// 清空/etc/hosts， 并写入 "127.0.0.1 localhost"
	_, err = f.WriteString("127.0.0.1 localhost\n")
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to write to hosts file")
		return
	}

	// 修改host文件
	proxy.hostList = hostUpdate.HostList
	for _, host := range proxy.hostList {
		_, err = f.WriteString(host + "\n")
		if err != nil {
			k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to write to hosts file")
			return
		}
	}

	// 以下内容更新nginx的配置文件
	switch hostUpdate.Action {
	case message.CREATE:
		k8log.InfoLog("Kubeproxy", "HandleHostUpdate: create Host action")
		err = nginx.WriteConf(*hostUpdate.DnsTarget.ToDns(), hostUpdate.DnsConfig)
		if err != nil {
			k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to write nginx conf")
			return
		}
	case message.UPDATE:
		err = nginx.DeleteConf(*hostUpdate.DnsTarget.ToDns())
		if err != nil {
			k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to delete nginx conf")
			return
		}
		err = nginx.WriteConf(*hostUpdate.DnsTarget.ToDns(), hostUpdate.DnsConfig)
		if err != nil {
			k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to write nginx conf")
			return
		}
	case message.DELETE:
		k8log.DebugLog("Kubeproxy", "HandleHostUpdate: delete Host action")
		err = nginx.DeleteConf(*hostUpdate.DnsTarget.ToDns())
		if err != nil {
			k8log.ErrorLog("Kubeproxy", "HandleHostUpdate: failed to delete nginx conf")
			return
		}
	}

	time.Sleep(3 * time.Second)

	// 更新nginx的配置文件后，reload nginx
	proxy.updateNginxConfig()
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

func (proxy *KubeProxy) CreateNginxPod() {
	k8log.DebugLog("Dns-Controller", "Run: start to create nginx pod")

	// 检查etcd中是否存在nginx pod
	GetNodePodURL := stringutil.Replace(config.NodeAllPodsURL, config.URL_PARAM_NAME_PART, host.GetHostName())
	GetNodePodURL = config.GetAPIServerURLPrefix() + GetNodePodURL
	k8log.DebugLog("Dns-Controller", "Run: get pods from etcd, URL is "+GetNodePodURL)
	podsOnNode := []apiObject.PodStore{}

	code, err := netrequest.GetRequestByTarget(GetNodePodURL, &podsOnNode, "data")
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to get pods from etcd"+err.Error())
		return
	}
	if code != http.StatusOK {
		k8log.ErrorLog("Dns-Controller", "Run: failed to get pods from etcd, code is not 200")
		return
	}

	//  若存在，说明kubeproxy不是第一次启动
	for _, pod := range podsOnNode {
		if strings.Contains(pod.GetPodName(), "dns-nginx") {
			k8log.DebugLog("Dns-Controller", "Run: nginx pod already exists")
			proxy.nginxPod = pod.ToPod()
			return
		}
	}

	// 不存在，根据pod yaml 创建nginx pod
	path := NginxPodYamlPath
	fileContent, err := file.ReadFile(path)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to read file"+err.Error())
		return
	}

	// 将文件内容转换为Pod对象
	nginxPod := &apiObject.Pod{}
	err = yaml.Unmarshal(fileContent, nginxPod)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to unmarshal"+err.Error())
		return
	}

	// 判断namespace是否为空
	if nginxPod.GetObjectNamespace() == "" {
		nginxPod.Metadata.Namespace = config.DefaultNamespace
	}

	CreateNginxPodURL := stringutil.Replace(config.PodsURL, config.URL_PARAM_NAMESPACE_PART, nginxPod.GetObjectNamespace())
	CreateNginxPodURL = config.GetAPIServerURLPrefix() + CreateNginxPodURL

	nginxPod.Metadata.Name += "-" + stringutil.GenerateRandomStr(5)
	nginxPod.Spec.NodeName = host.GetHostName()

	code, _, err = netrequest.PostRequestByTarget(CreateNginxPodURL, nginxPod)
	if err != nil {
		k8log.ErrorLog("Dns-Controller", "Run: failed to post request"+err.Error())
		return
	}
	if code != http.StatusCreated {
		k8log.ErrorLog("Dns-Controller", "Run: failed to create pod")
		return
	}

	k8log.InfoLog("Dns-Controller", "HandleServiceUpdate: success to create nginx pod")

	// 读出nginx pod，获取其UUID
	k8log.DebugLog("Kubeproxy", "updateNginxConfig: get all pods from apiserver")
	code, err = netrequest.GetRequestByTarget(GetNodePodURL, &podsOnNode, "data")
	if err != nil {
		k8log.ErrorLog("Kubeproxy", "updateNginxConfig: failed to get all pods from apiserver")
	}
	if code != http.StatusOK {
		k8log.ErrorLog("Kubeproxy", "updateNginxConfig: failed to get all pods from apiserver")
	}

	// 筛选出nginx pod, 记录其UUID
	for _, pod := range podsOnNode {
		if strings.EqualFold(pod.Metadata.Name, nginxPod.GetObjectName()) {
			// 更新proxy的nginxPod
			proxy.nginxPod = pod.ToPod()
			break
		}
	}

}

func (proxy *KubeProxy) Run() {
	// 创建nginx的pod
	proxy.CreateNginxPod()
	// serviceUpdate
	go proxy.lw.WatchQueue_Block(message.ServiceUpdateWithNode(host.GetHostName()), proxy.HandleServiceUpdate, make(chan struct{}))

	// hostUpdate, 来自dnsController
	go proxy.lw.WatchQueue_Block(message.HostUpdateWithNode(host.GetHostName()), proxy.HandleHostUpdate, make(chan struct{}))

	// 持续监听serviceUpdates和dnsUpdates的channel
	for proxy.syncLoopIteration(proxy.serviceUpdates, proxy.dnsUpdates) {
	}

}

func (proxy *KubeProxy) Stop() {
	// TODO: stop
}
